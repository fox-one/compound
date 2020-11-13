package views

// Default default view
type Default struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// DefaultSuccess default success view
var DefaultSuccess = Default{
	Code:    0,
	Message: "success",
}
