package utils

import "github.com/gin-gonic/gin"

func Success(c *gin.Context, data interface{}) {
	c.JSON(200, gin.H{"code": 0, "msg": "success", "data": data})
}

func Fail(c *gin.Context, code int, msg string) {
	c.JSON(200, gin.H{"code": code, "msg": msg, "data": nil})
}
