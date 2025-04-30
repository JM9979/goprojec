package entity

type HelloResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
	TraceID string `json:"trace_id,omitempty"`
}
