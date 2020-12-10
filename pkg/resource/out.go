// Copyright 2020 program was created by VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package resource

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/vmware/concourse-vrealize-automation-resource/internal/csp"
	"github.com/vmware/concourse-vrealize-automation-resource/internal/vra"
)

const (
	defaultWaitTimeoutMinutes = 1440 // 24 hours
	pollIntervalSeconds       = 30
)

func out(source VRASource, params OutParams) (version interface{}, metadata []interface{}, err error) {
	// Authenticate
	log.Println("Authenticating with vRealize Automation...")
	cspClient := csp.New(source.APIToken)
	if err != nil {
		return nil, nil, fmt.Errorf("Error while authenticating:%w", err)
	}
	log.Println("vRealize Automation authentication is successful")

	// Fetch Pipeline ID
	log.Println("Fetching pipeline ID from name...")
	csClient := vra.New(cspClient)
	pipelineID, err := csClient.GetPipelineIDFromName(source.Pipeline)
	if err != nil {
		return nil, nil, fmt.Errorf("Error while getting pipeline ID from name:%w", err)
	}
	log.Println("Pipeline ID is fetched successfully: " + pipelineID)

	// Execute vRealize Automation pipeline
	log.Println("Triggering vRealize Automation pipeline...")

	// Construct pipeline input params
	exeReq := vra.PipelineExecutionReq{Comments: "Triggered by Concourse CI",
		Input: params.Input}
	execResp, err := csClient.ExecutePipeline(pipelineID, exeReq)
	if err != nil {
		return nil, nil, fmt.Errorf("Error while executing vRealize Automation pipeline:%w", err)
	}
	log.Println("vRealize Automation pipeline is triggered successfully")

	// Do not wait for the execution to be completed if wait is set to false
	if !params.Wait {
		var metadataSlice []interface{} = make([]interface{}, 1)
		metadataSlice = append(metadataSlice, MetadataField{Name: "executionId", Value: execResp.ExecutionID})
		return VRAVersion{Value: "TODO"}, metadataSlice, nil
	}

	// Use timeout value from config if provided
	var finalWaitTimeout int
	if params.WaitTimeout > 0 {
		finalWaitTimeout = params.WaitTimeout
	} else {
		finalWaitTimeout = defaultWaitTimeoutMinutes
	}

	// Wait for the execution to finish with timeout
	// Channels for polling and timing out
	pollChannel := time.Tick(time.Second * pollIntervalSeconds)
	timeoutChannel := time.After(time.Minute * time.Duration(finalWaitTimeout))

	// Buffered channels to receive pipeline execution and errors
	pipelineExecChannel := make(chan vra.PipelineExecution, 1)
	errorChannel := make(chan error, 1)

	// Keep getting the latest status of pipeline execution
	// in a separate Go routine
	go func() {
		for {
			select {
			case <-timeoutChannel:
				errorChannel <- errors.New("Timedout while waiting for vRealize Automation pipeline to complete")
				return
			case <-pollChannel:
				pipelineExec, err := csClient.GetPipelineExecution(execResp.ExecutionID)
				if err != nil {
					errorChannel <- fmt.Errorf("Error while getting pipeline status::%w", err)
					return
				}
				log.Println("vRealize Automation pipeline's current status: " + pipelineExec.Status)
				if pipelineExec.Status == "COMPLETED" || pipelineExec.Status == "FAILED" {
					pipelineExecChannel <- pipelineExec
					return
				}
			}
		}
	}()

	log.Println("Waiting for vRealize Automation pipeline to complete...")
	select {
	case pipelineExec := <-pipelineExecChannel:
		// Executions is either completed or failed
		log.Println("vRealize Automation pipeline finished execution with status: " + pipelineExec.Status)
		outputMeta := processOutput(pipelineExec)
		return VRAVersion{Value: "TODO"}, outputMeta, nil
	case err := <-errorChannel:
		return VRAVersion{Value: "TODO"}, nil, err
	}
}

func processOutput(execution vra.PipelineExecution) []interface{} {
	var metadataSlice []interface{} = make([]interface{}, 1)

	// Add execution ID and overall pipeline status
	metadataSlice = append(metadataSlice, MetadataField{Name: "executionId", Value: execution.ID})
	metadataSlice = append(metadataSlice, MetadataField{Name: "status", Value: execution.Status})

	// Add Output params
	for outputParam, outputParamVal := range execution.Output {
		metadataSlice = append(metadataSlice, MetadataField{Name: "output~" + outputParam, Value: outputParamVal})
	}

	// Add stage and tasks execution details
	for _, stageName := range execution.StageOrder {
		stageExec := execution.Stages[stageName]
		for _, taskName := range stageExec.TaskOrder {
			taskExec := stageExec.Tasks[taskName]

			// Add task status
			metadataSlice = append(metadataSlice, MetadataField{Name: stageName + "~" + taskName + "~status", Value: taskExec.Status})

			// Based on task type, add additional data
			switch taskExec.Type {
			case "Jenkins":
				metadataSlice = append(metadataSlice, MetadataField{Name: stageName + "~" + taskName + "~jobUrl", Value: taskExec.Output["jobUrl"].(string)})
				// TODO: Add more cases for other task types
			}
		}
	}

	return metadataSlice
}