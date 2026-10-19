package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var mu = &sync.Mutex{}
var bc Blockchain

func main() {
	bc = NewBlockchain()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", getBlockchain)
	mux.HandleFunc("POST /addblock", writeBlock)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Println("Starting API server on :8080")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server dynamic crash: %v", err)
		}
	}()

	// Listening loop for OS graceful interrupts
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Initiating graceful teardown components...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP shutdown error: %v", err)
	}
	log.Println("Server processing stopped. Exiting.")
}

func getBlockchain(w http.ResponseWriter, r *http.Request) {
	log.Printf("Successfully retrieved %d blocks", len(bc))
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(bc); err != nil {
		log.Printf("JSON encoding erro: %v", err)
	}
}

func writeBlock(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Value int `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("addblock parse error: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mu.Lock()
	prevBlock := bc[len(bc)-1]
	newBlock := generateBlock(prevBlock, req.Value)
	if isBlockValid(newBlock, prevBlock) {
		bc = append(bc, newBlock)
	}
	mu.Unlock()

	log.Printf("Block created with hash: %s", newBlock.Hash)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(newBlock); err != nil {
		log.Printf("JSON encoding error: %v", err)
	}
}
