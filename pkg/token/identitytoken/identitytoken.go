package identitytoken

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/hewlettpackard/hpegl-provider-lib/pkg/token/errors"
)

const (
	retryLimit = 3
)

type GenerateTokenInput struct {
	TenantID     string `json:"tenant_id"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	GrantType    string `json:"grant_type"`
}

type TokenResponse struct {
	TokenType       string    `json:"token_type"`
	AccessToken     string    `json:"access_token"`
	RefreshToken    string    `json:"refresh_token"`
	Expiry          time.Time `json:"expiry"`
	ExpiresIn       int       `json:"expires_in"`
	Scope           string    `json:"scope"`
	AccessTokenOnly bool      `json:"accessTokenOnly"`
}

type Client struct {
	identityServiceURL string
	httpClient         httpClient
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// New creates a new identity Client object
func New(identityServiceURL string) *Client {
	client := &http.Client{Timeout: 10 * time.Second}
	identityServiceURL = strings.TrimRight(identityServiceURL, "/")
	return &Client{
		identityServiceURL: identityServiceURL,
		httpClient:         client,
	}
}

func doRetries(call func() (*http.Response, error), retries int) (*http.Response, error) {
	var resp *http.Response
	var err error

	for {
		resp, err = call()
		if err != nil {
			return nil, err
		}

		if !isStatusRetryable(resp.StatusCode) || retries == 0 {
			break
		}

		time.Sleep(3 * time.Second)
		retries--
	}

	return resp, nil
}

func (c *Client) GenerateToken(ctx context.Context, tenantID, clientID, clientSecret string) (string, error) {
	params := GenerateTokenInput{
		TenantID:     tenantID,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		GrantType:    "client_credentials",
	}

	url := fmt.Sprintf("%s/v1/token", c.identityServiceURL)

	b, err := json.Marshal(params)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(b)))

	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := doRetries(func() (*http.Response, error) {
		return c.httpClient.Do(req)
	}, retryLimit)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		break
	case http.StatusBadRequest:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		msg := fmt.Sprintf("Bad request: %v", string(body))
		err = errors.MakeErrBadRequest(errors.ErrorResponse{
			ErrorCode: "ErrGenerateTokenBadRequest",
			Message:   msg,
		})

		return "", err
	case http.StatusForbidden:
		err = errors.MakeErrForbidden(clientID)

		return "", err
	case http.StatusUnauthorized:
		err = errors.MakeErrUnauthorized(clientID)

		return "", err
	default:
		msg := fmt.Sprintf("Unexpected status code %v", resp.StatusCode)
		err = errors.MakeErrInternalError(errors.ErrorResponse{
			ErrorCode: "ErrGenerateTokenUnexpectedResponseCode",
			Message:   msg,
		})

		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var token TokenResponse

	err = json.Unmarshal(body, &token)
	if err != nil {
		return "", err
	}

	return token.AccessToken, nil
}

func isStatusRetryable(statusCode int) bool {
	if statusCode == http.StatusInternalServerError || statusCode == http.StatusTooManyRequests {
		return true
	}

	return false
}
