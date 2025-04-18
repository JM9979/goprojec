package logic

import (
	"ginproject/entity"
)

type CurrentTimeLogic struct{}

// GetCurrentTimeLogic 获取当前时间逻辑处理
func GetCurrentTimeLogic() (entity.TimeResponse, error) {
	return entity.NewTimeResponse(200, "成功获取当前时间"), nil
}
