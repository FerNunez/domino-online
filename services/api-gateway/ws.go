package main

import (
	"encoding/json"
	"log"
	"net/http"
	"rebu/services/api-gateway/grpc_clients"
	"rebu/shared/contracts"
	"rebu/shared/messaging"
	driver "rebu/shared/proto/driver"
)

// connManager is a package-level singleton shared by all WebSocket handlers
// All handlers in the same process share one connection map, enablig cross-handler message delivery(eg. a RabbitMQ consumer pushing to a rider whose connection was registered by handleRidersWebsocket
var connManager = messaging.NewConnectionManager()

func handleRidersWebsocket(w http.ResponseWriter, r *http.Request, rb *messaging.RabbitMQ) {
	conn, err := connManager.Upgrade(w, r)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	userID := r.URL.Query().Get("userID")
	if userID == "" {
		log.Println("WS /ws/riders: missing userID")
		return
	}

	connManager.Add(userID, conn)
	defer connManager.Remove(userID)

	// Start one queue consumer per queue par connected rider
	// Each consumer reads messages addressed to this userID and writes them to the rider's Websocket connection
	for _, q := range []string{
		messaging.NotifyDriverNoDriversFoundQueue,
		messaging.NotifyDriverAssignQueue,
		messaging.NotifyPaymentSessionCreatedQueue,
	} {
		consumer := messaging.NewQueueConsumer(rb, connManager, q)
		if err := consumer.Start(); err != nil {
			log.Printf("Failed to start consumer for %s: %v", q, err)
		}
	}

	// Read loop
	// Keeps the handler goroutine alive and drains any client messages
	// Riders do not currently send messages; the loop exists on WebSocket close
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

func handleDriversWebsocket(w http.ResponseWriter, r *http.Request, rb *messaging.RabbitMQ) {
	conn, err := connManager.Upgrade(w, r)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	userID := r.URL.Query().Get("userID")
	packageSlug := r.URL.Query().Get("packageSlug")

	if userID == "" || packageSlug == "" {
		log.Println("WS /ws/drivers: missing userID or packageSlug")
		return
	}

	connManager.Add(userID, conn)

	ctx := r.Context()

	driverSvc, err := grpc_clients.NewDriverServiceClient()
	if err != nil {
		log.Fatal(err)
	}

	// Deferred cleanup: runs when the WebSocket closes (driver disconnects or crashes)
	defer func() {
		connManager.Remove(userID)
		driverSvc.Client.UnregisterDriver(ctx, &driver.RegisterDriverRequest{
			DriverID:    userID,
			PackageSlug: packageSlug,
		})
		driverSvc.Close()
		log.Printf("Driver %s unregistered", userID)
	}()

	// Register the driver and send the registration data back to the browser
	// so the fronten can render the driver's car marker at the starttng position
	driverData, err := driverSvc.Client.RegisterDriver(ctx, &driver.RegisterDriverRequest{
		DriverID:    userID,
		PackageSlug: packageSlug,
	})
	if err != nil {
		log.Printf("RegisterDriver failed: %v", err)
		return
	}
	if err := connManager.SendMessage(userID, contracts.WSMessage{
		Type: contracts.DriverCmdRegister,
		Data: driverData.Driver,
	}); err != nil {
		log.Printf("Failed to send register response to driver: %s: %v", userID, err)
		return
	}

	// Subscribe to trip requests for this driver
	consumer := messaging.NewQueueConsumer(rb, connManager, messaging.DriverCmdTripRequestQueue)
	if err := consumer.Start(); err != nil {
		log.Printf("Failed to start driver trip request consumer: %v", err)
	}

	// Bidirectional read loop: the driver sends location updates and trip responses
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var msg struct {
			Type string          `json:"type"`
			Data json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Failed to unmarshal driver message: %v", err)
			continue
		}

		switch msg.Type {
		case contracts.DriverCmdLocation:
			// Location updates are no-op for now. Planned for future implementations
			continue
		case contracts.DriverCmdTripAccept, contracts.DriverCmdTripDecline:
			// Forward the driver's decision to RabbitMQ
			// Owner = userID so the trip service knows which driver reponded
			if err := rb.PublishMessage(ctx, msg.Type, contracts.AmqpMessage{
				OwnerID: userID,
				Data:    msg.Data, // json.RawMessage prserves exact bytes
			}); err != nil {
				log.Printf("Failed to publish driver response: %v", err)
			}
		default:
			log.Printf("Unknown driver message type :%s", msg.Type)
		}
	}

}
