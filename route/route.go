package route

import (
	"fmt"

	"github.com/dingkegithub/gateway/webfilter"
	"github.com/gin-gonic/gin"
)

func Route() {
	corsInstance := webfilter.NewCors()
	antispam := webfilter.NewAntispam()
	route := gin.Default()
	route.Use(corsInstance.CorsCheck)
	route.Use(webfilter.SessionInstance.SessionCheck)
	route.Use(antispam.SpamCheck)
	route.OPTIONS("/api/:model/*cmd", Entry)
	route.POST("/api/:model/*cmd", Entry)
	route.GET("/api/:model/*cmd", Entry)

	err := route.Run(":18082")
	if err != nil {
		fmt.Println("com.dk.gateway start failed ", err.Error())
	}
}
