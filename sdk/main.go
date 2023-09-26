package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"fakecloud/handlers/vm"
)

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

func (c *Client) CreateVM(vm *vm.VirtualMachine) error {
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

// other CRUD functions...
