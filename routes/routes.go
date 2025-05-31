package routes

import (
	"github.com/gin-gonic/gin"
	"yasumiProject-Backend/controllers"
)

func InitRoutes(r *gin.Engine) {
	// 无需验证 JWT
	r.POST("/api/login", controllers.Login)

	// 需要验证 JWT
	//auth := r.Group("/api", middleware.JWTAuth())
	//{
	//	auth.GET("/user/info", controllers.GetUserInfo)
	//}
}
