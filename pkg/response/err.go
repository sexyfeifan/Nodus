package response

type ErrorResponse struct {
	Error string `json:"error"`
}

func InternalServerError() ErrorResponse {
	return ErrorResponse{Error: "Internal Server Error"}
}

// Message returns a user-facing error response with the given message.
func Message(msg string) ErrorResponse {
	return ErrorResponse{Error: msg}
}

// Error returns a generic error response without leaking internal details.
func Error(err error) ErrorResponse {
	return ErrorResponse{Error: "Internal Server Error"}
}
