package wslogger_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bay0/wslogger"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestWSLogger(t *testing.T) {
	// Create a new WSLogger
	wsl := wslogger.NewWSLogger()
	wsl.Start()

	// Set up a test server
	server := httptest.NewServer(http.HandlerFunc(wsl.HandleConnections))
	defer server.Close()

	// Replace "http://" with "ws://" to create WebSocket URL
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect to the WebSocket server
	dialer := websocket.Dialer{}
	ws, _, err := dialer.Dial(wsURL, nil)
	assert.NoError(t, err)
	defer ws.Close()

	// Create a WSWriter and write a test message
	wsw := wsl.NewWSWriter()
	testMessage := []byte("This is a test message")
	_, err = wsw.Write(testMessage)
	assert.NoError(t, err)

	// Read the message from the WebSocket
	_, receivedMessage, err := ws.ReadMessage()
	assert.NoError(t, err)

	// Check if the received message is the same as the sent message
	assert.Equal(t, testMessage, receivedMessage)
}
