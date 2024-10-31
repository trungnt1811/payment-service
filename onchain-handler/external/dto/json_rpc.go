package dto

// JSON-RPC request structure
type JSONRPCRequest struct {
	Jsonrpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
}

// JSON-RPC response structure
type JSONRPCResponse struct {
	Jsonrpc string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Result  string        `json:"result"` // balance in hex
	Error   *JSONRPCError `json:"error,omitempty"`
}

// JSON-RPC error structure
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
