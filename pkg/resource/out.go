// Copyright 2020 program was created by VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package resource

import (
	"fmt"
	"log"
	"time"

	"github.com/vmware/concourse-vrealize-automation-resource/internal/csp"
	"github.com/vmware/concourse-vrealize-automation-resource/internal/vra"
)

const (
	waitTimeoutMinutes  = 1440 // 24 hours
	pollIntervalSeconds = 5
)

func out(source VRASource, params OutParams) (version interface{}, metadata []interface{}, err error) {
	log.Println("Authenticating with vRealize Automation...")
	cspClient := csp.New(source.APIToken)
	accessToken, err := cspClient.GetAccessToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error while authenticating:%w", err)
	}
	log.Println("vRealize Automation authentication is successful")

	log.Println("Fetching pipeline ID from name...")
	csClient := vra.New(cspClient)
	pipelineID, err := csClient.GetPipelineIDFromName(source.Pipeline, accessToken)
	if err != nil {
		return nil, nil, fmt.Errorf("Error while getting pipeline ID from name:%w", err)
	}
	log.Println("Pipeline ID is fetched successfully: " + pipelineID)

	log.Println("Triggering vRealize Automation pipeline...")
	inputMap := map[string]string{"Changeset": params.Changeset}
	exeReq := vra.PipelineExecutionReq{Comments: "Triggered by Concourse CI",
		Input: inputMap}
	execResp, err := csClient.ExecutePipeline(pipelineID, exeReq)
	if err != nil {
		return nil, nil, fmt.Errorf("Error while executing vRealize Automation pipeline:%w", err)
	}
	log.Println("vRealize Automation pipeline is triggered successfully")

	// Do not wait for the execution to be completed if wait is set to false
	if !params.Wait {
		var metadataSlice []interface{} = make([]interface{}, 1)
		metadataSlice = append(metadataSlice, MetadataField{Name: "vRealize Automation pipeline execution ID:", Value: execResp.ExecutionID})
		return VRAVersion{Value: "TODO"}, metadataSlice, nil
	}

	endSignal := make(chan bool, 1)
	timeout := time.After(time.Minute * waitTimeoutMinutes)
	pollInt := time.Second * pollIntervalSeconds

	var pipelineExec vra.PipelineExecution
	log.Println("Waiting for vRealize Automation pipeline to complete...")
	for {
		select {
		case <-endSignal:
			log.Println("vRealize Automation pipeline finished execution with status: " + pipelineExec.Status)
			var metadataSlice []interface{} = make([]interface{}, 1)
			metadataSlice = append(metadataSlice, MetadataField{Name: "Execution ID:", Value: execResp.ExecutionID})
			metadataSlice = append(metadataSlice, MetadataField{Name: "Status:", Value: pipelineExec.Status})

			for outputParam, outputParamVal := range pipelineExec.Output {
				metadataSlice = append(metadataSlice, MetadataField{Name: "Pipeline Output: " + outputParam, Value: outputParamVal})
			}

			for _, stageName := range pipelineExec.StageOrder {
				stageExec := pipelineExec.Stages[stageName]
				for _, taskName := range stageExec.TaskOrder {
					taskExec := stageExec.Tasks[taskName]
					switch taskExec.Type {
					case "Jenkins":
						metadataSlice = append(metadataSlice, MetadataField{Name: stageName + " (s) > " + taskName + " (t)", Value: "Status: " + taskExec.Status + ", Job URL:" + taskExec.Output["jobUrl"].(string)})
					default:
						metadataSlice = append(metadataSlice, MetadataField{Name: stageName + " (s) > " + taskName + " (t)", Value: "Status: " + taskExec.Status})
					}
				}
			}
			return VRAVersion{Value: "TODO"}, metadataSlice, nil
		case <-timeout:
			break
		default:
			pipelineExec, err = csClient.GetPipelineExecution(execResp.ExecutionID)
			if err != nil {
				return nil, nil, fmt.Errorf("Error while getting pipeline status::%w", err)
			}

			if pipelineExec.Status == "COMPLETED" || pipelineExec.Status == "FAILED" {
				endSignal <- true
			}
			log.Println("vRealize Automation pipeline's current status: " + pipelineExec.Status + ". Still waiting...")
		}
		time.Sleep(pollInt)
	}
}
