// package GinProject
package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type User struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

func main() {
	r := gin.Default()

	// 用户注册接口
	r.POST("/register", func(c *gin.Context) {
		var user User

		// 使用 ShouldBindBodyWith 绑定请求体，并缓存 body
		if err := c.ShouldBindBodyWithJSON(&user); err != nil {
			// 如果 JSON 不合法，返回错误
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var user2 User
		if err := c.ShouldBindBodyWithJSON(&user2); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error2": err.Error()})
			return
		}
		// 返回成功响应
		c.JSON(http.StatusOK, gin.H{
			"message": "User registered successfully",
			"user":    user,
			"user2":   user2,
		})
	})

	err := r.Run(":8080")
	if err != nil {
		return
	} // 启动服务器
}
