package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/phuslu/log"
	"github.com/prigas-dev/backoffice-ai/operations"
	"github.com/victormf2/gosyringe"
)

func OperationsExecute(container *gosyringe.Container) {
	type ExecuteOperationRequestBody struct {
		Parameters map[string]any `json:"parameters"`
	}

	type ExecuteOperationSuccessResponseBody struct {
		Success bool `json:"success"`
		Result  any  `json:"result"`
	}

	type ExecuteOperationErrorResponseBody struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}

	http.HandleFunc("/operations/execute/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			log.Warn().Msgf("request with invalid method: %v", r.Method)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ExecuteOperationErrorResponseBody{
				Success: false,
				Message: "only POST method is allowed",
			})
			return
		}

		operationName, err := getOperationName(r)
		if err != nil {
			log.Warn().Msgf("request with invalid path %v: %v", r.URL.Path, err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ExecuteOperationErrorResponseBody{
				Success: false,
				Message: err.Error(),
			})
			return
		}

		requestBody := ExecuteOperationRequestBody{}
		err = json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			log.Warn().Msgf("failed to parse request parameters: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ExecuteOperationErrorResponseBody{
				Success: false,
				Message: "failed to parse request parameters",
			})
			return
		}

		executor, err := gosyringe.Resolve[operations.IOperationExecutor](container)
		if err != nil {
			log.Error().Msgf("error instantiating operation executor: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ExecuteOperationErrorResponseBody{
				Success: false,
				Message: fmt.Sprintf("error instantiating operation executor: %v", err),
			})
			return
		}

		result, err := executor.Execute(operationName, requestBody.Parameters)
		if err != nil {
			log.Error().Msgf("error on operation execution: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ExecuteOperationErrorResponseBody{
				Success: false,
				Message: fmt.Sprintf("error on operation execution: %v", err),
			})
			return
		}

		json.NewEncoder(w).Encode(ExecuteOperationSuccessResponseBody{
			Success: true,
			Result:  result,
		})
	})
}

func getOperationName(r *http.Request) (string, error) {

	pathSegments := strings.Split(
		strings.TrimLeft(r.URL.Path, "/"),
		"/",
	)

	if len(pathSegments) == 0 {
		return "", fmt.Errorf("request path is empty")
	}

	// /operations/execute/{operationName}
	if len(pathSegments) != 3 {
		return "", fmt.Errorf("invalid request path")
	}

	operationName := pathSegments[len(pathSegments)-1]

	return operationName, nil
}
