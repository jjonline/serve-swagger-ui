package route

import (
	"github.com/gin-gonic/gin"
	"github.com/jjonline/go-lib-backend/logger"
	"github.com/jjonline/serve-swagger-ui/app/service"
	"github.com/jjonline/serve-swagger-ui/conf"
	"go.uber.org/zap"
)

// router 包内路由变量，请勿覆盖
//  - 一般扩展路由是基于该变量链式添加，为了识别可将固定前缀的路由拆分文件
var router *gin.Engine

// iniRoute 路由init-logger、recovery、cors 等
func iniRoute() {
	router = gin.New()

	// set base middleware
	router.Use(logger.GinLogger(appendEmailIfExist), logger.GinRecovery)
	if conf.Config.Server.Cors {
		router.Use(logger.GinCors)
	}

	// 请求找不到路由时输出错误
	router.NoRoute(notRoute)
}

// appendEmailIfExist append email field to logger if exist
func appendEmailIfExist(ctx *gin.Context) []zap.Field {
	if tokenInter, exist := ctx.Get("token"); exist {
		if token, ok := tokenInter.(service.Token); ok && token.Authenticated {
			filed := make([]zap.Field, 0)
			return append(filed, zap.String("email", token.Email))
		}
	}
	return nil
}

// Bootstrap 引导初始化路由route
func Bootstrap() *gin.Engine {
	iniRoute()
	routeSetting()
	return router
}
