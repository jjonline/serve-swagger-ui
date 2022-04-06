package route

import (
	"github.com/gin-gonic/gin"
	"github.com/jjonline/serve-swagger-ui/app/service"
	"github.com/jjonline/serve-swagger-ui/conf"
	"github.com/jjonline/serve-swagger-ui/render"
	"net/http"
)

// 中间件 -- 本质上是一个路由操作方法等同 gin.HandlerFunc

// notRoute 找不到路由输出
func notRoute(ctx *gin.Context) {
	ctx.JSON(http.StatusNotFound, render.H(http.StatusNotFound, http.StatusText(http.StatusNotFound), ""))
	return
}

// tryAuthenticate try to authenticate request and set safe flag value to ctx
func tryAuthenticate(ctx *gin.Context) {
	ctx.Set("token", service.OauthService.CheckAuthorization(ctx)) // set token anyway. Does not check for login logic
	ctx.Next()
}

// authenticate login status
func authenticate(ctx *gin.Context) {
	ctx.Header("Cache-Control", "public, max-age=1800")
	if conf.Config.ShouldLogin && !service.OauthService.CheckIsLoginUsingToken(ctx) {
		// need login, reset cookie then redirect to index page
		service.OauthService.DeleteCookie(ctx)
		ctx.Redirect(http.StatusFound, "/")
		ctx.Abort()
		return
	}
	ctx.Next()
}

// redirectIfAuthenticated login status should redirect to index, or public accessible
func redirectIfAuthenticated(ctx *gin.Context) {
	if !conf.Config.ShouldLogin || service.OauthService.CheckIsLoginUsingToken(ctx) {
		// login status auto redirect to index page
		ctx.Redirect(http.StatusFound, "/")
		ctx.Abort()
		return
	}

	ctx.Next()
}
