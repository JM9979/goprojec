package tbcapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type TbcApiService struct {
}

func NewTbcApiService() *TbcApiService {
	return &TbcApiService{}
}


func (s *TbcApiService) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Turing API is running.",
		"data":    gin.H{},
	})
}


