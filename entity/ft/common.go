package ft

// 错误码定义
const (
	// 成功
	CodeSuccess = 200
	// 参数错误
	CodeInvalidParams = 400
	// 未找到记录
	CodeNotFound = 404
	// 内部服务器错误
	CodeServerError = 500
)

// APIResponse API通用响应结构
type APIResponse struct {
	// 错误码
	Code int `json:"code"`
	// 消息
	Message string `json:"message"`
	// 数据
	Data interface{} `json:"data"`
}

// NewSuccessResponse 创建成功响应
func NewSuccessResponse(data interface{}) APIResponse {
	return APIResponse{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
	}
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(code int, message string) APIResponse {
	return APIResponse{
		Code:    code,
		Message: message,
		Data:    nil,
	}
}
