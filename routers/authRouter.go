package routers

import (
	"golang-jwtauth/controllers"

	"github.com/gin-gonic/gin"
)

func AuthRoutes(incomingRequest *gin.Engine) {
	incomingRequest.POST("/signup", controllers.SignUp())
	incomingRequest.POST("/login", controllers.Login())
}
