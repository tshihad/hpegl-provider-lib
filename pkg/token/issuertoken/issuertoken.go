package issuertoken

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/hewlettpackard/hpegl-provider-lib/pkg/token/errors"
	tokenutil "github.com/hewlettpackard/hpegl-provider-lib/pkg/token/token-util"
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

func GenerateToken(ctx context.Context, tenantID, clientID, clientSecret string, identityServiceURL string, httpClient tokenutil.HttpClient) (string, error) {
	params := url.Values{}
	params.Add("client_id", clientID)
	params.Add("client_secret", clientSecret)
	params.Add("grant_type", "client_credentials")
	params.Add("scope", "hpe-tenant")

	url := fmt.Sprintf("%s/v1/token", identityServiceURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(params.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := tokenutil.DoRetries(func() (*http.Response, error) {
		return httpClient.Do(req)
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
