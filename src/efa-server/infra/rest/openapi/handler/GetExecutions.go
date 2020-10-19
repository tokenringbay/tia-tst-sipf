package handler

import (
	"efa-server/infra"
	"encoding/json"
	"fmt"
	"net/http"

	"bufio"
	"efa-server/infra/constants"
	"efa-server/infra/rest/generated/server/go"
	"github.com/gorilla/mux"
	"os"
	"strconv"
	"strings"
	"time"
)

//ExecutionListHandler retrieves the executions done on the application
func ExecutionListHandler(w http.ResponseWriter, r *http.Request) {
	statusMsg := ""
	var jsonResponse []byte

	vars := mux.Vars(r)
	limit := vars["limit"]
	status := vars["status"]

	if status != "all" && status != "failed" && status != "succeeded" {
		fmt.Println("Unsupported value for flag \"status\". It should be either \"succeeded/failed/all\".")
		return
	}

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		fmt.Println("Execution list : Limit flag value could not be converted from string to integer. Limit : ", limit)
		return
	}
	fmt.Println("Called Execution list with limit : " + limit + ", status : " + status)

	if ExecutionList, properr := infra.GetUseCaseInteractor().Db.GetExecutionLogList(limitInt, status); properr == nil {
		//Prepare the OpenAPI model
		var OpenAPIExecutionsResponse swagger.ExecutionsResponse
		OpenAPIExecutionsResponse.Items = make([]swagger.ExecutionResponse, 0, len(ExecutionList))
		for _, executionLog := range ExecutionList {
			startTime, _ := time.Parse(constants.DefaultTimeFormat, executionLog.StartTime)
			endTime, _ := time.Parse(constants.DefaultTimeFormat, executionLog.EndTime)
			OpenAPIExecutionsResponse.Items = append(OpenAPIExecutionsResponse.Items,
				swagger.ExecutionResponse{Id: executionLog.UUID, Command: executionLog.Command,
					Status:    executionLog.Status,
					StartTime: startTime, EndTime: endTime})

		}
		jsonResponse, _ = json.Marshal(OpenAPIExecutionsResponse)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonResponse)

	} else {
		statusMsg = fmt.Sprintf("Unable to retrieve Execution List\n")
		http.Error(w, statusMsg, http.StatusInternalServerError)
	}

}

//ExecutionGetHandler retrieves the detailed logs of the given execution-id
func ExecutionGetHandler(w http.ResponseWriter, r *http.Request) {
	statusMsg := ""
	var jsonResponse []byte

	vars := mux.Vars(r)
	execID := vars["id"]

	fmt.Println("Called Execution get with id : ", execID)

	if ExecutionLog, properr := infra.GetUseCaseInteractor().Db.GetExecutionLogByUUID(execID); properr == nil {
		logs := getLogsForExecutionID(execID)

		if len(logs) == 0 {
			logs = "<Logs not available for the Execution ID : " + execID + ">\n"
		}

		//Prepare the OpenAPI model
		//var OpenAPIDetailedExecutionResponse swagger.DetailedExecutionResponse
		startTime, _ := time.Parse(constants.DefaultTimeFormat, ExecutionLog.StartTime)
		endTime, _ := time.Parse(constants.DefaultTimeFormat, ExecutionLog.EndTime)
		OpenAPIDetailedExecutionResponse := swagger.DetailedExecutionResponse{
			Id:         ExecutionLog.UUID,
			Command:    ExecutionLog.Command,
			Parameters: ExecutionLog.Params,
			Status:     ExecutionLog.Status,
			StartTime:  startTime,
			EndTime:    endTime,
			Logs:       logs}
		jsonResponse, _ = json.Marshal(OpenAPIDetailedExecutionResponse)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonResponse)
	} else {
		statusMsg = fmt.Sprintf("Unable to retrieve Execution Log\n")
		http.Error(w, statusMsg, http.StatusInternalServerError)
	}
}

func getLogsForExecutionID(execID string) string {
	var logs string
	logFile, err := os.Open(constants.LogLocation)
	if err != nil {
		fmt.Println("Error while opening the file : ", err.Error())
	}
	defer logFile.Close()
	scanner := bufio.NewScanner(logFile)
	for scanner.Scan() {
		//fmt.Println(scanner.Text())
		if strings.Contains(scanner.Text(), execID) {
			logs += scanner.Text() + "\n"
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}
	return logs
}
