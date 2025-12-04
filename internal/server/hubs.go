package server

import (
    "net/url"
    "time"
    "github.com/gorilla/websocket"
)

type bootstrapConn struct {
    uri        string
    ws         *websocket.Conn
    connected  bool
    lastAttempt int64
    attemptNum int
    reconnectTimer *time.Timer
}

type hubInfo struct {
    PeerId       string
    RegisteredAt int64
    LastActivity int64
    NetworkName  string
    Data         map[string]interface{}
}

func (s *Server) connectToBootstrapHubs() {
    for _, uri := range s.opts.BootstrapHubs {
        s.connectToHub(uri, 0)
    }
}

func (s *Server) connectToHub(uri string, attempt int) {
    u, err := url.Parse(uri)
    if err != nil {
        return
    }
    if u.Host == s.opts.Host && u.Port() == itoa(s.port) {
        return
    }
    ws, _, err := websocket.DefaultDialer.Dial(uri+"?peerId="+s.hubPeerId, nil)
    if err != nil {
        if attempt == 0 {
            return
        }
        return
    }
    info := &bootstrapConn{uri: uri, ws: ws, connected: true, lastAttempt: nowMs(), attemptNum: attempt}
    s.bootstrapMu.Lock()
    s.bootstrapConns[uri] = info
    s.bootstrapMu.Unlock()
    s.handleBootstrapOpen(info)
}

func (s *Server) handleBootstrapOpen(b *bootstrapConn) {
    s.emitBootstrapConnected(b.uri)
    s.sendAnnouncementToBootstrap(b.ws)
    go func() {
        for {
            _, data, err := b.ws.ReadMessage()
            if err != nil {
                break
            }
            s.handleBootstrapMessage(b.uri, data)
        }
        s.handleBootstrapClose(b)
    }()
}

func (s *Server) handleBootstrapClose(b *bootstrapConn) {
    s.bootstrapMu.Lock()
    b.connected = false
    s.bootstrapMu.Unlock()
    if s.running && b.attemptNum < s.opts.MaxReconnectAttempts {
        b.reconnectTimer = time.AfterFunc(time.Duration(s.opts.ReconnectIntervalMs)*time.Millisecond, func() {
            s.connectToHub(b.uri, b.attemptNum+1)
        })
    } else {
        s.bootstrapMu.Lock()
        delete(s.bootstrapConns, b.uri)
        s.bootstrapMu.Unlock()
    }
}

func (s *Server) disconnectBootstrap() {
    s.bootstrapMu.Lock()
    for _, b := range s.bootstrapConns {
        if b.reconnectTimer != nil {
            b.reconnectTimer.Stop()
        }
        if b.ws != nil {
            b.ws.Close()
        }
    }
    s.bootstrapConns = map[string]*bootstrapConn{}
    s.bootstrapMu.Unlock()
}

func (s *Server) sendAnnouncementToBootstrap(ws *websocket.Conn) {
    msg := map[string]interface{}{
        "type": "announce",
        "networkName": s.opts.HubMeshNamespace,
        "data": map[string]interface{}{
            "isHub": true,
            "port": s.port,
            "host": s.opts.Host,
            "capabilities": []string{"signaling", "relay"},
            "timestamp": nowMs(),
        },
    }
    ws.WriteJSON(msg)
    s.announceLocalPeersToBootstrap(ws)
}

func (s *Server) announceLocalPeersToBootstrap(ws *websocket.Conn) {
    s.networkMu.Lock()
    for netName, set := range s.networkPeers {
        for peerId := range set {
            pi := s.peerData[peerId]
            if pi == nil || !pi.Announced {
                continue
            }
            payload := map[string]interface{}{
                "type": "peer-discovered",
                "data": map[string]interface{}{
                    "peerId": peerId,
                    "isHub": netName == s.opts.HubMeshNamespace,
                },
                "networkName": netName,
                "fromPeerId": "system",
                "timestamp": nowMs(),
            }
            ws.WriteJSON(payload)
        }
    }
    s.networkMu.Unlock()
}

