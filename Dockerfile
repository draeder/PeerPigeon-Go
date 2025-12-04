FROM golang:1.22-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o peerpigeon ./cmd/peerpigeon

FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /build/peerpigeon .

EXPOSE 8080

ENV PORT=8080 \
    HOST=0.0.0.0 \
    MAX_CONNECTIONS=1000 \
    CORS_ORIGIN=* \
    HUB_MESH_NAMESPACE=pigeonhub-mesh \
    IS_HUB=false \
    BOOTSTRAP_HUBS="" \
    AUTH_TOKEN=""

CMD ["./peerpigeon"]
