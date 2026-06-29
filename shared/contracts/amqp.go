package contracts

// AmqpMessage is the envelope for every message published to RabbitMQ
// OwnerID routes the message to the correct Websocket connection in the gateway
// Data is the raw JSON payload, ketp as bytes so eadch consumer can unmashal into own type
type AmqpMessage struct {
	OwnerID string `json:"ownerId"`
	Data    []byte `json:"data"` // NOTE: Array of bytes to unmarshal into own type struct
}

// NOTE:
// *.events.* = something happens
// *.cmd.* =instruction directed at a specific actor
const (
	// Trip events
	TripEventCreated              = "trip.event.created"
	TripEventDriverAssigned       = "trip.event.driver_assigned"
	TripEventNoDriversFound       = "trip.event.no_drivers_found"
	TripEventDriversNotInterested = "trip.event.drivers_not_interested"

	// Driver Commands
	DriverCmdTripRequest = "driver.cmd.trip_request"
	DriverCmdTripAccept  = "driver.cmd.trip_accepted"
	DriverCmdTripDecline = "driver.cmd.trip_decline"
	DriverCmdLocation    = "driver.cmd.location"
	DriverCmdRegister    = "driver.cmd.register"

	// Payments events
	PaymentEventSessionCreated = "payment.event.session_created"
	PaymentEventSuccess        = "payment.event.success"
	PaymentEventFailed         = "payment.event.failed"
	PaymentEventCancelled      = "payment.event.cancelled"

	// Payment commands
	PaymentCmdCreateSession = "payment.cmd.create_session"
)
