package main

import (
	"fmt"
	"golang-jwtauth/controllers"
	"golang-jwtauth/routers"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error on loading env file. err : %v", err)
		return
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	controllers.MakeCollection()
	controllers.CreateUniqueIndex()
	router := gin.New()
	router.Use(gin.Logger())

	routers.AuthRoutes(router)
	//routers.UserRoutes(router)
	fmt.Println("starting server...")
	router.Run(":" + port)

}
