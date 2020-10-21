package netutils

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type RespBody map[string]interface{}

func BuildResponseBody(err int, msg string, data interface{}) RespBody {
	return map[string]interface{}{
		"error": err,
		"msg":   msg,
		"data":  data,
	}
}

func StdResponse(c *gin.Context, status int, b RespBody) {
	if b == nil {
		c.JSON(status, nil)
		return
	}
	c.JSON(status, b)
}

func IsPreflight(c *gin.Context) bool {
	isOptionReq := strings.ToUpper(c.Request.Method) == http.MethodOptions
	hasOrigin := len(c.Request.Header.Get("Origin")) > 0
	hasAccessCtrReqMethod := len(c.Request.Header.Get("Access-Control-Request-Method")) > 0
	return isOptionReq && hasOrigin && hasAccessCtrReqMethod
}
