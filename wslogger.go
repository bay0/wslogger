package wslogger

import (
	"io"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

const (
	bufferSize = 1000
)

// WSLogger is a WebSocket-based logger that broadcasts log messages to connected clients.
type WSLogger struct {
	upgrader     websocket.Upgrader       // Upgrades HTTP connections to WebSocket connections
	clients      map[*websocket.Conn]bool // Stores WebSocket connections of all connected clients
	broadcast    chan []byte              // Channel for broadcasting log messages to connected clients
	bufferedLogs chan []byte              // Channel for buffering log messages
}

// NewWSLogger creates a new WSLogger instance.
func NewWSLogger() *WSLogger {
	return &WSLogger{
		upgrader:     websocket.Upgrader{},
		clients:      make(map[*websocket.Conn]bool),
		broadcast:    make(chan []byte, bufferSize),
		bufferedLogs: make(chan []byte),
	}
}

// HandleConnections upgrades the HTTP connection to a WebSocket connection and manages
// the connected clients.
func (wsl *WSLogger) HandleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := wsl.upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to open WebSocket connection", http.StatusInternalServerError)
		return
	}
	defer ws.Close()

	wsl.clients[ws] = true

	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			delete(wsl.clients, ws)
			break
		}
	}
}

// handleMessages reads log messages from the broadcast channel and sends them to all
// connected clients.
func (wsl *WSLogger) handleMessages() {
	for {
		msg := <-wsl.broadcast
		for client := range wsl.clients {
			err := client.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				client.Close()
				delete(wsl.clients, client)
			}
		}
	}
}

// Close closes the WSWriter, preventing further writes.
func (wsw *WSWriter) Close() error {
	wsw.mutex.Lock()
	defer wsw.mutex.Unlock()

	if wsw.closed {
		return io.ErrClosedPipe
	}

	wsw.closed = true
	return nil
}

// WSWriter is an io.Writer implementation that writes log messages to a WSLogger's
// broadcast channel.
type WSWriter struct {
	wsLogger *WSLogger
	mutex    sync.Mutex
	closed   bool
}

// NewWSWriter creates a new WSWriter instance for the given WSLogger.
func (wsl *WSLogger) NewWSWriter() *WSWriter {
	return &WSWriter{wsLogger: wsl}
}

// Write writes the given byte slice (log message) to the WSLogger's broadcast channel.
func (wsw *WSWriter) Write(p []byte) (n int, err error) {
	wsw.mutex.Lock()
	defer wsw.mutex.Unlock()

	if wsw.closed {
		return 0, io.ErrClosedPipe
	}

	msg := make([]byte, len(p))
	copy(msg, p)
	wsw.wsLogger.broadcast <- msg

	return len(p), nil
}

// Start starts the message handling loop in a separate goroutine.
func (wsl *WSLogger) Start() {
	go wsl.handleMessages()
}
