package http

type Response struct {
	Status   int  `json:"status"`
	Data     any  `json:"data"`
	IsCached bool `json:"is_cached,omitempty"`
}

type ErrorResponse struct {
	Status int `json:"status"`
	Errors any `json:"errors"`
}

type GeneralError struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Errors  []string `json:"errors,omitempty"`
}

type ErrorMap struct {
	Errors map[string]any `json:"errors"`
}
