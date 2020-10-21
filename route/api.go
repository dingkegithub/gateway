package route

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func Entry(c *gin.Context) {
	fmt.Println("api request", c.Request.URL, c.Request.RequestURI)
	c.JSONP(200, gin.H{
		"err": 0,
		"msg": "ok",
	})
}
