package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	gorillaWebSocket "github.com/gorilla/websocket"
)

var upgrader = gorillaWebSocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

// WebSocketHandler returns a handler for WebSocket connections
func WebSocketHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get websocket manager from context
		state, exists := c.Get("state")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "State not available"})
			return
		}
		appState := state.(*AppState)

		// Upgrade to WebSocket
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to upgrade connection"})
			return
		}

		clientID := uuid.New().String()
		
		// Handle the connection in a goroutine
		go appState.WebSocketManager.HandleConnection(conn, clientID)
		
		// Return immediately, connection is handled asynchronously
		c.JSON(http.StatusOK, gin.H{"client_id": clientID, "status": "connected"})
	}
}
