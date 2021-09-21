package issuertoken

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hewlettpackard/hpegl-provider-lib/pkg/token/errors"
)

const (
	retryLimit = 3
)

type TokenResponse struct {
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
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
	params := url.Values{}
	params.Add("client_id", clientID)
	params.Add("client_secret", clientSecret)
	params.Add("grant_type", "client_credentials")
	params.Add("scope", "hpe-tenant")

	url := fmt.Sprintf("%s/v1/token", c.identityServiceURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(params.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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
	case http.StatusUnauthorized, http.StatusForbidden:
		err = errors.MakeErrForbidden(clientID)

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