func (s *Server) handleBootstrapMessage(uri string, data []byte) {
    var msg inboundMessage
    if err := decodeJSON(data, &msg); err != nil {
        return
    }
    switch msg.Type {
    case "connected":
    case "peer-discovered":
        if m, ok := msg.Data.(map[string]interface{}); ok {
            id, _ := m["peerId"].(string)
            isHub := false
            if v, ok := m["isHub"].(bool); ok {
                isHub = v
            }
            netName := msg.NetworkName
            if netName == "" {
                netName = "global"
            }
            if isHub {
                s.emitHubDiscovered(id, uri)
                return
            }
            s.cacheCrossHubPeer(netName, id, m)
            s.forwardToLocalPeers(netName, outboundMessage{Type: "peer-discovered", Data: m, FromPeerId: "system", NetworkName: netName, Timestamp: nowMs()})
            
            // Forward to all OTHER bootstrap hubs (mesh mesh)
            s.announceToBootstrapExcept(id, netName, false, m, uri)
            
            // ALSO echo back to the originating hub so it knows the peer was received
            s.bootstrapMu.Lock()
            for origUri, b := range s.bootstrapConns {
                if origUri == uri && b.connected && b.ws != nil {
                    payload := map[string]interface{}{
                        "type": "peer-discovered",
                        "data": mergeMap(m, map[string]interface{}{
                            "peerId": id,
                            "isHub": false,
                        }),
                        "networkName": netName,
                        "fromPeerId": "system",
                        "timestamp": nowMs(),
                    }
                    b.ws.WriteJSON(payload)
                    break
                }
            }
            s.bootstrapMu.Unlock()
        }
    case "offer", "answer", "ice-candidate":
        if msg.TargetPeer != "" {
            s.forwardToLocalTarget(msg.TargetPeer, outboundMessage{Type: msg.Type, Data: msg.Data, FromPeerId: msg.FromPeerId, TargetPeer: msg.TargetPeer, NetworkName: msg.NetworkName, Timestamp: nowMs()})
        }
    }
}

func (s *Server) getHubStats() map[string]interface{} {
    s.bootstrapMu.Lock()
    bs := make([]map[string]interface{}, 0, len(s.bootstrapConns))
    for uri, info := range s.bootstrapConns {
        bs = append(bs, map[string]interface{}{"uri": uri, "connected": info.connected, "lastAttempt": info.lastAttempt, "attemptNumber": info.attemptNum})
    }
    s.bootstrapMu.Unlock()
    hubs := s.getConnectedHubs()
    return map[string]interface{}{"totalHubs": len(hubs), "connectedHubs": len(hubs), "hubs": hubs, "bootstrapHubs": bs}
}

func (s *Server) announceToBootstrap(peerId, netName string, isHub bool, data map[string]interface{}) {
    s.bootstrapMu.Lock()
    conns := make([]*websocket.Conn, 0, len(s.bootstrapConns))
    for _, b := range s.bootstrapConns {
        if b.connected && b.ws != nil {
            conns = append(conns, b.ws)
        }
    }
    s.bootstrapMu.Unlock()
    
    payload := map[string]interface{}{
        "type": "peer-discovered",
        "data": map[string]interface{}{
            "peerId": peerId,
            "isHub": isHub,
        },
        "networkName": netName,
        "fromPeerId": "system",
        "timestamp": nowMs(),
    }
    
    if data != nil {
        if m, ok := payload["data"].(map[string]interface{}); ok {
            for k, v := range data {
                m[k] = v
            }
        }
    }
    
    for _, ws := range conns {
        ws.WriteJSON(payload)
    }
}

func (s *Server) announceToBootstrapExcept(peerId, netName string, isHub bool, data map[string]interface{}, excludeUri string) {
    s.bootstrapMu.Lock()
    conns := make([]*websocket.Conn, 0, len(s.bootstrapConns))
    for uri, b := range s.bootstrapConns {
        if uri != excludeUri && b.connected && b.ws != nil {
            conns = append(conns, b.ws)
        }
    }
    s.bootstrapMu.Unlock()
    
    payload := map[string]interface{}{
        "type": "peer-discovered",
        "data": map[string]interface{}{
            "peerId": peerId,
            "isHub": isHub,
        },
        "networkName": netName,
        "fromPeerId": "system",
        "timestamp": nowMs(),
    }
    
    if data != nil {
        if m, ok := payload["data"].(map[string]interface{}); ok {
            for k, v := range data {
                m[k] = v
            }
        }
    }
    
    for _, ws := range conns {
        ws.WriteJSON(payload)
    }
}

func (s *Server) getConnectedHubs() []hubInfo {
    s.hubsMu.Lock()
    out := make([]hubInfo, 0, len(s.hubs))
    for _, h := range s.hubs {
        out = append(out, *h)
    }
    s.hubsMu.Unlock()
    return out
}

