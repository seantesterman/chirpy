package main

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/seantesterman/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerPolkaWebhook(w http.ResponseWriter, r *http.Request) {
	type UserRequest struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	decoder := json.NewDecoder(r.Body)
	userRequest := UserRequest{}
	err := decoder.Decode(&userRequest)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't read request body", err)
		return
	}

	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Could not read API Key", err)
		return
	}

	if apiKey != cfg.polkaKey {
		respondWithError(w, http.StatusUnauthorized, "API Keys do not match", err)
		return
	}

	if userRequest.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	err = cfg.db.UpdateUserRed(r.Context(), userRequest.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't update user to Red", err)
	}

	w.WriteHeader(http.StatusNoContent)

}
