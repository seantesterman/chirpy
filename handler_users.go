package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/seantesterman/chirpy/internal/auth"
	"github.com/seantesterman/chirpy/internal/database"
)

type User struct {
	ID             uuid.UUID `json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Email          string    `json:"email"`
	HashedPassword string    `json:"hashed_password"`
}

func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	type UserResponse struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	HashedPW, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create password", err)
		return
	}

	user_params := database.CreateUserParams{
		Email: params.Email,
		HashedPassword: sql.NullString{
			String: HashedPW,
			Valid:  true,
		},
	}
	user, err := cfg.db.CreateUser(r.Context(), user_params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create user", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, UserResponse{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	})
}

func (cfg *apiConfig) handlerUsersUpdate(w http.ResponseWriter, r *http.Request) {
	type UserRequest struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	type UserResponse struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	bearerToken, err := auth.GetBearerToken(r.Header)
	fmt.Println(bearerToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Token not found", err)
		return
	}

	userID, err := auth.ValidateJWT(bearerToken, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	userRequest := UserRequest{}
	err = decoder.Decode(&userRequest)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't read request body", err)
		return
	}

	hashedPW, err := auth.HashPassword(userRequest.Password)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't update password", err)
		return
	}

	updateStruct := database.UpdateUserParams{
		Email: userRequest.Email,
		HashedPassword: sql.NullString{
			String: hashedPW,
			Valid:  true,
		},
		ID: userID,
	}
	user, err := cfg.db.UpdateUser(r.Context(), updateStruct)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't update email and password", err)
		return
	}

	user.UpdatedAt = time.Now().UTC()

	respondWithJSON(w, http.StatusOK, UserResponse{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     userRequest.Email,
	})

}
