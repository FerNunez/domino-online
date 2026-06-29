package contracts

import "encoding/json"

// NOTE: Creating/sending first message (uses any that then is marshaled json bytes)
// WS message is sent from the server to a WebSocket client
// Type mirrors the RabbitMQ routine key so the frontend can switch on it
type WSMessage struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

// NOTE: Forwarding message (already marshaled into json)
// WSDriverMessage is used when Data must be forwarded raw without deserialising
// The driver Websocket handler receives a message from the browser and re-publmishes
// it to RabbitMQ without touching the data bates
type WSDriverMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

//// DOC:
//  WSMessage.Data is any (outbound, you are constructing it)
//  You own the data. You have a Go struct or value and you want to send it over the wire.
//  Using any lets you assign any Go value and let json.Marshal serialize it for you:
//
//  msg := WSMessage{
//      Type: "location_update",
//      Data: LocationPayload{Lat: 37.7, Lng: -122.4}, // Go struct
//  }
//  // json.Marshal turns Data into {"lat":37.7,"lng":-122.4}
//
//  WSDriverMessage.Data is json.RawMessage (forwarding, you received it from elsewhere)
//  You received this JSON from a driver's WebSocket connection — it's already encoded bytes like
//  {"lat":37.7749,"lng":-122.4194}. You need to relay it to a rider verbatim.
//
//  If Data were any, the pipeline would be:
//  received bytes → Unmarshal into any → Marshal back to bytes
