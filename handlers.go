package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/wjseele/chirpy/internal/database"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		cfg.fileserverHits.Store(cfg.fileserverHits.Add(1))
		next.ServeHTTP(w, req)
	})
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(code)
	_, err := io.WriteString(w, msg)
	if err != nil {
		log.Printf("Error in the responder: %s", err)
		return
	}
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	dat, err := json.Marshal(payload)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("%s", err))
		return
	}
	w.Write(dat)
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, req *http.Request) {
	type chirpPost struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	type chirpResponse struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}

	decoder := json.NewDecoder(req.Body)
	post := chirpPost{}
	err := decoder.Decode(&post)
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("%s", err))
		return
	}

	if len(post.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	cleanedBody := badWordFilter(post.Body)
	newChirp := database.CreateChirpParams{
		Body:   cleanedBody,
		UserID: post.UserID,
	}
	response, err := cfg.dbQueries.CreateChirp(req.Context(), newChirp)
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("%s", err))
	}

	resp := chirpResponse{
		ID:        response.ID,
		CreatedAt: response.CreatedAt,
		UpdatedAt: response.UpdatedAt,
		Body:      response.Body,
		UserID:    response.UserID,
	}
	respondWithJSON(w, 201, resp)
}

func badWordFilter(s string) string {
	bodyWords := strings.Split(s, " ")
	for i := range bodyWords {
		switch strings.ToLower(bodyWords[i]) {
		case "kerfuffle", "sharbert", "fornax":
			bodyWords[i] = "****"
		}
	}
	cleanedBody := strings.Join(bodyWords, " ")
	return cleanedBody
}

func handlerHealthz(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err := io.WriteString(w, "OK")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func (cfg *apiConfig) handlerCounter(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	content := fmt.Sprintf(`
		<html>
		  <body>
		    <h1>Welcome, Chirpy Admin</h1>
		    <p>Chirpy has been visited %d times!</p>
		  </body>
		</html>
		`, cfg.fileserverHits.Load())
	w.WriteHeader(http.StatusOK)
	_, err := io.WriteString(w, content)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, req *http.Request) {
	if cfg.platform != "dev" {
		respondWithError(w, 403, "This only works on dev platforms")
	}
	cfg.fileserverHits.Store(0)
	err := cfg.dbQueries.ResetDB(req.Context())
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("%s", err))
	}
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, req *http.Request) {
	type emailPost struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(req.Body)
	post := emailPost{}
	err := decoder.Decode(&post)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("%s", err))
	}

	response, err := cfg.dbQueries.CreateUser(req.Context(), post.Email)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("%s", err))
	}
	jsonResponse := User{
		ID:        response.ID,
		CreatedAt: response.CreatedAt,
		UpdatedAt: response.UpdatedAt,
		Email:     response.Email,
	}
	respondWithJSON(w, 201, jsonResponse)
}
