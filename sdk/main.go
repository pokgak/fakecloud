package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type VirtualMachine struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	InstanceType string `json:"instance_type"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string, username string, password string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

func (c *Client) CreateVM(vm *VirtualMachine) error {
	body, err := json.Marshal(vm)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.baseURL+"/vms", bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) GetVMs() ([]VirtualMachine, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/vms", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var vms []VirtualMachine
	if err := json.NewDecoder(resp.Body).Decode(&vms); err != nil {
		return nil, err
	}

	return vms, nil
}

// other CRUD functions...
