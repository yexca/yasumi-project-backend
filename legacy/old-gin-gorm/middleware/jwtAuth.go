package middleware

import (
	"github.com/gin-gonic/gin"
	"yasumiProject-Backend/utils"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("JWTToken")
		if token == "" {
			utils.Fail(c, 401, "Token 不存在")
			return
		}

		claims, err := utils.ParseToken(token)
		if err != nil {
			utils.Fail(c, 401, "Token 验证失败")
			return
		}

		// 将需要的字段保存到上下文
		c.Set("userID", claims.UserID)
		c.Next()
	}
}
