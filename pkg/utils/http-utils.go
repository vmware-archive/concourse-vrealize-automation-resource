// Copyright 2020 program was created by VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"io/ioutil"
	"net"
	"net/http"
	"time"

	resty "github.com/go-resty/resty/v2"
)

// Response holds the processed data
// from HTTP calls
type Response struct {
	Code           int
	Message        string
	ResponseString string
	Headers        http.Header
}

const (
	defaultRetryCount              = 3
	defaultRetryWaitSeconds        = time.Second * 3
	keepAliveTimeout               = time.Second * 300 // 5 minutes
	maxIdleConnectionsLimit        = 100
	maxIdleConnectionsPerHostLimit = 100
)

var restyClient *resty.Client

// Initialize the new Resty client
// as part of init and reuse it
func init() {
	restyClient = getNewRestyClient()
}

// Get makes an HTTP call to given URL and returns Response
func Get(URL string) (Response, error) {
	return GetHeadersCustomRetry(URL, nil, -1, -1)
}

// GetRetry makes an HTTP call to given URL and returns Response
func GetRetry(URL string) (Response, error) {
	return GetHeadersCustomRetry(URL, nil, defaultRetryCount, defaultRetryWaitSeconds)
}

// GetCustomRetry makes an HTTP call to given URL with custom
// retry options and returns Response
func GetCustomRetry(URL string, retryCount int, retryWaitSeconds time.Duration) (Response, error) {
	return GetHeadersCustomRetry(URL, nil, retryCount, retryWaitSeconds)
}

// GetHeaders makes an HTTP call to given URL with headers
// and returns response
func GetHeaders(URL string, headers map[string]string) (Response, error) {
	return GetHeadersCustomRetry(URL, headers, -1, -1)
}

// GetHeadersRetry makes an HTTP call to given URL with headers
// and returns response. It also retries for failures.
func GetHeadersRetry(URL string, headers map[string]string) (Response, error) {
	return GetHeadersCustomRetry(URL, headers, defaultRetryCount, defaultRetryWaitSeconds)
}

// GetHeadersCustomRetry makes an HTTP call to given URL with headers
// and returns response. It also retries for failures with given retry
// count and wait seconds.
func GetHeadersCustomRetry(URL string, headers map[string]string, retryCount int, retryWaitSeconds time.Duration) (Response, error) {
	request := getNewRestyRequest(retryCount, retryWaitSeconds)
	resp, err := request.
		SetHeaders(headers).
		Get(URL)
	return processResponse(resp), err
}

// Post makes an HTTP call to given URL and returns response
func Post(URL string, requestBody string) (Response, error) {
	return PostHeadersCustomRetry(URL, requestBody, nil, -1, -1)
}

// PostHeaders makes an HTTP call to given URL with headers
// and returns response
func PostHeaders(URL string, requestBody string, headers map[string]string) (Response, error) {
	return PostHeadersCustomRetry(URL, requestBody, headers, -1, -1)
}

// PostHeadersRetry makes an HTTP call to given URL with headers
// and returns response. It also retries for failures.
func PostHeadersRetry(URL string, requestBody string, headers map[string]string) (Response, error) {
	return PostHeadersCustomRetry(URL, requestBody, headers, defaultRetryCount, defaultRetryWaitSeconds)
}

// PostHeadersCustomRetry makes an HTTP call to given URL with headers
// and returns response. It also retries for failures with given retry
// count and wait seconds.
func PostHeadersCustomRetry(URL string, requestBody string, headers map[string]string, retryCount int, retryWaitSeconds time.Duration) (Response, error) {
	request := getNewRestyRequest(retryCount, retryWaitSeconds)
	if requestBody != "" {
		request.
			SetBody(requestBody)
	}
	resp, err := request.
		SetHeaders(headers).
		Post(URL)
	return processResponse(resp), err
}

// PostHeadersFormDataRetry makes an HTTP call to given URL with form data,
// headers and returns response. It also retries for failures.
func PostHeadersFormDataRetry(URL string, formData map[string]string, headers map[string]string) (Response, error) {
	return PostHeadersFormDataCustomRetry(URL, formData, headers, defaultRetryCount, defaultRetryWaitSeconds)
}

// PostHeadersFormDataCustomRetry makes an HTTP call to given URL with form data,
// headers and returns response. It also retries for failures with given retry
// count and wait seconds.
func PostHeadersFormDataCustomRetry(URL string, formData map[string]string, headers map[string]string, retryCount int, retryWaitSeconds time.Duration) (Response, error) {
	request := getNewRestyRequest(-1, -1)
	if formData != nil {
		request.
			SetFormData(formData)
	}
	resp, err := request.
		SetHeaders(headers).
		Post(URL)
	return processResponse(resp), err
}

// ParseResponse reads given Response body
// and return its string type value
func ParseResponse(response *http.Response) (string, error) {
	// Read response string
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	responseString := string(body)
	return responseString, nil
}

func processResponse(resp *resty.Response) Response {
	return Response{resp.StatusCode(), resp.Status(), resp.String(), resp.Header()}
}

func getNewRestyRequest(retryCount int, retryWaitSeconds time.Duration) *resty.Request {
	// If restyClient is  not initialized
	// already, assign a new one
	if restyClient == nil {
		restyClient = getNewRestyClient()
	}
	if retryCount > 0 {
		restyClient.
			SetRetryCount(retryCount)
	}
	if retryWaitSeconds.Seconds() > 0 {
		restyClient.
			SetRetryWaitTime(retryWaitSeconds)
	}
	return restyClient.R()
}

func getNewRestyClient() *resty.Client {
	// Create new resty client
	restyClient := resty.New()

	// Set limits to connections so that
	// connections are not blocked
	customTransport := &http.Transport{
		//TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // Uncomment to disable SSL check
		Dial: (&net.Dialer{
			KeepAlive: keepAliveTimeout,
		}).Dial,
		MaxIdleConns:        maxIdleConnectionsLimit,
		MaxIdleConnsPerHost: maxIdleConnectionsPerHostLimit}
	restyClient.SetTransport(customTransport)

	restyClient.
		SetRedirectPolicy(resty.
			FlexibleRedirectPolicy(10))
	return restyClient
}
