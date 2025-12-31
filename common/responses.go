package common

type ErrorResponse struct {
	Error string `json:"error"`
}

type EzqTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func GetErrorResponse(val string) ErrorResponse {
	return ErrorResponse{val}
}
