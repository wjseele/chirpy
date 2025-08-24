package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	serveMux := http.NewServeMux()
	server := http.Server{
		Addr:    ":8080",
		Handler: serveMux,
	}
	apiCfg := apiConfig{}

	serveMux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	serveMux.HandleFunc("/healthz", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, err := io.WriteString(w, "OK")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	})

	serveMux.HandleFunc("/metrics", apiCfg.handlerCounter)
	serveMux.HandleFunc("/reset", apiCfg.handlerReset)

	err := server.ListenAndServe()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	cfg.fileserverHits.Add(1)
	return next
}

func (cfg *apiConfig) handlerCounter(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err := io.WriteString(w, fmt.Sprintf("%v", cfg.fileserverHits.Load()))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, req *http.Request) {
	cfg.fileserverHits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err := io.WriteString(w, fmt.Sprintf("Reset counter to %v", cfg.fileserverHits.Load()))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
