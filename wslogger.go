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

type WSLogger struct {
	upgrader  websocket.Upgrader
	clients   sync.Map
	broadcast chan []byte
}

func NewWSLogger() *WSLogger {
	return &WSLogger{
		upgrader:  websocket.Upgrader{},
		broadcast: make(chan []byte, bufferSize),
	}
}

func (wsl *WSLogger) HandleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := wsl.upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to open WebSocket connection", http.StatusInternalServerError)
		return
	}
	defer ws.Close()

	client := &Client{
		conn: ws,
		send: make(chan []byte, bufferSize),
	}
	wsl.clients.Store(client, struct{}{})

	client.run()

	wsl.clients.Delete(client)
}

func (wsl *WSLogger) handleMessages() {
	for msg := range wsl.broadcast {
		wsl.clients.Range(func(key, value interface{}) bool {
			client := key.(*Client)
			select {
			case client.send <- msg:
			default:
				close(client.send)
				wsl.clients.Delete(client)
			}
			return true
		})
	}
}

type WSWriter struct {
	wsLogger *WSLogger
	mutex    sync.RWMutex
	closed   bool
}

func (wsl *WSLogger) NewWSWriter() *WSWriter {
	return &WSWriter{wsLogger: wsl}
}

func (wsw *WSWriter) Write(p []byte) (n int, err error) {
	wsw.mutex.RLock()
	defer wsw.mutex.RUnlock()

	if wsw.closed {
		return 0, io.ErrClosedPipe
	}

	msg := make([]byte, len(p))
	copy(msg, p)
	wsw.wsLogger.broadcast <- msg

	return len(p), nil
}

func (wsw *WSWriter) Close() error {
	wsw.mutex.Lock()
	defer wsw.mutex.Unlock()

	if wsw.closed {
		return io.ErrClosedPipe
	}

	wsw.closed = true
	return nil
}

func (wsl *WSLogger) Start() {
	go wsl.handleMessages()
}

type Client struct {
	conn *websocket.Conn
	send chan []byte
}

func (c *Client) run() {
	go c.readPump()
	c.writePump()
}

func (c *Client) readPump() {
	defer c.conn.Close()
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (c *Client) writePump() {
	for msg := range c.send {
		err := c.conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			break
		}
	}
}
