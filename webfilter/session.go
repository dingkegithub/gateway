/*
 * session: 具有一定时效性的加密字符串，通常服务端生成和解析
 * 一般包含字段：DeviceId, ClientType, Uid, timestamp
 * 0. 判定是否要session验证
 * 1. cookie 中的session和用户请求session是否一致, 不一致则返回，要求用户重新登录
 * 2. session 过期，则请求用户登录服务刷新用户session
 * 3. 生成请求唯一标志logid， logstr
 */
package webfilter

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dingkegithub/gateway/common"
	"github.com/dingkegithub/gateway/utils/netutils"
	"github.com/gin-gonic/gin"
)

type Session struct {
}

func NewSession() *Session {
	return &Session{}
}

var SessionInstance *Session

func init() {
	SessionInstance = NewSession()
}

func (s *Session) SessionCheck(c *gin.Context) {

	uri := c.Request.RequestURI
	if strings.Contains(uri, "nologin") {
		c.Next()
		return
	}

	var uid int64
	var session string

	cookies := c.Request.Cookies()
	// when login, response set
	// Set-Cookie:: uid=xxx
	// Set-Cookie: session=xxxxxxxxxxxx
	if cookies != nil && len(cookies) > 0 {
		for _, cookie := range cookies {
			if common.REQ_ID == cookie.Name {
				uid, _ = strconv.ParseInt(cookie.Value, 10, 64)
			} else if common.REQ_SESSION == cookie.Name {
				session = cookie.Value
			}
		}
	}

	if session == "" || uid < 0 {
		netutils.StdResponse(c, http.StatusUnauthorized, netutils.BuildResponseBody(403, "need login", nil))
		c.Abort()
		return
	}

	ok := s.check(session, uid, c)
	if !ok {
		netutils.StdResponse(c, http.StatusUnauthorized, netutils.BuildResponseBody(403, "login expire and relogin", nil))
		c.Abort()
		return
	}

	logId := s.genLogId(uid)
	c.Set(common.LOG_ID, logId)

	logStr := fmt.Sprintf("logid=%d ip=%s cmd=%s", logId, c.Request.RemoteAddr, c.Request.RequestURI)
	c.Set(common.LOG_STR, logStr)

	c.Next()
}

func (s *Session) check(session string, uid int64, c *gin.Context) bool {
	if session == "" {
		return false
	}

	var tokenStr string
	tokenStr = c.Query(common.REQ_TOKEN)
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

	if tokenStr == session {
		return true
	}

	// ToDo: session expire
	//       if session expire, gw need flush session by
	//       call user login service
	//func (c *Context) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	//c.SetCookie("uid", "1111111", time.Hour, "/",  "dgw.com", false, false)
	return false
}

func (s *Session) genLogId(param int64) int64 {
	ms := time.Now().Unix() * 1000
	return ms&0x7FFFFFFF | (param>>8&65535)<<47
}
