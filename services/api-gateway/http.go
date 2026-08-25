package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"domino/services/api-gateway/grpc_clients"
	"domino/shared/contracts"
	"domino/shared/jwt"
	pbh "domino/shared/proto/history"
	pbl "domino/shared/proto/lobby"
	pbu "domino/shared/proto/user"
	"domino/shared/tracing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var tracer = tracing.GetTracer("api-gateway")

// historySvc is a single long-lived connection shared across requests
var historySvc, historySvcErr = grpc_clients.NewHistoryServiceClient()

func handleCreateLobby(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "handleCreateLobby")
	defer span.End()

	userID, ok := ctx.Value("userID").(string)
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var req createLobbyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "failed to parse JSON", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	lobbySvc, err := grpc_clients.NewLobbyServiceClient()
	if err != nil {
		log.Fatal(err)
	}

	protoReq := newProtoCreateLobbyRequest(&req, userID)
	lobby, err := lobbySvc.Client.CreateLobby(ctx, protoReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to create lobby: %v", err), http.StatusInternalServerError)
		return
	}

	tokenWs, err := jwt.NewLobbyTicket(lobby.Lobby.Id, userID)
	if err != nil {
		log.Printf("couldn't create jwt lobby ticket for user: %v and lobbyId: %v, err: %s\n", userID, lobby.Lobby.Id, err)
		http.Error(w, "couldnt create lobby", http.StatusInternalServerError)
		return
	}

	// Generate jwt from lobby
	response := newProtoCreateLobbyResponse(lobby.Lobby.Id, tokenWs)

	// Then in join, check the connection claims if they are real. when you invite a friend I guess you copy this join/ with body a wsToken and wsURL
	// Should we encode wsURL in the token? what if someone tries to use a valid token to connect to another WsURL?
	writeJSON(w, http.StatusCreated, contracts.APIResponse{Data: response})
}

func handleGetUser(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "handleGetUser")
	defer span.End()

	userSvc, err := grpc_clients.NewUserServiceClient()
	if err != nil {
		log.Fatal(err)
	}

	userID := r.PathValue("id")
	if userID == "" {
		http.Error(w, "user id is required", http.StatusBadRequest)
		return
	}

	user, err := userSvc.Client.GetUser(ctx, &pbu.GetUserRequest{
		UserID: userID,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("ailed to get user: %v", err), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, contracts.APIResponse{Data: user})
}

func handleJoinLobby(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "handleJoinLobby")
	defer span.End()

	userID, ok := ctx.Value("userID").(string)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	lobbyID := r.PathValue("id")
	if lobbyID == "" {
		http.Error(w, "lobby id is required", http.StatusBadRequest)
		return
	}

	lobbySvc, err := grpc_clients.NewLobbyServiceClient()
	if err != nil {
		log.Fatal(err)
	}

	protoReq := newProtoJoinLobbyRequest(userID, lobbyID)
	lobby, err := lobbySvc.Client.JoinLobby(ctx, protoReq)
	if err != nil {
		http.Error(w, "failed to join lobby", http.StatusInternalServerError)
		return
	}

	tokenWs, err := jwt.NewLobbyTicket(lobby.Lobby.Id, userID)
	if err != nil {
		log.Printf("couldn't create jwt lobby ticket for user: %v and lobbyId: %v, err: %s\n", userID, lobby.Lobby.Id, err)
		http.Error(w, "couldnt create lobby", http.StatusInternalServerError)
		return
	}
	response := newProtoCreateLobbyResponse(lobby.Lobby.Id, tokenWs)
	writeJSON(w, http.StatusCreated, contracts.APIResponse{Data: response})
}

