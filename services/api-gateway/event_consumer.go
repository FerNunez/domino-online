package main

import (
	"domino/shared/contracts"
	"domino/shared/messaging"
	"encoding/json"
	"log"
)

// WebsocketEventConsumer takes a domain event off RabbitMQ and pushes it down a websocket to a connected browser client.
type WebsocketEventConsumer struct {
	rb        *messaging.RabbitMQ
	connMgr   *ConnectionManager
	queueName string
}

func NewWebsocketEventConsumer(rb *messaging.RabbitMQ, connMgr *ConnectionManager, queueName string) *WebsocketEventConsumer {
	return &WebsocketEventConsumer{
		rb:        rb,
		connMgr:   connMgr,
		queueName: queueName,
	}
}

func (qc *WebsocketEventConsumer) Start() error {
	msgs, err := qc.rb.Channel.Consume(qc.queueName,
		"",
		true, // auto-ack: message is acknowledged immediately on delivery
		false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for msg := range msgs {
			// here msg is marshaled in bytes
			var envelope contracts.DominoEvent
			if err := json.Unmarshal(msg.Body, &envelope); err != nil {
				log.Println("WebsocketEventConsumer: failed to unmarshal envelope: ", err)
				continue
			}

			// Deserialize the Data field so the frontend receives a proper JSON object,
			// Not a base64-enconded string (what a raw []byte would produce)
			var payload any
			if envelope.Data != nil {
				if err := json.Unmarshal(envelope.Data, &payload); err != nil {
					log.Println("WebsocketEventConsumer: failed to unmarshal payload", err)
					continue
				}
			}

			wsMsg := contracts.WSMessage{
				Type: msg.RoutingKey,
				Data: payload,
			}

			// TargetID set -> directed to one player. Empty -> lobby-wide broadcast.
			if envelope.TargetID != "" {
				if err := qc.connMgr.SendMessage(envelope.TargetID, wsMsg); err != nil {
					log.Printf("WebsocketEventConsumer: failed to send to user %s: %v", envelope.TargetID, err)
				}
				continue
			}
			qc.connMgr.BroadcastToLobby(envelope.LobbyID, wsMsg)
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
