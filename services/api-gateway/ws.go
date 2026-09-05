package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"domino/shared/contracts"
	"domino/shared/jwt"
	"domino/shared/messaging"
	pbg "domino/shared/proto/game"
	"domino/shared/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// NOTE: connManager is a package-level singleton shared by all WebSocket handlers
// All handlers in the same process share one connection map, enablig cross-handler message delivery(eg. a RabbitMQ consumer pushing to UserID whose connection was registered by handleLobbyWebsocket

func handleLobbyWebsocket(w http.ResponseWriter, r *http.Request, rmq *messaging.RabbitMQ) {
	// check inputs: lobbyID, claimID
	lobbyID := r.PathValue("id")
	if lobbyID == "" {
		http.Error(w, "lobby id is required", http.StatusBadRequest)
		return
	}
	claims, err := jwt.ParseLobbyTicket(r.URL.Query().Get("wsToken"))
	if err != nil {
		fmt.Printf("couldnt approuve ws token: %v", err)
		http.Error(w, "couldnt approuve ws token", http.StatusUnauthorized)
		return
	}
	if claims.LobbyID != lobbyID {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Upgrade
	conn, err := connManager.Upgrade(w, r)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// NOTE: broadcast/target events (lobby.*, game.*) reach connManager via the
	// single gateway-wide consumer started once in main.go, not per connection here.

	// check if there is a game running
	var stateSnapshot *pbg.GetGameStateResponse
	snapshotResp, err := gameClient.Client.GetGameState(r.Context(), &pbg.GetGameStateRequest{
		LobbyId: lobbyID,
		UserId:  claims.UserID,
	})
	if err != nil {
		if status.Code(err) != codes.NotFound {
			log.Printf("couldn't fetch game state for %s: %v", claims.UserID, err)
		}
		// codes.NotFound just means no game has started yet — expected while
		// still in the lobby. Either way, connect without a snapshot.
	} else {
		stateSnapshot = snapshotResp
	}

	// Add in Connection Manager
	connManager.AddToLobby(claims.LobbyID, claims.UserID)
	connManager.Add(claims.UserID, conn)

	// Send the snapshot as the first thing this connection receives
	if stateSnapshot != nil {
		WSMsg := contracts.WSMessage{
			Type: contracts.GameStateSync,
			Data: toGameStateSnapshot(stateSnapshot),
		}
		if err := connManager.SendMessage(claims.UserID, WSMsg); err != nil {
			log.Printf("couldn't send game state snapshot to %s: %v", claims.UserID, err)
		}
	}

	// Prepare msg to publish PlayerConnected
	playerConnectedMsg := messaging.PlayerConnectedData{
		UserID:  claims.UserID,
		LobbyID: claims.LobbyID,
	}
	data, err := json.Marshal(playerConnectedMsg)
	if err != nil {
		fmt.Printf("couldn't marshal joined player: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Publish player conenctedbefore connect
	if err := rmq.PublishMessage(r.Context(), contracts.PlayerConnected, &contracts.DominoEvent{
		LobbyID:  claims.LobbyID,
		TargetID: "",
		Data:     data,
	}); err != nil {
		fmt.Printf("couldn't notify player: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// This defer function runs when the websocket is dropt. Need to clean up and inform
	defer func() {
		// clean up:
		connManager.RemoveFromLobby(claims.LobbyID, claims.UserID)
		connManager.Remove(claims.UserID)
		// Prepare msg to publish Disconnect
		playerDisonnectedMsg := messaging.PlayerDisconnectedData{
			UserID:  claims.UserID,
			LobbyID: claims.LobbyID,
		}
		data, err := json.Marshal(playerDisonnectedMsg)
		if err != nil {
			log.Printf("couldn't marshal disconnected player: %v", err)
			return
		}
		// Publish player disconnected
		if err := rmq.PublishMessage(r.Context(), contracts.PlayerDisconnected, &contracts.DominoEvent{
			LobbyID:  claims.LobbyID,
			TargetID: "",
			Data:     data,
		}); err != nil {
			log.Printf("couldn't notify player disconnected: %v", err)
			return
		}
	}()

	// Read loop => read each user msg => check type and handle it accordingly
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		// Unmarshal into {Type and Data}
		type playerMessage struct {
			Type string          `json:"type"`
			Data json.RawMessage `json:"data"`
		}
		var playerMsg playerMessage
		if err := json.Unmarshal(message, &playerMsg); err != nil {
			log.Printf("Error unmarshaling player message: %v", err)
			continue
		}

		log.Printf("%v: got message %v", claims.UserID, playerMsg)

		// Handle by Msg Type
		switch playerMsg.Type {

		// -- Play a tile
		case contracts.PlayTileCmd:
			var cmd playTileCmd
			if err := json.Unmarshal(playerMsg.Data, &cmd); err != nil {
				log.Printf("Error unmarshaling play tile command: %v", err)
				continue
			}
			req := &pbg.PlayTileRequest{
				LobbyId: lobbyID,
				UserId:  claims.UserID,
				Tile:    &pbg.Tile{Left: cmd.Tile.Left, Right: cmd.Tile.Right},
				Side:    cmd.Side,
			}
			responseGrpc, err := gameClient.Client.PlayTile(r.Context(), req)
			if err != nil {
				log.Printf("couldn't play tile to server: %v", err)
				continue
			}

			response := messaging.PlayTileResponseData{
				Board:       toTiles(responseGrpc.Board),
				Hand:        toTiles(responseGrpc.Hand),
				RoundResult: toRoundResult(responseGrpc.RoundResult),
			}

			// Send response to WebSocket
			WSMsg := contracts.WSMessage{
				Type: contracts.PlayTileResponse,
				Data: response,
			}
			if err := connManager.SendMessage(claims.UserID, WSMsg); err != nil {
				log.Printf("couldt send message: %v", err)
			}

		// -- PassTurn
		case contracts.PassTurnCmd:
			req := &pbg.PassTurnRequest{
				LobbyId: lobbyID,
				UserId:  claims.UserID,
			}
			responseGrpc, err := gameClient.Client.PassTurn(r.Context(), req)
			if err != nil {
				log.Printf("couldn't pass turn to server: %v", err)
				continue
			}
			response := messaging.PassTurnResponseData{
				RoundResult: toRoundResult(responseGrpc.RoundResult),
			}

			// Send response to WebSocket
			WSMsg := contracts.WSMessage{
				Type: contracts.PassTurnResponse,
				Data: response,
			}
			if err := connManager.SendMessage(claims.UserID, WSMsg); err != nil {
				log.Printf("couldt send message: %v", err)
			}

		// -- PassTurn
		case contracts.NextRoundCmd:
			req := &pbg.NextRoundRequest{
				LobbyId: lobbyID,
				UserId:  claims.UserID,
			}
			responseGrpc, err := gameClient.Client.NextRound(r.Context(), req)
			if err != nil {
				log.Printf("couldn't request next round to server: %v", err)
				continue
			}
			response := messaging.NextRoundResponseData{
				RoundNumber: int(responseGrpc.RoundNumber),
			}

			// Send response to WebSocket
			WSMsg := contracts.WSMessage{
				Type: contracts.NextRoundResponse,
				Data: response,
			}
			if err := connManager.SendMessage(claims.UserID, WSMsg); err != nil {
				log.Printf("couldt send message: %v", err)
			}
		default:
			log.Printf("Unknown message type: %s", playerMsg.Type)
		}
	}
}

func toGameStateSnapshot(resp *pbg.GetGameStateResponse) messaging.GameStateSnapshotData {
	teamScores := make(map[types.TeamID]int, len(resp.TeamScores))
	for team, score := range resp.TeamScores {
		teamScores[types.TeamID(team)] = int(score)
	}
	handSizes := make(map[string]int, len(resp.HandSizes))
	for playerID, size := range resp.HandSizes {
		handSizes[playerID] = int(size)
	}

	return messaging.GameStateSnapshotData{
		GameID:      resp.GameId,
		GameNumber:  int(resp.GameNumber),
		Status:      resp.Status,
		TeamScores:  teamScores,
		TeamWinner:  types.TeamID(resp.TeamWinner),
		GoalScore:   int(resp.GoalScore),
		RoundID:     resp.RoundId,
		RoundNumber: int(resp.RoundNumber),
		RoundStatus: resp.RoundStatus,
		PlayerOrder: resp.PlayerOrder,
		CurrentTurn: resp.CurrentTurn,
		Board: messaging.BoardData{
			Tiles:    toTiles(resp.Board.Tiles),
			LeftEnd:  int(resp.Board.LeftEnd),
			RightEnd: int(resp.Board.RightEnd),
		},
		Hand:        toTiles(resp.Hand),
		HandSizes:   handSizes,
		RoundResult: toRoundResult(resp.RoundResult),
	}
}

func toTiles(tiles []*pbg.Tile) []types.Tile {
	out := make([]types.Tile, len(tiles))
	for idx, t := range tiles {
		out[idx] = types.Tile{
			Left:  int(t.Left),
			Right: int(t.Right),
		}
	}
	return out
}

// Converts a proto RoundResult to a type Round RoundResult. It can be nil
func toRoundResult(roundResult *pbg.RoundResult) *types.RoundResult {
	if roundResult == nil {
		return nil
	}

	pipCounts := make(map[types.TeamID]int, len(roundResult.PipCounts))
	for key, val := range roundResult.PipCounts {
		pipCounts[types.TeamID(key)] = int(val)
	}

	return &types.RoundResult{
		WinnerTeamID: types.TeamID(roundResult.WinnerId),
		Reason:       types.Reason(roundResult.Reason),
		PipCounts:    pipCounts,
	}
}
