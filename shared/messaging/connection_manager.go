package messaging

//When a RabbitMQ message arrives addressed to user X,
//the gateway looks up X's WebSocket connection and writes the message to it.

import (
	"errors"
	"log"
	"net/http"
	"rebu/shared/contracts"
	"sync"

	"github.com/gorilla/websocket"
)

var ErrConnectionNotFound = errors.New(" connection not found")

// connWrapper pairs a WebSocket connectionw ith a per-connection mutex
// The WebSocket library is not concurrent-safe for writes — two goroutines
// writing to the same connection simultaneously causes a data race and a panic.
type connWrapper struct {
	conn  *websocket.Conn
	mutex sync.Mutex
}

// ConnectionManager is a thread-safe map from userID to WebSoccket connection
// The outer RWMutex protects the map itself (Add/R/G)
// The inner per-connection mutex protects concurrent writes to one connection
type ConnectionManager struct {
	connections map[string]*connWrapper
	mutex       sync.RWMutex
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins. FIX:for production
	},
}

func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[string]*connWrapper),
	}
}

// TODO: this does not need ot have reference to cm? or just to call it like an object func
// Creates Websocket upgrader and upgrades request
func (cm *ConnectionManager) Upgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	return upgrader.Upgrade(w, r, nil)
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
