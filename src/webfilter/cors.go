/*
 * cors: cross origin resource sharding
 */
package webfilter

import (
	"github.com/fatih/set"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

const CorsUrl = "http://gz-data.com"

type Cors struct {
	AllowedHost    string
	AllowedHostSet set.Interface
}

func NewCors() *Cors {
	allowedSet := set.New(set.ThreadSafe)
	allowedSet.Add(CorsUrl)
	return &Cors{
		AllowedHost:    CorsUrl,
		AllowedHostSet: allowedSet,
	}
}

var CorsInstance *Cors

func init() {
	CorsInstance = NewCors()
}

func (cors *Cors) CorsCheck(c *gin.Context) {
	/*
	 * option 请求得到response中的header
	 * Access-Control-Allow-Credentials:true;
	 * Access-Control-Allow-Methods:GET,POST,OPTIONS;
	 * Access-Control-Allow-Origin:gz-data.com;
	 * Access-Control-Max-Age:3600;
	 * Content-Type:application/json;
	 */
	// 允许垮cookie访问
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	method := c.Request.Method

	// detect request
	if strings.ToLower(method) == "options" {
		originUrl := c.Request.Header.Get("Origin")

		// 查看origin是否来自于豁免清单里
		if originUrl != "" && cors.AllowedHostSet.Has(originUrl) {
			c.Header("Access-Control-Allow-Origin", originUrl)
		} else {
			c.Header("Access-Control-Allow-Origin", cors.AllowedHost)
		}

		c.Header("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		c.Header("Access-Control-Max-Age", "3600")
		c.Header("Content-Type", "application/json")

		// 响应探测请求
		c.JSON(http.StatusOK, "option request")
		return
	} else {
		c.Next()
	}
}
