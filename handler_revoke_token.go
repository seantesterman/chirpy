package main

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/seantesterman/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerRevokeToken(w http.ResponseWriter, r *http.Request) {
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

	refreshToken.RevokedAt = sql.NullTime{
		Time:  time.Now().UTC(),
		Valid: true,
	}

	refreshToken.UpdatedAt = time.Now().UTC()

	err = cfg.db.RevokeToken(r.Context(), refreshToken.Token)

	w.WriteHeader(http.StatusNoContent)
}
