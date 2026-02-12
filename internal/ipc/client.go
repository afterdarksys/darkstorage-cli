package ipc

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

type Client struct {
	socketPath string
	timeout    time.Duration
}

func NewClient(socketPath string) *Client {
	return &Client{
		socketPath: socketPath,
		timeout:    5 * time.Second,
	}
}

func (c *Client) SendCommand(cmd *Command) (*Response, error) {
	conn, err := net.DialTimeout("unix", c.socketPath, c.timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(c.timeout))

	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(cmd); err != nil {
		return nil, fmt.Errorf("failed to send command: %w", err)
	}

	var response Response
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to receive response: %w", err)
	}

	return &response, nil
}

func (c *Client) GetStatus() (*StatusResponse, error) {
	cmd := &Command{Type: "status"}
	resp, err := c.SendCommand(cmd)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("command failed: %s", resp.Error)
	}

	var status StatusResponse
	if err := json.Unmarshal(resp.Data, &status); err != nil {
		return nil, err
	}

	return &status, nil
}

func (c *Client) AddSyncFolder(req *AddSyncFolderRequest) (*AddSyncFolderResponse, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	cmd := &Command{
		Type: "add_sync_folder",
		Data: data,
	}

	resp, err := c.SendCommand(cmd)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("command failed: %s", resp.Error)
	}

	var result AddSyncFolderResponse
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) GetConfig() (*GetConfigResponse, error) {
	cmd := &Command{Type: "get_config"}
	resp, err := c.SendCommand(cmd)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("command failed: %s", resp.Error)
	}

	var config GetConfigResponse
	if err := json.Unmarshal(resp.Data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *Client) SetConfig(config map[string]interface{}) error {
	data, err := json.Marshal(&SetConfigRequest{Config: config})
	if err != nil {
		return err
	}

	cmd := &Command{
		Type: "set_config",
		Data: data,
	}

	resp, err := c.SendCommand(cmd)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("command failed: %s", resp.Error)
	}

	return nil
}
