package controllers

import (
	"github.com/gin-gonic/gin"
	"yasumiProject-Backend/utils"
)

func Login(c *gin.Context) {
	var loginData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBind(&loginData); err != nil {
		utils.Fail(c, 400, "不正确输入-1")
	}

	// TODO: 验证用户名密码
	if loginData.Username == "" || loginData.Password == "" {
		utils.Fail(c, 400, "不正确输入-2")
		return
	} else {
		// 测试数据
		if loginData.Username == "admin" && loginData.Password == "123456" {
			token, err := utils.GenerateToken(1)
			if err != nil {
				utils.Fail(c, 400, "不正确输入-3")
				return
			}
			utils.Success(c, gin.H{"token": token})
		} else {
			utils.Fail(c, 400, "不正确输入-4")
			return
		}
	}
}
