package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var bc Blockchain

func main() {
	bc = NewBlockchain()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", getBlockchain)
	mux.HandleFunc("POST /addblock", writeBlock)
	mux.HandleFunc("POST /balance", getBalance)

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
	log.Printf("Successfully retrieved %d blocks", len(bc.blocks))
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(bc.blocks); err != nil {
		log.Printf("JSON encoding erro: %v", err)
	}
}

func writeBlock(w http.ResponseWriter, r *http.Request) {
	var req struct {
		From, To string
		Value    int
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("addblock parse error: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	prevBlock := bc.blocks[len(bc.blocks)-1]
	newBlock, err := generateBlock(prevBlock, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if isBlockValid(newBlock, prevBlock) {
		bc.AddBlock(newBlock)
	}

	log.Printf("Block created with hash: %s", newBlock.Hash)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(newBlock); err != nil {
		log.Printf("JSON encoding error: %v", err)
	}
}

func getBalance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		Address string
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("balance parse error: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	balance := 0
	UTXOs := bc.FindUTXO(req.Address)

	for _, out := range UTXOs {
		balance += out.Value
	}

	log.Printf("Balance %d", balance)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(balance); err != nil {
		log.Printf("JSON encoding error: %v", err)
	}
}
