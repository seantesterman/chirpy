package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/seantesterman/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password     string `json:"password"`
		Email        string `json:"email"`
		ExpiresInSec int    `json:"expires_in_seconds"`
	}

	type UserResponse struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
		Token     string    `json:"token"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't process login", err)
		return
	}
	user, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	if user.HashedPassword.Valid {
		hashedPasswordStr := user.HashedPassword.String
		err = auth.CheckPasswordHash(params.Password, hashedPasswordStr)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Incorrect password", err)
			return
		}
	} else {
		respondWithError(w, http.StatusUnauthorized, "Invalid user or password", err)
		return
	}

	if params.ExpiresInSec > 3600 {
		params.ExpiresInSec = 3600
	}

	if params.ExpiresInSec == 0 {
		params.ExpiresInSec = 3600
	}

	token, err := auth.MakeJWT(user.ID, cfg.secret, time.Duration(params.ExpiresInSec)*time.Second)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Cannot make token", err)
		return
	}

	if err == nil {
		respondWithJSON(w, http.StatusOK, UserResponse{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
			Token:     token,
		})
	} else {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}
}
