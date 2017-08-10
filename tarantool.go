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

type ThirdPartyError struct {
	s string
}

func newThirdPartyError(msg interface{}, args ...interface{}) error {
	switch msg.(type) {
	case error:
		s, _ := msg.(error)
		return &ThirdPartyError{s: s.Error()}
	case string:
		s, _ := msg.(string)
		return &ThirdPartyError{s: fmt.Sprintf(s, args...)}
	default:
		panic("unimplemented type")
	}
}

func (e *ThirdPartyError) Error() string {
	return e.s
}

type AvroSchemaError struct {
	s string
}

func newAvroSchemaError(msg interface{}, args ...interface{}) error {
	switch msg.(type) {
	case error:
		s, _ := msg.(error)
		return &AvroSchemaError{s: s.Error()}
	case string:
		s, _ := msg.(string)
		return &AvroSchemaError{s: fmt.Sprintf(s, args...)}
	default:
		panic("unimplemented type")
	}
}

func (e *AvroSchemaError) Error() string {
	return e.s
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
		return newAvroSchemaError(err)
	}

	// Setup the request
	req, err := http.NewRequest(http.MethodPost, tarantoolURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return newThirdPartyError(err)
	}

	// Call the Tarantool validator
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return newThirdPartyError(err)
	}
	if res.StatusCode != 200 {
		return newThirdPartyError("bad response from server: %d", res.StatusCode)
	}

	// parse the response
	var validationRes response
	if err := json.NewDecoder(res.Body).Decode(&validationRes); err != nil {
		return newAvroSchemaError(err)
	}
	defer res.Body.Close()

	// Make sure the op did not error out
	if validationRes.Error != responseErrorZero {
		return newAvroSchemaError("failed %+v", validationRes.Error)
	}

	// extract the op result
	opRes, ok := validationRes.Result[0][0].(string)
	if !ok {
		return newAvroSchemaError("Unknown Result: %s", validationRes.Result[0][0])
	}

	switch opRes {
	case resultOK:
		return nil
	case resultInvalidJSON:
		return newAvroSchemaError("invalid JSON")
	default:
		return newAvroSchemaError("result: %s", opRes)
	}
}
