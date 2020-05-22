/*
 * session: 具有一定时效性的加密字符串，通常服务端生成和解析
 * 一般包含字段：DeviceId, ClientType, Uid, timestamp
 * 1. 首先查看请求是否需要进行session验证，要么url标记，要么通过uri获取service判断服务是否登录验证
 * 2. 从url中提取或从cookie中提取token
 * 3. 校验token是否有效或过期，若是有效或者过期，重新生成一遍刷新用户token
 */
package webfilter

import (
	"com.dk.gateway/src/utils/webutils/token"
	"fmt"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
	"time"
)

type Session struct {
}

func NewSession() *Session {
	return &Session{
	}
}

var SessionInstance *Session

func init() {
	SessionInstance = NewSession()
}


func (s *Session) SessionCheck(c *gin.Context) {

	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	uri := c.Request.RequestURI
	if strings.Contains(uri, "transfer") {
		c.Next()
		return
	}

	var uid int64
	var session string

	cookies := c.Request.Cookies()
	if cookies != nil && len(cookies) > 0 {
		for _, cookie := range cookies {
			if "uid" == cookie.Name {
				uid, _ = strconv.ParseInt(cookie.Value, 10, 64)
			} else if "session" == cookie.Name {
				session = cookie.Value
			}
		}
	}

	ok := s.check(session, uid, c)
	c.Set("uid", uid)
	logId := s.genLogId(uid)

	logStr := fmt.Sprintf(" logid=%s ip=%s cmd=", logId, c.Request.RemoteAddr, c.Request.RequestURI)
	c.Set("logstr", logStr)
	c.Set("logid", logId)

	if (!ok) {
		c.JSONP(200, "session failed")
		return
	}

	newToken, err := token.Encode(&token.UserPayload{
		Uid:      uid,
		Ip:       c.Request.RemoteAddr,
		DeviceId: c.Param("device_id"),
	})

	if err != nil {
		c.SetCookie("PPU", newToken, 86400 * 30, "/", "gz-data.com", true, true)
	}
	c.Set("REQ_UID", uid)

	c.Next()
}

func (s *Session) check(session string, uid int64, c *gin.Context) bool  {
	if session == "" {
		return false
	}

	var tokenStr string
	tokenStr = c.Query("token")
	if tokenStr == "" {
		tokenStr = c.PostForm("token")
		if tokenStr == "" {
			cookie, err := c.Request.Cookie("token")
			if err == nil {
				tokenStr = cookie.Value
			}
		}
	}

	if tokenStr == "" {
		return false
	}

	_, expire, err := token.Decode(tokenStr)
	if err != nil {
		return false
	}

	if expire {
		//ToDo 调用用户逻辑层请求更新用户token
	}

	return true
}

func (s *Session) genLogId(param int64) int64 {
	ms := time.Now().Unix() * 1000
	return ms & 0x7FFFFFFF | (param >> 8 & 65535) << 47
}
