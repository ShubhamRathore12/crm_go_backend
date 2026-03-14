package websocket

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

// Client represents a WebSocket client
type Client struct {
	ID   string
	Conn *websocket.Conn
	Send chan []byte
}

// Manager manages WebSocket connections
type Manager struct {
	Clients    map[string]*Client
	mutex      sync.RWMutex
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan []byte
}

// NewManager creates a new WebSocket manager
func NewManager() *Manager {
	return &Manager{
		Clients:    make(map[string]*Client),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan []byte, 100),
	}
}

// Run starts the WebSocket manager
func (m *Manager) Run() {
	for {
		select {
		case client := <-m.Register:
			m.mutex.Lock()
			m.Clients[client.ID] = client
			m.mutex.Unlock()
			log.Printf("WebSocket client connected: %s", client.ID)

		case client := <-m.Unregister:
			m.mutex.Lock()
			if _, ok := m.Clients[client.ID]; ok {
				delete(m.Clients, client.ID)
				close(client.Send)
			}
			m.mutex.Unlock()
			log.Printf("WebSocket client disconnected: %s", client.ID)

		case message := <-m.Broadcast:
			m.mutex.RLock()
			for _, client := range m.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(m.Clients, client.ID)
				}
			}
			m.mutex.RUnlock()
		}
	}
}

// SendToClient sends a message to a specific client
func (m *Manager) SendToClient(clientID string, message []byte) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if client, ok := m.Clients[clientID]; ok {
		select {
		case client.Send <- message:
		default:
			log.Printf("Failed to send message to client %s", clientID)
		}
	}
}

// BroadcastToAll sends a message to all connected clients
func (m *Manager) BroadcastToAll(message []byte) {
	select {
	case m.Broadcast <- message:
	default:
		log.Println("Broadcast channel full")
	}
}

// HandleConnection handles a WebSocket connection
func (m *Manager) HandleConnection(conn *websocket.Conn, clientID string) {
	client := &Client{
		ID:   clientID,
		Conn: conn,
		Send: make(chan []byte, 256),
	}

	m.Register <- client

	// Start writer
	go func() {
		defer func() {
			m.Unregister <- client
			conn.Close()
		}()

		for message := range client.Send {
			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Error writing to client %s: %v", clientID, err)
				return
			}
		}
	}()

	// Start reader
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			m.Unregister <- client
			break
		}

		// Handle incoming message
		m.handleMessage(client, message)
	}
}

// handleMessage processes incoming WebSocket messages
func (m *Manager) handleMessage(client *Client, message []byte) {
	log.Printf("Received message from %s: %s", client.ID, string(message))
	// TODO: Implement message handling logic
}