func handleStartGame(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "handleStartGame")
	defer span.End()

	userID, ok := ctx.Value("userID").(string)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	lobbyID := r.PathValue("id")
	if lobbyID == "" {
		http.Error(w, "lobby id is required", http.StatusBadRequest)
		return
	}

	req := startLobbyRequest{
		LobbyID: lobbyID,
		UserID:  userID,
	}

	lobbySvc, err := grpc_clients.NewLobbyServiceClient()
	if err != nil {
		log.Fatal(err)
	}

	game, err := lobbySvc.Client.StartLobby(ctx, req.toProto())
	if err != nil {
		fmt.Printf("error: %v\n", err)
		http.Error(w, "failed to start game in server", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, contracts.APIResponse{Data: game})
}

func handleGetLobby(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "handleGetLobby")
	defer span.End()

	lobbyID := r.PathValue("id")
	if lobbyID == "" {
		http.Error(w, "lobby id is required", http.StatusBadRequest)
		return
	}

	lobbySvc, err := grpc_clients.NewLobbyServiceClient()
	if err != nil {
		log.Fatal(err)
	}

	lobby, err := lobbySvc.Client.GetLobby(ctx, &pbl.GetLobbyRequest{LobbyID: lobbyID})
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get lobby: %v", err), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, contracts.APIResponse{Data: lobby.Lobby})
}

// returns []actions and []hands for a round
func handleGetRoundActions(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "handleGetRoundActions")
	defer span.End()

	roundID := r.PathValue("id")
	if roundID == "" {
		http.Error(w, "round id is required", http.StatusBadRequest)
		return
	}

	actions, err := historySvc.Client.GetRoundActions(ctx, &pbh.GetRoundActionsRequest{RoundId: roundID})
	if err != nil {
		// NotFound means the round hasn't been fully persisted by history-service yet
		if status.Code(err) == codes.NotFound {
			http.Error(w, "round history not available yet", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("failed to get round actions: %v", err), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, contracts.APIResponse{Data: struct {
		Actions []*pbh.Action `json:"actions"`
		Hands   []*pbh.Hand   `json:"hands"`
	}{Actions: actions.Actions, Hands: actions.Hands}})
}

// returns []GameSumary for a user
func handleGetPlayerGames(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "handleGetPlayerGames")
	defer span.End()

	userID, ok := ctx.Value("userID").(string)
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	games, err := historySvc.Client.GetPlayerGames(ctx, &pbh.GetPlayerGamesRequest{PlayerId: userID})
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get player games: %v", err), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, contracts.APIResponse{Data: games.Games})
}

// GameHistory means []RoundSumary for a game
func handleGetGameHistory(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "handleGetGameHistory")
	defer span.End()

	gameID := r.PathValue("id")
	if gameID == "" {
		http.Error(w, "game id is required", http.StatusBadRequest)
		return
	}

	history, err := historySvc.Client.GetGameHistory(ctx, &pbh.GetGameHistoryRequest{GameId: gameID})
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get game history: %v", err), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, contracts.APIResponse{Data: history.Rounds})
}

func handleAuthGuest(w http.ResponseWriter, r *http.Request) {
	_, span := tracer.Start(r.Context(), "handleAuthGuest")
	defer span.End()

	// Guest
	token, err := jwt.NewGuestToken()
	if err != nil {
		http.Error(w, "guest jwt generation failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Authorization", fmt.Sprintf("Bearer %v", token))

	writeJSON(w, http.StatusOK, contracts.APIResponse{Data: map[string]string{
		"token": token,
	}})
}

func handleAuthRegiter(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "handleAuthRegister")
	defer span.End()

	userID, ok := ctx.Value("userID").(string)
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "couldnt register", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	userSvc, err := grpc_clients.NewUserServiceClient()
	if err != nil {
		fmt.Printf("couldnt reach user service: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	auth, err := userSvc.Client.Register(ctx, &pbu.RegisterRequest{
		UserID:      userID,
		Password:    req.Password,
		DisplayName: req.DisplayName,
		Email:       req.Email,
	})
	if err != nil {
		fmt.Printf("couldnt register user: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, contracts.APIResponse{Data: auth.User})
}
