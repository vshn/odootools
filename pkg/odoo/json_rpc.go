package odoo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/google/uuid"
)

// JsonRpcRequest represents a generic json-rpc request
type JsonRpcRequest struct {
	// ID should be a randomly generated value, either as a string or int. The
	// server will return this value in the response.
	ID string `json:"id,omitempty"`

	// Jsonrpc is always set to "2.0"
	Jsonrpc string `json:"jsonrpc,omitempty"`

	// Method to call, usually just "call"
	Method string `json:"method,omitempty"`

	// Params includes the actual request payload.
	Params interface{} `json:"params,omitempty"`
}

// NewJsonRpcRequest returns a JSON RPC request with its protocol fileds populated:
//
// * "id" will be set to a random UUID
// * "jsonrpc" will be set to "2.0"
// * "method" will be set to "call"
// * "params" will be set to whatever was passed in
func NewJsonRpcRequest(params interface{}) *JsonRpcRequest {
	return &JsonRpcRequest{
		ID:      uuid.NewString(),
		Jsonrpc: "2.0",
		Method:  "call",
		Params:  params,
	}
}

// Encode encodes the request as JSON in a buffer and returns the buffer.
func (r *JsonRpcRequest) Encode() (io.Reader, error) {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(r); err != nil {
		return nil, err
	}

	return buf, nil
}

type JsonRpcResponse struct {
	// ID that was sent with the request
	ID string `json:"id,omitempty"`
	// Jsonrpc is always set to "2.0"
	Jsonrpc string `json:"jsonrpc,omitempty"`
	// Result payload
	Result *json.RawMessage `json:"result,omitempty"`
}

// DecodeResult takes a buffer, decodes the intermediate JsonRpcResponse and
// then the contained "result" field into "result".
func DecodeResult(buf io.ReadCloser, result interface{}) error {
	defer buf.Close()

	// Decode intermediate
	var res JsonRpcResponse
	if err := json.NewDecoder(buf).Decode(&res); err != nil {
		return fmt.Errorf("decode intermediate: %w", err)
	}

	return json.Unmarshal(*res.Result, result)
}
