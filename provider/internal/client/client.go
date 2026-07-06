// Package client is a minimal HTTP client for the fakecloud API.
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	endpoint string
	http     *http.Client
}

func New(endpoint string) *Client {
	return &Client{endpoint: endpoint, http: &http.Client{}}
}

// APIError carries the server's status code and error message, so the
// provider can surface messages like "not X's turn" directly in the
// terraform apply output.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("fakecloud API error (%d): %s", e.StatusCode, e.Message)
}

func IsNotFound(err error) bool {
	apiErr, ok := err.(*APIError)
	return ok && apiErr.StatusCode == http.StatusNotFound
}

type Board struct {
	ID         int64      `json:"id"`
	Name       string     `json:"name"`
	Mode       string     `json:"mode"`
	Cells      []string   `json:"cells"`
	NextPlayer string     `json:"next_player"`
	Winner     string     `json:"winner"`
	Nameplate  *Nameplate `json:"nameplate"`
}

type Move struct {
	ID       int64  `json:"id"`
	BoardID  int64  `json:"board_id"`
	Player   string `json:"player"`
	Position int64  `json:"position"`
}

type Nameplate struct {
	ID      int64  `json:"id"`
	BoardID int64  `json:"board_id"`
	Text    string `json:"text"`
}

// do performs a request and decodes the JSON response into out (if non-nil).
func (c *Client) do(method, path string, body, out any) error {
	var buf *bytes.Buffer
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		buf = bytes.NewBuffer(data)
	} else {
		buf = &bytes.Buffer{}
	}

	req, err := http.NewRequest(method, c.endpoint+path, buf)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("cannot reach fakecloud at %s: %w (is the server running?)", c.endpoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var apiErr struct {
			Error string `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&apiErr)
		if apiErr.Error == "" {
			apiErr.Error = resp.Status
		}
		return &APIError{StatusCode: resp.StatusCode, Message: apiErr.Error}
	}

	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

func (c *Client) CreateBoard(name, mode string) (Board, error) {
	var board Board
	err := c.do("POST", "/tictactoe/boards", Board{Name: name, Mode: mode}, &board)
	return board, err
}

func (c *Client) GetBoard(id int64) (Board, error) {
	var board Board
	err := c.do("GET", fmt.Sprintf("/tictactoe/boards/%d", id), nil, &board)
	return board, err
}

func (c *Client) DeleteBoard(id int64) error {
	return c.do("DELETE", fmt.Sprintf("/tictactoe/boards/%d", id), nil, nil)
}

func (c *Client) CreateMove(boardID int64, player string, position int64) (Move, error) {
	var move Move
	err := c.do("POST", "/tictactoe/moves", Move{BoardID: boardID, Player: player, Position: position}, &move)
	return move, err
}

func (c *Client) GetMove(id int64) (Move, error) {
	var move Move
	err := c.do("GET", fmt.Sprintf("/tictactoe/moves/%d", id), nil, &move)
	return move, err
}

func (c *Client) DeleteMove(id int64) error {
	return c.do("DELETE", fmt.Sprintf("/tictactoe/moves/%d", id), nil, nil)
}

func (c *Client) CreateNameplate(boardID int64, text string) (Nameplate, error) {
	var plate Nameplate
	err := c.do("POST", "/tictactoe/nameplates", Nameplate{BoardID: boardID, Text: text}, &plate)
	return plate, err
}

func (c *Client) GetNameplate(id int64) (Nameplate, error) {
	var plate Nameplate
	err := c.do("GET", fmt.Sprintf("/tictactoe/nameplates/%d", id), nil, &plate)
	return plate, err
}

func (c *Client) UpdateNameplate(id int64, text string) (Nameplate, error) {
	var plate Nameplate
	err := c.do("PUT", fmt.Sprintf("/tictactoe/nameplates/%d", id), Nameplate{Text: text}, &plate)
	return plate, err
}

func (c *Client) DeleteNameplate(id int64) error {
	return c.do("DELETE", fmt.Sprintf("/tictactoe/nameplates/%d", id), nil, nil)
}
