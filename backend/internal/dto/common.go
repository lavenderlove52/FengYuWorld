package dto

const (
	CodeSuccess           = 0
	CodeBadRequest        = 40000
	CodeDuplicateUser     = 40001
	CodeLoginFailed       = 40002
	CodeUnauthorized      = 40100
	CodeForbidden         = 40300
	CodeNotFound          = 40400
	CodeConflict          = 40900
	CodeInternalError     = 50000
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func OK(data interface{}) Response {
	return Response{Code: CodeSuccess, Message: "ok", Data: data}
}

func Err(code int, message string) Response {
	return Response{Code: code, Message: message}
}
