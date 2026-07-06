// Package store holds all fakecloud resources in memory and implements the
// rules of tic-tac-toe. Everything is guarded by a single mutex; this is a
// learning playground, not a real cloud.
package store

import (
	"errors"
	"fmt"
	"sort"
	"sync"
)

var ErrNotFound = errors.New("not found")

// ConflictError is returned when a request is well-formed but not allowed by
// the current state (cell taken, wrong turn, game over).
type ConflictError struct{ Reason string }

func (e ConflictError) Error() string { return e.Reason }

// ValidationError is returned when a request is malformed.
type ValidationError struct{ Reason string }

func (e ValidationError) Error() string { return e.Reason }

type VM struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	InstanceType string `json:"instance_type"`
}

type Move struct {
	ID       int    `json:"id"`
	GameID   int    `json:"game_id"`
	Player   string `json:"player"`
	Position int    `json:"position"`
}

type Game struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	// Board has 9 cells, row by row; each is "", "X", or "O". Board,
	// NextPlayer, Winner, and Moves are derived from the game's moves.
	Board      []string `json:"board"`
	NextPlayer string   `json:"next_player"`
	Winner     string   `json:"winner"`
	Moves      []Move   `json:"moves"`
}

type Store struct {
	mu         sync.Mutex
	nextVMID   int
	nextGameID int
	nextMoveID int
	vms        map[int]VM
	games      map[int]string // id -> name
	moves      map[int]Move
}

func New() *Store {
	return &Store{
		nextVMID:   1,
		nextGameID: 1,
		nextMoveID: 1,
		vms:        map[int]VM{},
		games:      map[int]string{},
		moves:      map[int]Move{},
	}
}

// --- VMs ---

func (s *Store) CreateVM(name, instanceType string) (VM, error) {
	if name == "" {
		return VM{}, ValidationError{"name is required"}
	}
	if instanceType == "" {
		return VM{}, ValidationError{"instance_type is required"}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	vm := VM{ID: s.nextVMID, Name: name, InstanceType: instanceType}
	s.nextVMID++
	s.vms[vm.ID] = vm
	return vm, nil
}

func (s *Store) ListVMs() []VM {
	s.mu.Lock()
	defer s.mu.Unlock()
	vms := make([]VM, 0, len(s.vms))
	for _, vm := range s.vms {
		vms = append(vms, vm)
	}
	sort.Slice(vms, func(i, j int) bool { return vms[i].ID < vms[j].ID })
	return vms
}

func (s *Store) GetVM(id int) (VM, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	vm, ok := s.vms[id]
	if !ok {
		return VM{}, ErrNotFound
	}
	return vm, nil
}

func (s *Store) UpdateVM(id int, name, instanceType string) (VM, error) {
	if name == "" {
		return VM{}, ValidationError{"name is required"}
	}
	if instanceType == "" {
		return VM{}, ValidationError{"instance_type is required"}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	vm, ok := s.vms[id]
	if !ok {
		return VM{}, ErrNotFound
	}
	vm.Name = name
	vm.InstanceType = instanceType
	s.vms[id] = vm
	return vm, nil
}

func (s *Store) DeleteVM(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.vms[id]; !ok {
		return ErrNotFound
	}
	delete(s.vms, id)
	return nil
}

// --- Tic-tac-toe ---

var winningLines = [8][3]int{
	{0, 1, 2}, {3, 4, 5}, {6, 7, 8},
	{0, 3, 6}, {1, 4, 7}, {2, 5, 8},
	{0, 4, 8}, {2, 4, 6},
}

// gameState derives board, next player, winner, and ordered moves.
// Caller must hold s.mu.
func (s *Store) gameState(gameID int) Game {
	game := Game{ID: gameID, Name: s.games[gameID], Board: make([]string, 9), Moves: []Move{}}
	for _, move := range s.moves {
		if move.GameID == gameID {
			game.Board[move.Position] = move.Player
			game.Moves = append(game.Moves, move)
		}
	}
	sort.Slice(game.Moves, func(i, j int) bool { return game.Moves[i].ID < game.Moves[j].ID })

	for _, line := range winningLines {
		if game.Board[line[0]] != "" && game.Board[line[0]] == game.Board[line[1]] && game.Board[line[1]] == game.Board[line[2]] {
			game.Winner = game.Board[line[0]]
		}
	}
	if game.Winner == "" && len(game.Moves) == 9 {
		game.Winner = "draw"
	}
	if game.Winner == "" {
		// X always starts; whoever has fewer marks on the board goes next.
		countX := 0
		for _, cell := range game.Board {
			if cell == "X" {
				countX++
			}
		}
		if countX == len(game.Moves)-countX {
			game.NextPlayer = "X"
		} else {
			game.NextPlayer = "O"
		}
	}
	return game
}

func (s *Store) CreateGame(name string) (Game, error) {
	if name == "" {
		return Game{}, ValidationError{"name is required"}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	id := s.nextGameID
	s.nextGameID++
	s.games[id] = name
	return s.gameState(id), nil
}

func (s *Store) ListGames() []Game {
	s.mu.Lock()
	defer s.mu.Unlock()
	ids := make([]int, 0, len(s.games))
	for id := range s.games {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	games := make([]Game, 0, len(ids))
	for _, id := range ids {
		games = append(games, s.gameState(id))
	}
	return games
}

func (s *Store) GetGame(id int) (Game, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.games[id]; !ok {
		return Game{}, ErrNotFound
	}
	return s.gameState(id), nil
}

func (s *Store) DeleteGame(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.games[id]; !ok {
		return ErrNotFound
	}
	delete(s.games, id)
	for moveID, move := range s.moves {
		if move.GameID == id {
			delete(s.moves, moveID)
		}
	}
	return nil
}

func (s *Store) CreateMove(gameID int, player string, position int) (Move, error) {
	if player != "X" && player != "O" {
		return Move{}, ValidationError{`player must be "X" or "O"`}
	}
	if position < 0 || position > 8 {
		return Move{}, ValidationError{"position must be between 0 and 8"}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.games[gameID]; !ok {
		return Move{}, ErrNotFound
	}
	game := s.gameState(gameID)
	if game.Winner != "" {
		return Move{}, ConflictError{"game is already over"}
	}
	if game.Board[position] != "" {
		return Move{}, ConflictError{fmt.Sprintf("position %d is already taken by %s", position, game.Board[position])}
	}
	if player != game.NextPlayer {
		return Move{}, ConflictError{fmt.Sprintf("not %s's turn, it is %s's turn", player, game.NextPlayer)}
	}
	move := Move{ID: s.nextMoveID, GameID: gameID, Player: player, Position: position}
	s.nextMoveID++
	s.moves[move.ID] = move
	return move, nil
}

func (s *Store) GetMove(id int) (Move, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	move, ok := s.moves[id]
	if !ok {
		return Move{}, ErrNotFound
	}
	return move, nil
}

func (s *Store) DeleteMove(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.moves[id]; !ok {
		return ErrNotFound
	}
	delete(s.moves, id)
	return nil
}
