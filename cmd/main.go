package main

import (
	api "Anokha-main/internal/adapters/httpAdapter"
	"io"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	gin.DisableConsoleColor()
	f, _ := os.Create("gin.log")
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)

	server := gin.New()
	api.SetupAPI(server)

	err := server.Run(":8082")
	if err != nil {
		log.Panic("Server Error: ", err.Error())
		return
	}
}
