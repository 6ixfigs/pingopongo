package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/6ixfigs/pingypongy/internal/config"
)

const baseURL = "https://slack.com/api"

func GetUserInfo(userID string) (string, error) {
	cfg, err := config.Get()
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/users.info", baseURL)
	payload := []byte(fmt.Sprintf("token=%s&user=%s", cfg.BotToken, userID))
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payload))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	var userInfoResponse UserInfoResponse
	err = json.NewDecoder(resp.Body).Decode(&userInfoResponse)
	if err != nil {
		return "", err
	}

	if !userInfoResponse.Ok {
		return "", fmt.Errorf("%s", userInfoResponse.Error)
	}

	if userInfoResponse.User.IsBot {
		return "", fmt.Errorf("%s", "user is bot")
	}

	return userInfoResponse.User.RealName, nil
}
