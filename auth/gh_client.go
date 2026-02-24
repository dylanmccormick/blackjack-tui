package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type GithubClient struct {
	client   *http.Client
	clientID string
}

func NewGithubClient(clientID string) *GithubClient {
	client := &http.Client{Timeout: 20 * time.Second}
	return &GithubClient{
		client:   client,
		clientID: clientID,
	}
}

func (gc *GithubClient) RequestDeviceCode(ctx context.Context) (*GHDeviceResponse, error) {
	url := "https://github.com/login/device/code"
	data := map[string]string{"client_id": gc.clientID}
	return doRequest[GHDeviceResponse](gc.client, "POST", url, data, nil)
}

func (gc *GithubClient) PollAccessToken(ctx context.Context, deviceCode string) (string, error) {
	grantType := fmt.Sprintf("urn:ietf:params:oauth:grant-type:%s", "device_code")
	url := "https://github.com/login/oauth/access_token"
	data := map[string]string{"client_id": gc.clientID, "device_code": deviceCode, "grant_type": grantType}
	resp, err := doRequest[GHPollData](gc.client, "POST", url, data, nil)
	if err != nil {
		return "", err
	}
	if resp.Error != "" {
		return "", fmt.Errorf("still polling github, %s", resp.Error)
	}
	return resp.AccessToken, nil
}

func (gc *GithubClient) GetUsername(ctx context.Context, token string) (string, error) {
	headers := map[string]string{"Authorization": "Bearer " + token}

	type response struct {
		Login string `json:"login"`
	}

	resp, err := doRequest[response](gc.client, "GET", "https://api.github.com/user", nil, headers)
	if err != nil {
		return "", err
	}
	return resp.Login, nil
}

func (gc *GithubClient) CheckStarred(ctx context.Context, token, repo string) (bool, error) {
	headers := map[string]string{"Authorization": "Bearer " + token}
	url := fmt.Sprintf("https://api.github.com/user/starred/%s", repo)

	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := gc.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == 204, nil
}

func doRequest[T any](client *http.Client, method, url string, body map[string]string, headers map[string]string) (*T, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result T
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
