package tarantool

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
)

const (
	tarantoolURL = "http://avro.tarantool.org/tarantool"
)

const (
	resultOK          = "Schema OK"
	resultInvalidJSON = "Invalid JSON"
)

type request struct {
	ID     string   `json:"id"`
	Method string   `json:"method"`
	Params []string `json:"params"`
}

type responseError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

var responseErrorZero = responseError{}

type response struct {
	ID     int             `json:"id"`
	Result [][]interface{} `json:"result"`
	Error  responseError   `json:"error"`
}

type ServiceError struct {
	err error
}

func newServiceError(err error) error {
	return &ServiceError{err}
}

func (e *ServiceError) Error() string {
	return e.err.Error()
}

// Validate takes an Avro schema and validates it using the online
// Tarantool Avro schema validator endpoint
func Validate(schema string) error {
	// Setup request body
	validationReq := &request{
		ID:     "1",
		Method: "compile",
		Params: []string{schema},
	}

	reqBody, err := json.Marshal(validationReq)
	if err != nil {
		return err
	}

	// Setup the request
	req, err := http.NewRequest(http.MethodPost, tarantoolURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}

	// Call the Tarantool validator
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		if netErr, ok := err.(net.Error); ok {
			if netErr.Temporary() {
				return newServiceError(err)
			}
		}

		return err
	}
	if res.StatusCode != 200 {
		return newServiceError(fmt.Errorf("bad response from server: %d", res.StatusCode))
	}

	// parse the response
	var validationRes response
	if err := json.NewDecoder(res.Body).Decode(&validationRes); err != nil {
		return err
	}
	defer res.Body.Close()

	// Make sure the op did not error out
	if validationRes.Error != responseErrorZero {
		return fmt.Errorf("failed %+v", validationRes.Error)
	}

	// extract the op result
	opRes, ok := validationRes.Result[0][0].(string)
	if !ok {
		return fmt.Errorf("Unknown Result: %s", validationRes.Result[0][0])
	}

	switch opRes {
	case resultOK:
		return nil
	case resultInvalidJSON:
		return fmt.Errorf("invalid JSON")
	default:
		return fmt.Errorf("result: %s", opRes)
	}
}
