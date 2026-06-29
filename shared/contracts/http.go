package contracts

// API Response is the standard response envelope for all HTTP endpoints
type APIResponse struct {
	Data  any       `json:"data,omitempty"`
	Error *APIError `json:"error,omitempty"`
}

// NOTE: omitempty tags so can ommit when not needed. If success, ommit error cause it is nil. When fail, ommit data case data should be nul

// APIError carries a machien-readable code and human-redeable message
type APIError struct {
	Code    string `json:"code"`
	Messgae string `json:"message"`
}
