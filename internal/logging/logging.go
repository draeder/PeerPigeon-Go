package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type LogLevel string

const (
	DEBUG LogLevel = "DEBUG"
	INFO  LogLevel = "INFO"
	WARN  LogLevel = "WARN"
	ERROR LogLevel = "ERROR"
)

type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     LogLevel               `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

var (
	minLevel = INFO
)

func SetLevel(level LogLevel) {
	minLevel = level
}

func shouldLog(level LogLevel) bool {
	levels := map[LogLevel]int{
		DEBUG: 0,
		INFO:  1,
		WARN:  2,
		ERROR: 3,
	}
	return levels[level] >= levels[minLevel]
}

func log(level LogLevel, message string, fields map[string]interface{}) {
	if !shouldLog(level) {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Level:     level,
		Message:   message,
		Fields:    fields,
	}

	if data, err := json.Marshal(entry); err == nil {
		fmt.Fprintf(os.Stderr, "%s\n", data)
	}
}

func Debug(message string, fields map[string]interface{}) {
	log(DEBUG, message, fields)
}

func Info(message string, fields map[string]interface{}) {
	log(INFO, message, fields)
}

func Warn(message string, fields map[string]interface{}) {
	log(WARN, message, fields)
}

func Error(message string, fields map[string]interface{}) {
	log(ERROR, message, fields)
}

// Convenience functions for common patterns
func PeerConnected(peerId string) {
	Info("peer_connected", map[string]interface{}{
		"peerId": peerId,
	})
}

func PeerDisconnected(peerId string, reason string) {
	Info("peer_disconnected", map[string]interface{}{
		"peerId": peerId,
		"reason": reason,
	})
}

func PeerAnnounced(peerId string, network string) {
	Info("peer_announced", map[string]interface{}{
		"peerId":  peerId,
		"network": network,
	})
}

func PeerDiscovered(peerId string, targetPeerId string, network string) {
	Info("peer_discovered", map[string]interface{}{
		"peerId":       peerId,
		"targetPeerId": targetPeerId,
		"network":      network,
	})
}

func HubConnected(hubId string, bootstrapUrl string) {
	Info("hub_connected", map[string]interface{}{
		"hubId":         hubId,
		"bootstrapUrl":  bootstrapUrl,
	})
}

func HubDisconnected(hubId string, reason string) {
	Info("hub_disconnected", map[string]interface{}{
		"hubId":  hubId,
		"reason": reason,
	})
}

func MessageRelayed(fromPeerId string, targetPeerId string, msgType string, network string) {
	Debug("message_relayed", map[string]interface{}{
		"fromPeerId":   fromPeerId,
		"targetPeerId": targetPeerId,
		"type":         msgType,
		"network":      network,
	})
}

func BootstrapAnnouncement(network string, peerCount int) {
	Info("bootstrap_announcement", map[string]interface{}{
		"network":   network,
		"peerCount": peerCount,
	})
}
