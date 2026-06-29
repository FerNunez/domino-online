package messaging

import (
	"encoding/json"
	"log"
	"rebu/shared/contracts"
)

// QueueConsumer reads from a rabbitMQ queue and forwards msg to the Websocket client indentified by
// AmqpMessage.OwnerID
type QueueConsumer struct {
	rb        *RabbitMQ
	connMgr   *ConnectionManager
	queueName string
}

func NewQueueConsumer(rb *RabbitMQ, connMgr *ConnectionManager, queueName string) *QueueConsumer {
	return &QueueConsumer{
		rb:        rb,
		connMgr:   connMgr,
		queueName: queueName,
	}
}

func (qc *QueueConsumer) Start() error {
	msgs, err := qc.rb.Channel.Consume(qc.queueName,
		"",
		true, // auto-ack: message is acknowledged immediately on delivery
		false, false, false, nil)
	if err != nil {
		return nil
	}

	go func() {
		for msg := range msgs {
			var envelope contracts.AmqpMessage
			if err = json.Unmarshal(msg.Body, &envelope); err != nil {
				log.Println("QueueConsumer: failed to unmarshal envelope: ", err)
				continue
			}

			// Deserialize the Data field so the frontend receives a proper JSON object,
			// Not a base64-enconded string (what a raw []byte would produce)
			var payload any
			if envelope.Data != nil {
				if err := json.Unmarshal(envelope.Data, &payload); err != nil {
					log.Println("QueueConsumer: failed to unmarshal payload", err)
					continue
				}
			}

			wsMsg := contracts.WSMessage{
				Type: msg.RoutingKey,
				Data: payload,
			}
			if err := qc.connMgr.SendMessage(envelope.OwnerID, wsMsg); err != nil {
				log.Printf("QueueConsumer: failed to send to user %s: %v", envelope.OwnerID, err)
			}
		}
	}()
	return nil
}

//   ---
// The full pipeline
//
// Step 1 — Publisher builds AmqpMessage
//
// Some service (e.g. trip service) does something like:
//
// payload, _ := json.Marshal(TripCreatedEvent{TripID: "abc", Status: "created"})
// // payload is now []byte → the raw UTF-8 JSON bytes: {"tripId":"abc","status":"created"}
//
// msg := contracts.AmqpMessage{
//     OwnerID: "1",
//     Data:    payload,   // []byte containing valid JSON
// }
//
// So Data is already-serialized JSON, stored as a byte slice.
//
// ---
// Step 2 — PublishMessage marshals the envelope
//
// jsonMsg, _ := json.Marshal(message)  // marshal the AmqpMessage struct
//
// This is the critical point you're missing. Go's encoding/json has a rule: []byte fields are always base64-encoded when
// marshaled to JSON (because JSON has no binary type). So the wire bytes sent to RabbitMQ look like:
//
// {"ownerId":"1","data":"eyJ0cmlwSWQiOiJhYmMiLCJzdGF0dXMiOiJjcmVhdGVkIn0="}
//
// data is now a base64 string, not a JSON object.
//
// ---
// Step 3 — Consumer's first unmarshal (recovers the envelope)
//
// var envelope contracts.AmqpMessage
// json.Unmarshal(msg.Body, &envelope)
//
// Because Data is declared as []byte, the JSON decoder base64-decodes that string back to the original bytes. So envelope.Data
// is now []byte containing {"tripId":"abc","status":"created"} — valid JSON again.
//
// ---
// Step 4 — Why the second unmarshal?
//
// If you skipped it and did:
//
// wsMsg := contracts.WSMessage{Type: ..., Data: envelope.Data}  // Data is []byte
// json.Marshal(wsMsg)  // ← what gets sent to WebSocket
//
// The Data any field holds a []byte value, so json.Marshal would base64-encode it again. The frontend would receive:
//
// {"type":"trip.event.created","data":"eyJ0cmlwSWQiOiJhYmMiLCJzdGF0dXMiOiJjcmVhdGVkIn0="}
//
// A base64 string — useless to the frontend.
//
// The second unmarshal breaks the cycle:
//
// var payload any
// json.Unmarshal(envelope.Data, &payload)
// // payload is now map[string]interface{}{"tripId": "abc", "status": "created"}
//
// wsMsg := contracts.WSMessage{Type: ..., Data: payload}
// json.Marshal(wsMsg)
// // → {"type":"trip.event.created","data":{"tripId":"abc","status":"created"}}
//
// Now Data is a Go map, not a byte slice, so it marshals as a proper JSON object.
