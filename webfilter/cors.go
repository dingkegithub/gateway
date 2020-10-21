package webfilter

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/dingkegithub/gateway/common/logging"
	"github.com/dingkegithub/gateway/config"
	"github.com/dingkegithub/gateway/utils/netutils"
	"github.com/fatih/set"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	corsAddr string = "http://dgw.com"
)

type CorsCfg struct {
	MaxAge           uint64   `json:"max_age"`
	AllowAllOrigin   bool     `json:"allow_all_origin"`
	AllowCredentials bool     `json:"allow_credentials"`
	ShortCircuit     bool     `json:"short_circuit"`
	AllowMethods     []string `json:"allow_methods"`
	AllowHeaders     []string `json:"allow_headers"`
	ExposeHeaders    []string `json:"expose_headers"`
	Whitelist        []string `json:"whitelist"`
}

var CorsInstance *Cors
var once sync.Once

type Cors struct {
	mutex     sync.RWMutex
	cfgKey    string
	cfg       *CorsCfg
	cfgChan   chan interface{}
	whitelist set.Interface
	cfgCenter *config.HotCfg
}

func NewCors() *Cors {
	once.Do(func() {
		var cfg *CorsCfg
		whitelist := set.New(set.ThreadSafe)

		h := config.Instance()
		cfgStr := h.Get("cors", "")

		err := json.Unmarshal([]byte(cfgStr), &cfg)
		if err != nil {
			logging.Error("marshal cfg string err", zap.String("cfg", cfgStr), zap.Error(err))

			cfg = &CorsCfg{
				MaxAge:           120,
				AllowCredentials: true,
				ShortCircuit:     true,
				ExposeHeaders:    []string{},
				Whitelist:        []string{"null", corsAddr},
				AllowMethods:     []string{"GET", "PUT"},
				AllowHeaders:     []string{},
			}
			whitelist.Add("null")
			whitelist.Add(corsAddr)
		} else {
			for _, v := range cfg.Whitelist {
				whitelist.Add(v)
			}
		}

		CorsInstance = &Cors{
			cfg:       cfg,
			cfgKey:    "cors",
			cfgChan:   make(chan interface{}),
			whitelist: whitelist,
			cfgCenter: h,
		}
		CorsInstance.cfgCenter.Register(CorsInstance.cfgKey, CorsInstance.cfgChan)
		go CorsInstance.listener()
	})

	CorsInstance.whitelist.Each(func(item interface{}) bool {
		logging.Debug("whitelist info", zap.String("name", item.(string)))
		return true
	})

	return CorsInstance
}

func (c *Cors) listener() {
	for {
		<-c.cfgChan
		logging.Info("new cors config notify")

		cfgStr := c.cfgCenter.Get("cors", "")
		logging.Debug("notify get config", zap.String("cfg", cfgStr))

		var cfg *CorsCfg

		err := json.Unmarshal([]byte(cfgStr), &cfg)
		if err != nil {
			logging.Error("json parse string failed", zap.String("cfg", cfgStr), zap.Error(err))
			return
		}

		whitelist := set.New(set.ThreadSafe)
		for _, v := range cfg.Whitelist {
			logging.Debug("whitelist", zap.String("w", v))
			whitelist.Add(v)
		}

		c.mutex.Lock()
		c.cfg = cfg
		c.whitelist = whitelist
		c.mutex.Unlock()
	}
}

func (c *Cors) canCrossOrigin(origin string) bool {
	if len(origin) == 0 || origin == "null" {
		return true
	}
	return c.whitelist.Has(origin)
}

func (c *Cors) isSameHost(ctx *gin.Context) bool {
	host := strings.Split(ctx.Request.Host, ":")[0]
	remote := strings.Split(ctx.Request.RemoteAddr, ":")[0]

	return host == remote
}

func (c *Cors) isPreflight(ctx *gin.Context) bool {
	isOptionReq := strings.ToUpper(ctx.Request.Method) == http.MethodOptions
	hasOrigin := len(ctx.Request.Header.Get("Origin")) > 0
	hasAccessCtrReqMethod := len(ctx.Request.Header.Get("Access-Control-Request-Method")) > 0
	return isOptionReq && hasOrigin && hasAccessCtrReqMethod
}

func (c *Cors) isSameOrigin(ctx *gin.Context) bool {
	host := fmt.Sprintf("%s://%s", ctx.Request.Proto, ctx.Request.Host)
	origin := ctx.Request.Header.Get("origin")

	return host == origin || len(origin) > 0
}

func (c *Cors) CorsCheck(ctx *gin.Context) {

	method := ctx.Request.Method
	origin := ctx.Request.Header.Get("Origin")
	accessCtrReqMethod := ctx.Request.Header.Get("Access-Control-Request-Method")

	logging.Info("request header origin",
		zap.String("method", method),
		zap.String("origin", origin),
		zap.String("proto", ctx.Request.Proto),
		zap.String("host", ctx.Request.Host),
		zap.String("access-control-request-method", accessCtrReqMethod))

	// 允许垮cookie访问
	if c.cfg.AllowCredentials {
		ctx.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	}

	if c.cfg.AllowAllOrigin {
		ctx.Header("Access-Control-Allow-Origin", "*")
	} else {
		if c.canCrossOrigin(origin) {
			ctx.Header("Access-Control-Allow-Origin", origin)
		} else if c.cfg.ShortCircuit {
			netutils.StdResponse(ctx, http.StatusForbidden, netutils.BuildResponseBody(403, "short circuit invalid origin", nil))
			ctx.Abort()
			return
		} else {
			ctx.Header("Access-Control-Allow-Origin", corsAddr)
		}
		ctx.Header("Vary", "Origin")
	}

	if netutils.IsPreflight(ctx) {
		ctx.Header("Access-Control-Allow-Methods", strings.Join(c.cfg.AllowMethods, ","))

		if len(c.cfg.AllowHeaders) > 0 {
			ctx.Header("Access-Control-Allow-Headers", strings.Join(c.cfg.AllowHeaders, ","))
		}

		if c.cfg.MaxAge > 0 {
			ctx.Header("Access-Control-Max-Age", fmt.Sprintf("%d", c.cfg.MaxAge))
		} else {
			ctx.Header("Access-Control-Max-Age", fmt.Sprintf("%d", 120))
		}

		ctx.Data(http.StatusNoContent, "", nil)
		ctx.Abort()
		return
	} else {
		if len(c.cfg.ExposeHeaders) > 0 {
			ctx.Header("Access-Control-Expose-Headers", strings.Join(c.cfg.ExposeHeaders, ","))
		}
	}

	logging.Info("no need cors check", zap.String("method", method))
	ctx.Next()
}
