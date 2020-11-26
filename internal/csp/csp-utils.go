// Copyright 2020 program was created by VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package csp

import (
	"encoding/json"
	"fmt"

	httpUtils "github.com/vmware/concourse-vrealize-automation-resource/pkg/utils"
)

const (
	cspAPIBaseURL     = "https://console.cloud.vmware.com/csp/gateway/am/api"
	cspAccessTokenURL = cspAPIBaseURL + "/auth/api-tokens/authorize"
)

// Client provides all util methods for given refresh token
type Client struct {
	RefreshToken string `json:"refreshToken"`
}

// New cretes client pointer for all CSP related utils
func New(refreshToken string) *Client {
	return &Client{RefreshToken: refreshToken}
}

type accessTokenResponse struct {
	AccessToken string `json:"access_token"`
}

// GetAccessToken generates CSP access token
func (cspClient *Client) GetAccessToken() (string, error) {
	// Construct headers
	headers := make(map[string]string)
	headers["Content-Type"] = "application/x-www-form-urlencoded"

	// Construct Form data
	formData := make(map[string]string)
	formData["refresh_token"] = cspClient.RefreshToken

	// Fire the request
	response, err := httpUtils.PostHeadersFormDataRetry(cspAccessTokenURL, formData, headers)
	if err != nil || response.Code != 200 {
		return "", fmt.Errorf("Error while getting the CSP access token : %s. Error : %w", response.Message, err)
	}

	// Unmarshall the access token response
	var accessTokenResponse accessTokenResponse
	err = json.Unmarshal([]byte(response.ResponseString), &accessTokenResponse)
	if err != nil {
		return "", fmt.Errorf("Error while unmarshalling the CSP access token. Error : %w", err)
	}
	return accessTokenResponse.AccessToken, err
}

// GetAuthHeaders constructs Authorization header map with CSP access token
func (cspClient *Client) GetAuthHeaders() (map[string]string, error) {
	// Get CSP access token
	accessToken, err := cspClient.GetAccessToken()
	if err != nil {
		return nil, err
	}

	// Construct auth headers map
	authHeaders := make(map[string]string)
	authorization := fmt.Sprintf("Bearer %s", accessToken)
	authHeaders["Authorization"] = authorization
	return authHeaders, err
}
