package messaging

//When a RabbitMQ message arrives addressed to user X,
//the gateway looks up X's WebSocket connection and writes the message to it.

import (
	"domino/shared/contracts"
	"errors"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var ErrConnectionNotFound = errors.New(" connection not found")

// connWrapper pairs a WebSocket connectionw ith a per-connection mutex
// The WebSocket library is not concurrent-safe for writes — two goroutines
// writing to the same connection simultaneously causes a data race and a panic.
type connWrapper struct {
	user  string
	conn  *websocket.Conn
	mutex sync.Mutex
}

// ConnectionManager is a thread-safe map from userID to WebSoccket connection
// The outer RWMutex protects the map itself (Add/R/G)
// The inner per-connection mutex protects concurrent writes to one connection
type ConnectionManager struct {
	connections map[string]*connWrapper
	mutex       sync.RWMutex

	lobbyMembers map[string]map[string]struct{}
	lobbyMutex   sync.RWMutex
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins. FIX:for production
	},
}

func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections:  make(map[string]*connWrapper),
		lobbyMembers: make(map[string]map[string]struct{}),
	}
}

// Creates Websocket upgrader and upgrades request
func (cm *ConnectionManager) Upgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	return upgrader.Upgrade(w, r, nil)
}

func (cm *ConnectionManager) AddToLobby(lobbyID, userID string) {
	cm.lobbyMutex.Lock()
	defer cm.lobbyMutex.Unlock()

	if cm.lobbyMembers[lobbyID] == nil {
		cm.lobbyMembers[lobbyID] = make(map[string]struct{})
	}
	cm.lobbyMembers[lobbyID][userID] = struct{}{}
}

func (cm *ConnectionManager) RemoveFromLobby(lobbyID, userID string) {
	cm.lobbyMutex.Lock()
	defer cm.lobbyMutex.Unlock()

	delete(cm.lobbyMembers[lobbyID], userID)
	if len(cm.lobbyMembers[lobbyID]) == 0 {
		delete(cm.lobbyMembers, lobbyID)
	}
}

func (cm *ConnectionManager) BroadcastToLobby(lobbyID string, message contracts.WSMessage) {
	cm.lobbyMutex.RLock()

	userList := make([]string, 0, len(cm.lobbyMembers[lobbyID]))
	for user := range cm.lobbyMembers[lobbyID] {
		userList = append(userList, user)
	}
	defer cm.lobbyMutex.RUnlock()

	for _, userID := range userList {
		if err := cm.SendMessage(userID, message); err != nil {
			log.Printf("broadcast to lobby %s: user %s: %v", lobbyID, userID, err)
		}
	}
}

// Add
// Remove
// Get
func (cm *ConnectionManager) Add(id string, conn *websocket.Conn) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.connections[id] = &connWrapper{conn: conn}
	log.Printf("Added WebSocket connection for user %s", id)
}

func (cm *ConnectionManager) Remove(id string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	delete(cm.connections, id)
	log.Printf("Removed WebSocket connection for users %s", id)
}

// Get returns (connection, true) of id string, if it is found. Else (nil, false)
func (cm *ConnectionManager) Get(id string) (*websocket.Conn, bool) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	conn, exists := cm.connections[id]
	if !exists {
		return nil, false
	}
	return conn.conn, true
}

// SendMessage sends a message to connection of Id
func (cm *ConnectionManager) SendMessage(id string, message contracts.WSMessage) error {
	// Lock manager to read
	cm.mutex.RLock()
	wrapper, exists := cm.connections[id]
	cm.mutex.RUnlock()

	if !exists {
		return ErrConnectionNotFound
	}

	// Lock the connection mutex
	// Lock the per-connection mutex before writing.
	// The outer RWMutex is released first to avoid holding it during I/O.
	wrapper.mutex.Lock()
	defer wrapper.mutex.Unlock()
	// Write JSON into connection
	return wrapper.conn.WriteJSON(message)
}

// Two-level locking:
// 1. cm.mutex (sync.RWMutex) protects the map. Multiple goroutines can read (look up connections) concurrently (RLock), but only one can write (add/remove) at a time (Lock).
// 2. wrapper.mutex (sync.Mutex) protects individual WebSocket connections. The gorilla/websocket library requires that concurrent writes to a single connection be serialised. Without this lock, two RabbitMQ consumers both trying to push to the same rider's WebSocket simultaneously would cause a panic
