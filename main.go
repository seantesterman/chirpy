package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/seantesterman/chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	secret         string
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}

	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("PLATFORM must be set")
	}

	secret := os.Getenv("SECRET")
	if secret == "" {
		log.Fatal("SECRET must be set")
	}

	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}
	dbQueries := database.New(dbConn)

	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
		platform:       platform,
		secret:         secret,
	}

	r := mux.NewRouter()
	r.Handle("/app/", http.StripPrefix("/app", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(filepathRoot)))))
	r.HandleFunc("/api/healthz", handlerReadiness).Methods("GET")
	r.HandleFunc("/api/users", apiCfg.handlerUsersCreate).Methods("POST")
	r.HandleFunc("/api/users", apiCfg.handlerUsersUpdate).Methods("PUT")
	r.HandleFunc("/api/chirps", apiCfg.handlerChirpsCreate).Methods("POST")
	r.HandleFunc("/api/chirps", apiCfg.handlerChirpsGet).Methods("GET")
	r.HandleFunc("/api/login", apiCfg.handlerLogin).Methods("POST")
	r.HandleFunc("/api/refresh", apiCfg.handlerRefreshToken).Methods("POST")
	r.HandleFunc("/api/revoke", apiCfg.handlerRevokeToken).Methods("POST")
	r.HandleFunc("/admin/metrics", apiCfg.handlerMetrics).Methods("GET")
	r.HandleFunc("/admin/reset", apiCfg.handlerReset).Methods("POST")
	r.HandleFunc("/api/chirps/{chirpID}", apiCfg.handlerChirpsID).Methods("GET")

	http.Handle("/", r)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())

}
