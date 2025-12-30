package common

type ErrorResponse struct {
	Error string `json:"error"`
}

func GetErrorResponse(val string) ErrorResponse {
	return ErrorResponse{val}
}
