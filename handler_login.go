package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/seantesterman/chirpy/internal/auth"
	"github.com/seantesterman/chirpy/internal/database"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	type UserResponse struct {
		ID           uuid.UUID `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
		IsChirpyRed  bool      `json:"is_chirpy_red"`
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

	token, err := auth.MakeJWT(user.ID, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Cannot make token", err)
		return
	}

	refreshTokenString, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Cannot create refresh token", err)
	}

	arg := database.CreateRefreshTokenParams{
		Token:  refreshTokenString,
		UserID: user.ID,
	}

	refreshToken, err := cfg.db.CreateRefreshToken(r.Context(), arg)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Cannot make refresh token", err)
	}

	refreshToken.RevokedAt = sql.NullTime{Time: time.Time{}, Valid: false}
	refreshToken.ExpiresAt = time.Now().UTC().Add(60 * 24 * time.Hour)

	if err == nil {
		respondWithJSON(w, http.StatusOK, UserResponse{
			ID:           user.ID,
			CreatedAt:    user.CreatedAt,
			UpdatedAt:    user.UpdatedAt,
			Email:        user.Email,
			Token:        token,
			RefreshToken: refreshTokenString,
			IsChirpyRed:  user.IsChirpyRed.Bool,
		})
	} else {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}
}
