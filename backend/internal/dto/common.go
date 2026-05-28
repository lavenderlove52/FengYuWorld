package dto

const (
	CodeSuccess           = 0
	CodeBadRequest        = 40000
	CodeDuplicateUser     = 40001
	CodeLoginFailed       = 40002
	CodeUnauthorized      = 40100
	CodeUserRevoked       = 40102 // JWT 有效但用户已被删除/吊销
	CodeForbidden         = 40300
	CodeAccountLocked     = 40301 // 账号已被锁定
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
