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
	mux.HandleFunc("POST /vms", a.createVM)
	mux.HandleFunc("GET /vms", a.listVMs)
	mux.HandleFunc("GET /vms/{id}", a.getVM)
	mux.HandleFunc("PUT /vms/{id}", a.updateVM)
	mux.HandleFunc("DELETE /vms/{id}", a.deleteVM)

	mux.HandleFunc("POST /tictactoe/games", a.createGame)
	mux.HandleFunc("GET /tictactoe/games", a.listGames)
	mux.HandleFunc("GET /tictactoe/games/{id}", a.getGame)
	mux.HandleFunc("DELETE /tictactoe/games/{id}", a.deleteGame)

	mux.HandleFunc("POST /tictactoe/moves", a.createMove)
	mux.HandleFunc("GET /tictactoe/moves/{id}", a.getMove)
	mux.HandleFunc("DELETE /tictactoe/moves/{id}", a.deleteMove)
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

// --- VMs ---

func (a *API) createVM(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.Path)
	var req store.VM
	if !decode(w, r, &req) {
		return
	}
	vm, err := a.store.CreateVM(req.Name, req.InstanceType)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, vm)
}

func (a *API) listVMs(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.store.ListVMs())
}

func (a *API) getVM(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, store.ErrNotFound)
		return
	}
	vm, err := a.store.GetVM(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, vm)
}

func (a *API) updateVM(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.Path)
	id, err := pathID(r)
	if err != nil {
		writeError(w, store.ErrNotFound)
		return
	}
	var req store.VM
	if !decode(w, r, &req) {
		return
	}
	vm, err := a.store.UpdateVM(id, req.Name, req.InstanceType)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, vm)
}

func (a *API) deleteVM(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.Path)
	id, err := pathID(r)
	if err != nil {
		writeError(w, store.ErrNotFound)
		return
	}
	if err := a.store.DeleteVM(id); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- Tic-tac-toe ---

func (a *API) createGame(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.Path)
	var req store.Game
	if !decode(w, r, &req) {
		return
	}
	game, err := a.store.CreateGame(req.Name)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, game)
}

func (a *API) listGames(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.store.ListGames())
}

func (a *API) getGame(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, store.ErrNotFound)
		return
	}
	game, err := a.store.GetGame(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, game)
}

func (a *API) deleteGame(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.Path)
	id, err := pathID(r)
	if err != nil {
		writeError(w, store.ErrNotFound)
		return
	}
	if err := a.store.DeleteGame(id); err != nil {
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
	move, err := a.store.CreateMove(req.GameID, req.Player, req.Position)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, move)
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
