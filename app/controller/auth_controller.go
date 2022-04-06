package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/jjonline/serve-swagger-ui/app/service"
	"github.com/jjonline/serve-swagger-ui/render"
	"net/http"
)

// authController auth controller
type authController struct{}

// LoginUsingGoogle to google oAuth login
func (s *authController) LoginUsingGoogle(ctx *gin.Context) {
	ctx.Redirect(http.StatusFound, service.OauthService.GoogleRedirectURL(ctx))
}

// CallbackUsingGoogle google oAuth login callback
func (s *authController) CallbackUsingGoogle(ctx *gin.Context) {
	if err := service.OauthService.GoogleCallback(ctx); err != nil {
		render.HtmlFail(ctx, err)
		return
	}
	ctx.Redirect(http.StatusFound, "/")
}

// LoginUsingMicrosoft to Microsoft oAuth login
func (s *authController) LoginUsingMicrosoft(ctx *gin.Context) {
	ctx.Redirect(http.StatusFound, service.OauthService.MicrosoftRedirectURL(ctx))
}

// CallbackUsingMicrosoft Microsoft oAuth login callback
func (s *authController) CallbackUsingMicrosoft(ctx *gin.Context) {
	if err := service.OauthService.MicrosoftCallback(ctx); err != nil {
		render.HtmlFail(ctx, err)
		return
	}
	ctx.Redirect(http.StatusFound, "/")
}

// Logout exit logout
func (s *authController) Logout(ctx *gin.Context) {
	service.OauthService.DeleteCookie(ctx)
	ctx.Redirect(http.StatusFound, "/")
}
