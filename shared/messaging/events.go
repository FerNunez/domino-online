package messaging

import (
	pbd "rebu/shared/proto/driver"
	pbt "rebu/shared/proto/trip"
)

// Queue name constants: single source of truth
const (
	FindAvailableDriversQueue        = "find_available_drivers_queue"
	DriverCmdTripRequestQueue        = "driver_cmd_trip_request_queue"
	DriverTripResponseQueue          = "driver_trip_response_queue"
	NotifyDriverNoDriversFoundQueue  = "notify_driver_no_drivers_found_queue"
	NotifyDriverAssignQueue          = "notify_driver_assign_queue"
	PaymentTripResponseQueue         = "payment_trip_response_queue"
	NotifyPaymentSessionCreatedQueue = "notify_payment_session_created_queue"
	NotifyPaymentSuccessQueue        = "notify_payment_success_queue"
	DeaedLetterQueue                 = "deaed_letter_queue"
)

// TripEventData is the payload data for trip.event.created
type TripEventData struct {
	Trip *pbt.Trip `json:"trip"`
}

// DriverTripResponseData is the payload data for driver.cmd.trip_accept/trip decline
type DriverTripResponseData struct {
	Driver  *pbd.Driver `json:"driver"`
	TripID  string      `json:"tripID"`
	RiderID string      `json:"driverID"`
}

// PaymentEventSessionCreatedData is the playload for the payment.event.session_created
type PaymentEventSessionCreatedData struct {
	TripID    string  `json:"tripID"`
	SessionID string  `json:"sessionID"`
	Amount    float64 `json:"amout"`
	Currency  string  `json:"currency"`
}

// PaymentTripResponseData is the payload for payment.cmd.create_session
type PaymentTripResponseData struct {
	TripID   string  `json:"tripID"`
	UserID   string  `json:"userID"`
	DriverID string  `json:"driverID"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// PaymentStatusUpdateData is the playload for the payment.event.success
type PaymentStatusUpdateData struct {
	TripID   string `json:"tripID"`
	UserID   string `json:"userID"`
	DriverID string `json:"driverID"`
}

// NOTE:
//Why payload types are here and not in each service: The DriverTripResponseData struct is produced by the API gateway (which reads it from the driver's WebSocket) and consumed by the trip service. Both need the same struct definition. Defining it here avoids duplication and drift between the two.
// ex: User -(WebSocket)-> ApiGateway -(publish)-> Queues
