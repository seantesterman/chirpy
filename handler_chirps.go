package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/seantesterman/chirpy/internal/auth"
	"github.com/seantesterman/chirpy/internal/database"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerChirpsCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Could not get Token", err)
		return
	}

	id, err := auth.ValidateJWT(tokenString, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unable to validate token", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	cleaned, err := validateChirp(params.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleaned,
		UserID: id,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}

func validateChirp(body string) (string, error) {
	const maxChirpLength = 140
	if len(body) > maxChirpLength {
		return "", errors.New("Chirp is too long")
	}

	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	cleaned := getCleanedBody(body, badWords)
	return cleaned, nil
}

func getCleanedBody(body string, badWords map[string]struct{}) string {
	words := strings.Split(body, " ")
	for i, word := range words {
		loweredWord := strings.ToLower(word)
		if _, ok := badWords[loweredWord]; ok {
			words[i] = "****"
		}
	}
	cleaned := strings.Join(words, " ")
	return cleaned
}

func (cfg *apiConfig) handlerChirpsGet(w http.ResponseWriter, r *http.Request) {
	authorID := r.URL.Query().Get("author_id")
	chirpsSort := r.URL.Query().Get("sort")

	if authorID == "" {

		if chirpsSort == "" || chirpsSort == "asc" {
			chirps, err := cfg.db.GetChirps(r.Context())
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Couldn't get chirps", err)
				return
			}
			listOfChirps := []Chirp{}
			for _, chi := range chirps {
				convertedChirp := Chirp{
					ID:        chi.ID,
					CreatedAt: chi.CreatedAt,
					UpdatedAt: chi.UpdatedAt,
					Body:      chi.Body,
					UserID:    chi.UserID,
				}
				listOfChirps = append(listOfChirps, convertedChirp)
			}
			respondWithJSON(w, http.StatusOK, listOfChirps)
			return
		}

		chirps, err := cfg.db.GetChirpsDesc(r.Context())
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't get chirps", err)
			return
		}
		listOfChirps := []Chirp{}
		for _, chi := range chirps {
			convertedChirp := Chirp{
				ID:        chi.ID,
				CreatedAt: chi.CreatedAt,
				UpdatedAt: chi.UpdatedAt,
				Body:      chi.Body,
				UserID:    chi.UserID,
			}
			listOfChirps = append(listOfChirps, convertedChirp)
		}
		respondWithJSON(w, http.StatusOK, listOfChirps)
		return
	}

	parsedAuthorID, err := uuid.Parse(authorID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't parse author ID", err)
		return
	}

	if chirpsSort == "" || chirpsSort == "asc" {
		chirps, err := cfg.db.GetChirpsByUser(r.Context(), parsedAuthorID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't get chirps", err)
			return
		}

		listOfChirps := []Chirp{}
		for _, chi := range chirps {
			convertedChirp := Chirp{
				ID:        chi.ID,
				CreatedAt: chi.CreatedAt,
				UpdatedAt: chi.UpdatedAt,
				Body:      chi.Body,
				UserID:    chi.UserID,
			}
			listOfChirps = append(listOfChirps, convertedChirp)
		}

		respondWithJSON(w, http.StatusOK, listOfChirps)
		return
	}
	chirps, err := cfg.db.GetChirpsByUserDesc(r.Context(), parsedAuthorID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get chirps", err)
		return
	}

	listOfChirps := []Chirp{}
	for _, chi := range chirps {
		convertedChirp := Chirp{
			ID:        chi.ID,
			CreatedAt: chi.CreatedAt,
			UpdatedAt: chi.UpdatedAt,
			Body:      chi.Body,
			UserID:    chi.UserID,
		}
		listOfChirps = append(listOfChirps, convertedChirp)
	}

	respondWithJSON(w, http.StatusOK, listOfChirps)
	return

}

func (cfg *apiConfig) handlerChirpsID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	chirpID, err := uuid.Parse(vars["chirpID"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Chirp ID", err)
	}
	chirp, err := cfg.db.GetChirp(r.Context(), chirpID)

	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chirp not found", err)
		return
	}

	respondWithJSON(w, http.StatusOK, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}
