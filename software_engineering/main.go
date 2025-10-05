package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
)

type Board struct {
	UID   string `json:"uid"`
	Title string `json:"title"`
}

var (
	boards = make(map[string]Board)
	mu     sync.Mutex
	apiKey = "secret123"
)

func auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != apiKey {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// POST /board — создать доску
func createBoard(w http.ResponseWriter, r *http.Request) {
	var b Board
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if b.UID == "" || b.Title == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	mu.Lock()
	boards[b.UID] = b
	mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(b)
}

// GET /board/{uid} — получить инфу о доске
func getBoard(w http.ResponseWriter, r *http.Request) {
	uid := mux.Vars(r)["uid"]

	if uid == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	mu.Lock()
	b, ok := boards[uid]
	mu.Unlock()

	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(b)
}

// DELETE /board/{uid} — удалить доску
func deleteBoard(w http.ResponseWriter, r *http.Request) {
	uid := mux.Vars(r)["uid"]

	if uid == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	mu.Lock()
	_, ok := boards[uid]
	if ok {
		delete(boards, uid)
	}
	mu.Unlock()

	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func main() {
	r := mux.NewRouter()

	r.Handle("/board", auth(http.HandlerFunc(createBoard))).Methods(http.MethodPost)
	r.Handle("/board/{uid}", auth(http.HandlerFunc(getBoard))).Methods(http.MethodGet)
	r.Handle("/board/{uid}", auth(http.HandlerFunc(deleteBoard))).Methods(http.MethodDelete)

	log.Println("server running at :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
