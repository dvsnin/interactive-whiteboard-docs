package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func setupRouter() *mux.Router {
	r := mux.NewRouter()
	r.Handle("/board", auth(http.HandlerFunc(createBoard))).Methods(http.MethodPost)
	r.Handle("/board/{uid}", auth(http.HandlerFunc(getBoard))).Methods(http.MethodGet)
	r.Handle("/board/{uid}", auth(http.HandlerFunc(deleteBoard))).Methods(http.MethodDelete)
	return r
}

func newReq(method, url string, body []byte) *http.Request {
	req := httptest.NewRequest(method, url, bytes.NewReader(body))
	req.Header.Set("X-API-Key", apiKey)
	return req
}

func TestBoardAPI(t *testing.T) {
	boards = make(map[string]Board)
	router := setupRouter()

	// Создать доску
	body := []byte(`{"uid":"b1","title":"Первая доска"}`)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, newReq(http.MethodPost, "/board", body))
	if w.Code != http.StatusCreated {
		t.Fatalf("ожидали 201, получили %d", w.Code)
	}

	// Получить доску
	w = httptest.NewRecorder()
	router.ServeHTTP(w, newReq(http.MethodGet, "/board/b1", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("ожидали 200, получили %d", w.Code)
	}
	var b Board
	_ = json.NewDecoder(w.Body).Decode(&b)
	if b.Title != "Первая доска" {
		t.Errorf("ожидали title=Первая доска, получили %s", b.Title)
	}

	// Удалить доску
	w = httptest.NewRecorder()
	router.ServeHTTP(w, newReq(http.MethodDelete, "/board/b1", nil))
	if w.Code != http.StatusNoContent {
		t.Fatalf("ожидали 204, получили %d", w.Code)
	}

	// Повторное удаление -> 404
	w = httptest.NewRecorder()
	router.ServeHTTP(w, newReq(http.MethodDelete, "/board/b1", nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("ожидали 404, получили %d", w.Code)
	}
}
