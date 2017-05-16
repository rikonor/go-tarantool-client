package tarantool

import (
	"bytes"
	"encoding/json"
	"fmt"
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
		return err
	}
	if res.StatusCode != 200 {
		return fmt.Errorf("bad response")
	}

	// parse the response
	var validationRes response
	if err := json.NewDecoder(res.Body).Decode(&validationRes); err != nil {
		return err
	}
	defer res.Body.Close()

	// Make sure the op did not error out
	if validationRes.Error != responseErrorZero {
		return fmt.Errorf("Failed %+v", validationRes.Error)
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
		return fmt.Errorf("Invalid JSON")
	default:
		return fmt.Errorf("Error: %s", opRes)
	}
}
