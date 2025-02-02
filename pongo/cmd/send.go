package cmd

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const serverURL = "http://localhost:8080"

func sendCommand(path string, formData map[string]string, method string) error {

	form := url.Values{}
	for key, value := range formData {
		form.Set(key, value)
	}

	req, err := http.NewRequest(method, serverURL+path, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Println("Response form server: ", resp.Status)
	return nil
}
