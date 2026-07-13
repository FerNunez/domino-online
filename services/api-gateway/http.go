package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"rebu/services/api-gateway/grpc_clients"
	"rebu/shared/contracts"
	"rebu/shared/env"
	"rebu/shared/jwt"
	"rebu/shared/messaging"
	pbu "rebu/shared/proto/user"
	"rebu/shared/tracing"

	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/webhook"
)

var tracer = tracing.GetTracer("api-getaway")

func handleTripPreview(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "handleTripReview")
	defer span.End()

	var req previewTripRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "failed to parse JSON", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.UserID == "" {
		http.Error(w, "userID is required", http.StatusBadRequest)
		return
	}

	tripSvc, err := grpc_clients.NewTripServiceClient()
	if err != nil {
		log.Fatal(err)
	}
	defer tripSvc.Close()

	preview, err := tripSvc.Client.PreviewTrip(ctx, req.toProto())
	if err != nil {
		log.Printf("PreviewTrip failed: %v", err)
		http.Error(w, "failed to preview trip", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, contracts.APIResponse{Data: preview})
}

func handleCreateLobby(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "handleCreateLobby")
	defer span.End()

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

	lobby, err := lobbySvc.Client.CreateLobby(ctx, req.toProto())
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to create lobby: %v", err), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, contracts.APIResponse{Data: lobby})
}

func handleCreateGuest(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "handleCreateGuest")
	defer span.End()

	userSvc, err := grpc_clients.NewUserServiceClient()
	if err != nil {
		log.Fatal(err)
	}

	user, err := userSvc.Client.CreateGuest(ctx, &pbu.CreateGuestRequest{})
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to create guest user: %v", err), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, contracts.APIResponse{Data: user})
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

	lobbyID := r.PathValue("id")
	if lobbyID == "" {
		http.Error(w, "lobby id is required", http.StatusBadRequest)
		return
	}
	var req joinLobbyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "failed to parse JSON", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	req.LobbyID = lobbyID

	lobbySvc, err := grpc_clients.NewLobbyServiceClient()
	if err != nil {
		log.Fatal(err)
	}

	lobby, err := lobbySvc.Client.JoinLobby(ctx, req.toProto())
	if err != nil {
		http.Error(w, "failed to join lobby", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, contracts.APIResponse{Data: lobby})
}

func handleStartGame(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "handleStartGame")
	defer span.End()

	lobbyID := r.PathValue("id")
	if lobbyID == "" {
		http.Error(w, "lobby id is required", http.StatusBadRequest)
		return
	}

	var req startGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "failed to parse JSON", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	req.LobbyID = lobbyID

	lobbySvc, err := grpc_clients.NewLobbyServiceClient()
	if err != nil {
		log.Fatal(err)
	}

	// TODO: This only have the game ID?
	// Should I here call the game service grpc to get state or something?
	game, err := lobbySvc.Client.StartGame(ctx, req.toProto())
	if err != nil {
		http.Error(w, "failed to start game", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, contracts.APIResponse{Data: game})
}

func handleTripStart(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "handleTripStart")
	defer span.End()

	var req startTripRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "failed to parse JSON", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	tripSvc, err := grpc_clients.NewTripServiceClient()
	if err != nil {
		log.Fatal(err)
	}
	defer tripSvc.Close()

	trip, err := tripSvc.Client.CreateTrip(ctx, req.toProto())
	if err != nil {
		log.Printf("CreateTrip failed: %v", err)
		http.Error(w, "failed to start trip", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, contracts.APIResponse{Data: trip})
}

func handleStripeWebhook(w http.ResponseWriter, r *http.Request, rb *messaging.RabbitMQ) {
	ctx, span := tracer.Start(r.Context(), "handleStripeWebhook")
	defer span.End()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	webhookSecret := env.GetString("STRIPE_WEBHOOK_KEY", "")
	if webhookSecret == "" {
		log.Println("STRIPE_WEBHOOK_KEY not set, skipping verification")
		return
	}

	event, err := webhook.ConstructEventWithOptions(
		body,
		r.Header.Get("Strip-Signature"),
		webhookSecret,
		webhook.ConstructEventOptions{IgnoreAPIVersionMismatch: true},
	)
	if err != nil {
		log.Printf("Stripe webhook signature invalid: %v", err)
		http.Error(w, "invalid signature", http.StatusBadRequest)
		return
	}

	var sess stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &sess); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	payload := messaging.PaymentStatusUpdateData{
		TripID:   sess.Metadata["trip_id"],
		UserID:   sess.Metadata["user_id"],
		DriverID: sess.Metadata["driver_id"],
	}

	data, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "marshal failed", http.StatusInternalServerError)
		return
	}

	if err := rb.PublishMessage(ctx, contracts.PaymentEventSuccess, contracts.AmqpMessage{
		OwnerID: sess.Metadata["user_id"],
		Data:    data,
	}); err != nil {
		log.Printf("Failed to publish payment.event.success: %v", err)
		http.Error(w, "publish failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
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

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	// NOTE: Do I need to check the userType == guest?

	userSvc, err := grpc_clients.NewUserServiceClient()
	if err != nil {
		fmt.Printf("couldnt reach user service: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	auth, err := userSvc.Client.Register(ctx, &pbu.RegisterRequest{
		UserID:      userID,
		Password:    "password",
		DisplayName: "displayNmae",
	})
	if err != nil {
		fmt.Printf("couldnt register user: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Printf("User: %v", auth.User)
	w.WriteHeader(http.StatusCreated)

	writeJSON(w, http.StatusCreated, contracts.APIResponse{Data: auth.User})
}
