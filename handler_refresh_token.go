package main

import (
	"net/http"
	"time"

	"github.com/seantesterman/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	type UserResponse struct {
		Token string `json:"token"`
	}

	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Token not found", err)
		return
	}

	refreshToken, err := cfg.db.GetToken(r.Context(), bearerToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Token not found", err)
		return
	}

	if refreshToken.ExpiresAt.Before(time.Now().UTC()) {
		respondWithError(w, http.StatusUnauthorized, "Token expired", err)
		return
	}

	if refreshToken.RevokedAt.Valid == true {
		respondWithError(w, http.StatusUnauthorized, "Token revoked", err)
		return
	}

	token, err := auth.MakeJWT(refreshToken.UserID, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Cannot create new token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, UserResponse{
		Token: token,
	})

}
