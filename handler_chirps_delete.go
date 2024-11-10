package main

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/seantesterman/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerChirpsDelete(w http.ResponseWriter, r *http.Request) {
	bearerToken, err := auth.GetBearerToken(r.Header)
	fmt.Println(bearerToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Token not found", err)
		return
	}

	userID, err := auth.ValidateJWT(bearerToken, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusForbidden, "Couldn't validate JWT", err)
		return
	}

	vars := mux.Vars(r)
	chirpID, err := uuid.Parse(vars["chirpID"])
	if err != nil {
		respondWithError(w, http.StatusForbidden, "Invalid Chirp ID", err)
		return
	}

	chirp, err := cfg.db.GetChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chirp not found", err)
		return
	}

	if userID != chirp.UserID {
		respondWithError(w, http.StatusForbidden, "Incorrect author of Chirp", err)
		return
	}

	err = cfg.db.DeleteChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusForbidden, "Cannot delete Chirp", err)
	}

	w.WriteHeader(http.StatusNoContent)

}
