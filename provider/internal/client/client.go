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

type VM struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	InstanceType string `json:"instance_type"`
}

type Game struct {
	ID         int64    `json:"id"`
	Name       string   `json:"name"`
	Board      []string `json:"board"`
	NextPlayer string   `json:"next_player"`
	Winner     string   `json:"winner"`
}

type Move struct {
	ID       int64  `json:"id"`
	GameID   int64  `json:"game_id"`
	Player   string `json:"player"`
	Position int64  `json:"position"`
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

func (c *Client) CreateVM(name, instanceType string) (VM, error) {
	var vm VM
	err := c.do("POST", "/vms", VM{Name: name, InstanceType: instanceType}, &vm)
	return vm, err
}

func (c *Client) GetVM(id int64) (VM, error) {
	var vm VM
	err := c.do("GET", fmt.Sprintf("/vms/%d", id), nil, &vm)
	return vm, err
}

func (c *Client) UpdateVM(id int64, name, instanceType string) (VM, error) {
	var vm VM
	err := c.do("PUT", fmt.Sprintf("/vms/%d", id), VM{Name: name, InstanceType: instanceType}, &vm)
	return vm, err
}

func (c *Client) DeleteVM(id int64) error {
	return c.do("DELETE", fmt.Sprintf("/vms/%d", id), nil, nil)
}

func (c *Client) CreateGame(name string) (Game, error) {
	var game Game
	err := c.do("POST", "/tictactoe/games", Game{Name: name}, &game)
	return game, err
}

func (c *Client) GetGame(id int64) (Game, error) {
	var game Game
	err := c.do("GET", fmt.Sprintf("/tictactoe/games/%d", id), nil, &game)
	return game, err
}

func (c *Client) DeleteGame(id int64) error {
	return c.do("DELETE", fmt.Sprintf("/tictactoe/games/%d", id), nil, nil)
}

func (c *Client) CreateMove(gameID int64, player string, position int64) (Move, error) {
	var move Move
	err := c.do("POST", "/tictactoe/moves", Move{GameID: gameID, Player: player, Position: position}, &move)
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
