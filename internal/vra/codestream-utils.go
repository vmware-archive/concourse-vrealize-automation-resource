package vra

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/vmware/concourse-vrealize-automation-resource/internal/csp"
	httpUtils "github.com/vmware/concourse-vrealize-automation-resource/pkg/utils"
)

const (
	vraAPIBaseURL          = "https://api.mgmt.cloud.vmware.com"
	pipelineExecutionURL   = "/codestream/api/executions/%s"
	pipelineExecutionModel = "/codestream/api/pipelines/%s/executions"
	pipelineIDURIPath      = "/codestream/api/pipelines"
	getExecutionURL        = vraAPIBaseURL + "/codestream/api/executions/%s?expand=PIPELINE_STAGE_TASK"
)

// Client provides all util methods for
// the given CSP client
type Client struct {
	CspClient *csp.Client
}

// New creates Code Stream client pointer
func New(cspClient *csp.Client) *Client {
	return &Client{CspClient: cspClient}
}

type Links struct {
	Links []string
}

// PipelineExecutionReq holds execute request body
type PipelineExecutionReq struct {
	Comments string            `json:"comments"`
	Input    map[string]string `json:"input"`
}

// PipelineExecutionResp holds execution response body
type PipelineExecutionResp struct {
	ExecutionID    string `json:"executionId"`
	ExecutionLink  string `json:"executionLink"`
	ExecutionIndex int    `json:"executionIndex"`
}

// PipelineExecution holds pipeline execution record
type PipelineExecution struct {
	ID            string                            `json:"id"`
	Index         int                               `json:"index"`
	Project       string                            `json:"project"`
	Status        string                            `json:"status"`
	StatusMessage string                            `json:"statusMessage"`
	Comments      string                            `json:"comments"`
	Output        map[string]string                 `json:"output"`
	StageOrder    []string                          `json:"stageOrder"`
	Stages        map[string]PipelineStageExecution `json:"stages"`
}

// PipelineStageExecution holds pipeline stage execution record
type PipelineStageExecution struct {
	Status    string                           `json:"status"`
	TaskOrder []string                         `json:"taskOrder"`
	Tasks     map[string]PipelineTaskExecution `json:"tasks"`
}

// PipelineTaskExecution holds pipeline task execution record
type PipelineTaskExecution struct {
	Status string                 `json:"status"`
	Type   string                 `json:"type"`
	Output map[string]interface{} `json:"output"`
}

// ExecutePipeline executes the pipeline with given request body
func (csClient *Client) ExecutePipeline(pipelineID string, execReq PipelineExecutionReq) (PipelineExecutionResp, error) {
	headers, err := getHeaders(csClient)
	if err != nil {
		return PipelineExecutionResp{}, err
	}

	// Marshal request struct to JSON
	requestBodyJSONBytes, err := json.Marshal(execReq)
	if err != nil {
		return PipelineExecutionResp{}, err
	}

	// Fire the request
	executePipelineURL := fmt.Sprintf(vraAPIBaseURL+pipelineExecutionModel, pipelineID)
	response, err := httpUtils.PostHeadersRetry(executePipelineURL, string(requestBodyJSONBytes), headers)
	if err != nil || response.Code != 202 {
		return PipelineExecutionResp{}, fmt.Errorf("Error while executing pipeline: %s. %w", response.Message, err)
	}

	// Parse execution response
	var executionResponse PipelineExecutionResp
	err = json.Unmarshal([]byte(response.ResponseString), &executionResponse)
	if err != nil {
		return PipelineExecutionResp{}, fmt.Errorf("Error while unmarshalling the execution response : %s. %v", response.Message, err)
	}
	return executionResponse, nil
}

// GetPipelineExecution fetches pipeline execution record for given
// executionID
func (csClient *Client) GetPipelineExecution(executionID string) (PipelineExecution, error) {
	headers, err := getHeaders(csClient)
	if err != nil {
		return PipelineExecution{}, err
	}

	// Fire the request
	getExecutionURL := fmt.Sprintf(getExecutionURL, executionID)
	response, err := httpUtils.GetHeadersRetry(getExecutionURL, headers)
	if err != nil || response.Code != 200 {
		return PipelineExecution{}, fmt.Errorf("Error while getting pipeline execution details: %s. %w", response.Message, err)
	}

	// Parse the pipeline execution
	var pipelineExecution PipelineExecution
	err = json.Unmarshal([]byte(response.ResponseString), &pipelineExecution)
	if err != nil {
		return PipelineExecution{}, fmt.Errorf("Error while unmarshalling the pipeline execution response : %s. %v", response.Message, err)
	}
	return pipelineExecution, nil
}

// GetPipelineIDFromName returns pipline ID of the given pipeline name
func (csClient *Client) GetPipelineIDFromName(pipelineName string, authToken string) (string, error) {
	// Construct API URL with query param encoding
	baseURL, _ := url.Parse(vraAPIBaseURL)
	baseURL.Path += pipelineIDURIPath
	params := url.Values{}
	pipilineNameAdd := fmt.Sprintf("name eq '%s'", pipelineName)
	params.Add("$filter", pipilineNameAdd)
	baseURL.RawQuery = params.Encode()
	searchPipelinesURL := baseURL.String()

	// Construct headers
	headers, err := getHeaders(csClient)
	if err != nil {
		return "", err
	}

	// Fire
	resp, err := httpUtils.GetHeadersRetry(searchPipelinesURL, headers)
	if err != nil {
		return "", err
	}

	// Parse Links
	var links Links
	err = json.Unmarshal([]byte(resp.ResponseString), &links)
	if err != nil {
		fmt.Println(err)
	}
	var linksExtracted []string
	for _, link := range links.Links {
		linksExtracted = append(linksExtracted, strings.SplitAfter(link, "/codestream/api/pipelines/")[1])
	}

	if len(linksExtracted) < 1 {
		return "", nil
	} else if len(linksExtracted) > 1 {
		return "", errors.New("More than 1 matching pipeline found for given name")
	}

	return linksExtracted[0], nil
}

func getHeaders(csClient *Client) (map[string]string, error) {
	headers, err := csClient.CspClient.GetAuthHeaders()
	if err != nil {
		return nil, err
	}
	headers["Content-Type"] = "application/json"
	return headers, nil
}
