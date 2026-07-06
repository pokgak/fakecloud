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

// Board modes.
//
// In freeplay anyone can mark any empty cell at any time — this is the mode
// the Terraform lessons use, since e.g. a count over five cells creates five
// moves in one apply. In duel mode the server referees a real game: X starts,
// turns alternate, and the board locks once someone wins.
const (
	ModeFreeplay = "freeplay"
	ModeDuel     = "duel"
)

type Move struct {
	ID       int    `json:"id"`
	BoardID  int    `json:"board_id"`
	Player   string `json:"player"`
	Position int    `json:"position"`
}

// Nameplate is a plaque attached to a board. Unlike moves it is updatable
// in place, which makes it the demo object for in-place updates and for the
// "two states fighting over one resource" lesson.
type Nameplate struct {
	ID      int    `json:"id"`
	BoardID int    `json:"board_id"`
	Text    string `json:"text"`
}

type Board struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Mode string `json:"mode"`
	// Cells has 9 entries, row by row; each is "", "X", or "O". Cells,
	// NextPlayer, Winner, and Moves are derived from the board's moves.
	Cells      []string   `json:"cells"`
	NextPlayer string     `json:"next_player"`
	Winner     string     `json:"winner"`
	Moves      []Move     `json:"moves"`
	Nameplate  *Nameplate `json:"nameplate,omitempty"`
}

type boardMeta struct {
	name string
	mode string
}

type Store struct {
	mu              sync.Mutex
	nextBoardID     int
	nextMoveID      int
	nextNameplateID int
	boards          map[int]boardMeta
	moves           map[int]Move
	nameplates      map[int]Nameplate
}

func New() *Store {
	return &Store{
		nextBoardID:     1,
		nextMoveID:      1,
		nextNameplateID: 1,
		boards:          map[int]boardMeta{},
		moves:           map[int]Move{},
		nameplates:      map[int]Nameplate{},
	}
}

var winningLines = [8][3]int{
	{0, 1, 2}, {3, 4, 5}, {6, 7, 8},
	{0, 3, 6}, {1, 4, 7}, {2, 5, 8},
	{0, 4, 8}, {2, 4, 6},
}

// boardState derives cells, next player, winner, and ordered moves.
// Caller must hold s.mu.
func (s *Store) boardState(boardID int) Board {
	meta := s.boards[boardID]
	board := Board{ID: boardID, Name: meta.name, Mode: meta.mode, Cells: make([]string, 9), Moves: []Move{}}
	for _, move := range s.moves {
		if move.BoardID == boardID {
			board.Cells[move.Position] = move.Player
			board.Moves = append(board.Moves, move)
		}
	}
	sort.Slice(board.Moves, func(i, j int) bool { return board.Moves[i].ID < board.Moves[j].ID })

	for _, plate := range s.nameplates {
		if plate.BoardID == boardID {
			p := plate
			board.Nameplate = &p
			break
		}
	}

	for _, line := range winningLines {
		if board.Cells[line[0]] != "" && board.Cells[line[0]] == board.Cells[line[1]] && board.Cells[line[1]] == board.Cells[line[2]] {
			board.Winner = board.Cells[line[0]]
			break
		}
	}
	if board.Winner == "" && len(board.Moves) == 9 {
		board.Winner = "draw"
	}

	// Turn order only means something in a refereed duel.
	if meta.mode == ModeDuel && board.Winner == "" {
		// X always starts; whoever has fewer marks on the board goes next.
		countX := 0
		for _, cell := range board.Cells {
			if cell == "X" {
				countX++
			}
		}
		if countX == len(board.Moves)-countX {
			board.NextPlayer = "X"
		} else {
			board.NextPlayer = "O"
		}
	}
	return board
}

func (s *Store) CreateBoard(name, mode string) (Board, error) {
	if name == "" {
		return Board{}, ValidationError{"name is required"}
	}
	if mode == "" {
		mode = ModeFreeplay
	}
	if mode != ModeFreeplay && mode != ModeDuel {
		return Board{}, ValidationError{`mode must be "freeplay" or "duel"`}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	id := s.nextBoardID
	s.nextBoardID++
	s.boards[id] = boardMeta{name: name, mode: mode}
	return s.boardState(id), nil
}

func (s *Store) ListBoards() []Board {
	s.mu.Lock()
	defer s.mu.Unlock()
	ids := make([]int, 0, len(s.boards))
	for id := range s.boards {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	boards := make([]Board, 0, len(ids))
	for _, id := range ids {
		boards = append(boards, s.boardState(id))
	}
	return boards
}

func (s *Store) GetBoard(id int) (Board, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.boards[id]; !ok {
		return Board{}, ErrNotFound
	}
	return s.boardState(id), nil
}

func (s *Store) DeleteBoard(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.boards[id]; !ok {
		return ErrNotFound
	}
	delete(s.boards, id)
	for moveID, move := range s.moves {
		if move.BoardID == id {
			delete(s.moves, moveID)
		}
	}
	for plateID, plate := range s.nameplates {
		if plate.BoardID == id {
			delete(s.nameplates, plateID)
		}
	}
	return nil
}

func (s *Store) CreateMove(boardID int, player string, position int) (Move, error) {
	if player != "X" && player != "O" {
		return Move{}, ValidationError{`player must be "X" or "O"`}
	}
	if position < 0 || position > 8 {
		return Move{}, ValidationError{"position must be between 0 and 8"}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	meta, ok := s.boards[boardID]
	if !ok {
		return Move{}, ErrNotFound
	}
	board := s.boardState(boardID)
	if board.Cells[position] != "" {
		return Move{}, ConflictError{fmt.Sprintf("position %d is already taken by %s", position, board.Cells[position])}
	}
	if meta.mode == ModeDuel {
		if board.Winner != "" {
			return Move{}, ConflictError{"game is already over"}
		}
		if player != board.NextPlayer {
			return Move{}, ConflictError{fmt.Sprintf("not %s's turn, it is %s's turn", player, board.NextPlayer)}
		}
	}
	move := Move{ID: s.nextMoveID, BoardID: boardID, Player: player, Position: position}
	s.nextMoveID++
	s.moves[move.ID] = move
	return move, nil
}

func (s *Store) CreateNameplate(boardID int, text string) (Nameplate, error) {
	if text == "" {
		return Nameplate{}, ValidationError{"text is required"}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.boards[boardID]; !ok {
		return Nameplate{}, ErrNotFound
	}
	for _, plate := range s.nameplates {
		if plate.BoardID == boardID {
			return Nameplate{}, ConflictError{fmt.Sprintf("board %d already has a nameplate (id=%d) — import it instead of creating another", boardID, plate.ID)}
		}
	}
	plate := Nameplate{ID: s.nextNameplateID, BoardID: boardID, Text: text}
	s.nextNameplateID++
	s.nameplates[plate.ID] = plate
	return plate, nil
}

func (s *Store) GetNameplate(id int) (Nameplate, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	plate, ok := s.nameplates[id]
	if !ok {
		return Nameplate{}, ErrNotFound
	}
	return plate, nil
}

func (s *Store) UpdateNameplate(id int, text string) (Nameplate, error) {
	if text == "" {
		return Nameplate{}, ValidationError{"text is required"}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	plate, ok := s.nameplates[id]
	if !ok {
		return Nameplate{}, ErrNotFound
	}
	plate.Text = text
	s.nameplates[id] = plate
	return plate, nil
}

func (s *Store) DeleteNameplate(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.nameplates[id]; !ok {
		return ErrNotFound
	}
	delete(s.nameplates, id)
	return nil
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
