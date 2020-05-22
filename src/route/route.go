package route

import (
	"com.dk.gateway/src/webfilter"
	"fmt"
	"github.com/gin-gonic/gin"
)

func Route()  {
	route := gin.Default()
	route.Use(webfilter.CorsInstance.CorsCheck)
	route.Use(webfilter.SessionInstance.SessionCheck)
	route.OPTIONS("/api/:model/*cmd", Entry)
	route.POST("/api/:model/*cmd", Entry)
	route.GET("/api/:model/*cmd", Entry)

	err := route.Run(":18082")
	if err != nil {
		fmt.Println("com.dk.gateway start failed ", err.Error())
	}
}
