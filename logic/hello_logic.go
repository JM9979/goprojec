package logic

import (
	"fmt"
	"ginproject/entity"
)

type HelloLogic struct{}

func (h *HelloLogic) SayHello(req entity.UsernameRequest) entity.HelloResponse {
	return entity.HelloResponse{
		Message: fmt.Sprintf("你好, %s!", req.Username),
		Code:    200,
	}
}
