package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const serverURL = "http://localhost:8080"

type CommandRequest struct {
	Command string `json:"command"`
	Text    string `json:"text"`
}

func sendCommand(command, text string) error {
	requestData := CommandRequest{
		Command: command,
		Text:    text,
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return err
	}

	resp, err := http.Post(serverURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var responseText map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&responseText); err != nil {
		return err
	}

	fmt.Println("Server Response:", responseText["text"])
	return nil
}
