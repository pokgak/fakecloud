// Package api exposes the fakecloud store as a JSON HTTP API.
package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/pokgak/fakecloud/server/store"
)

type API struct {
	store *store.Store
}

func New(s *store.Store) *API {
	return &API{store: s}
}

func (a *API) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /tictactoe/boards", a.createBoard)
	mux.HandleFunc("GET /tictactoe/boards", a.listBoards)
	mux.HandleFunc("GET /tictactoe/boards/{id}", a.getBoard)
	mux.HandleFunc("DELETE /tictactoe/boards/{id}", a.deleteBoard)

	mux.HandleFunc("POST /tictactoe/moves", a.createMove)
	mux.HandleFunc("GET /tictactoe/moves/{id}", a.getMove)
	mux.HandleFunc("DELETE /tictactoe/moves/{id}", a.deleteMove)

	mux.HandleFunc("POST /tictactoe/nameplates", a.createNameplate)
	mux.HandleFunc("GET /tictactoe/nameplates/{id}", a.getNameplate)
	mux.HandleFunc("PUT /tictactoe/nameplates/{id}", a.updateNameplate)
	mux.HandleFunc("DELETE /tictactoe/nameplates/{id}", a.deleteNameplate)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// writeError maps store errors to HTTP status codes and a JSON error body.
func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	var conflict store.ConflictError
	var invalid store.ValidationError
	switch {
	case errors.Is(err, store.ErrNotFound):
		status = http.StatusNotFound
	case errors.As(err, &conflict):
		status = http.StatusConflict
	case errors.As(err, &invalid):
		status = http.StatusBadRequest
	}
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func pathID(r *http.Request) (int, error) {
	return strconv.Atoi(r.PathValue("id"))
}

func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON: " + err.Error()})
		return false
	}
	return true
}

func (a *API) createBoard(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.Path)
	var req store.Board
	if !decode(w, r, &req) {
		return
	}
	board, err := a.store.CreateBoard(req.Name, req.Mode)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, board)
}

func (a *API) listBoards(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.store.ListBoards())
}

func (a *API) getBoard(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, store.ErrNotFound)
		return
	}
	board, err := a.store.GetBoard(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, board)
}

func (a *API) deleteBoard(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.Path)
	id, err := pathID(r)
	if err != nil {
		writeError(w, store.ErrNotFound)
		return
	}
	if err := a.store.DeleteBoard(id); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (a *API) createMove(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.Path)
	var req store.Move
	if !decode(w, r, &req) {
		return
	}
	move, err := a.store.CreateMove(req.BoardID, req.Player, req.Position)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, move)
}

func (a *API) createNameplate(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.Path)
	var req store.Nameplate
	if !decode(w, r, &req) {
		return
	}
	plate, err := a.store.CreateNameplate(req.BoardID, req.Text)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, plate)
}

func (a *API) getNameplate(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, store.ErrNotFound)
		return
	}
	plate, err := a.store.GetNameplate(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, plate)
}

func (a *API) updateNameplate(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.Path)
	id, err := pathID(r)
	if err != nil {
		writeError(w, store.ErrNotFound)
		return
	}
	var req store.Nameplate
	if !decode(w, r, &req) {
		return
	}
	plate, err := a.store.UpdateNameplate(id, req.Text)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, plate)
}

func (a *API) deleteNameplate(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.Path)
	id, err := pathID(r)
	if err != nil {
		writeError(w, store.ErrNotFound)
		return
	}
	if err := a.store.DeleteNameplate(id); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (a *API) getMove(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, store.ErrNotFound)
		return
	}
	move, err := a.store.GetMove(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, move)
}

func (a *API) deleteMove(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.Path)
	id, err := pathID(r)
	if err != nil {
		writeError(w, store.ErrNotFound)
		return
	}
	if err := a.store.DeleteMove(id); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
