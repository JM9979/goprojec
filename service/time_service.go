package service

import (
	"ginproject/entity"
	"ginproject/logic"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CurrentTimeService struct {
	logic *logic.CurrentTimeLogic
}

func NewCurrentTimeService() *CurrentTimeService {
	return &CurrentTimeService{
		logic: &logic.CurrentTimeLogic{},
	}
}

// GetCurrentTimeService 处理获取当前时间的服务逻辑
func (u *CurrentTimeService)GetCurrentTimeService(c *gin.Context) {
	timeResp, err := logic.GetCurrentTimeLogic()
	if err != nil {
		c.JSON(http.StatusOK, entity.NewTimeResponse(500, "内部错误"))
		return
	}
	c.JSON(http.StatusOK, timeResp)
}
