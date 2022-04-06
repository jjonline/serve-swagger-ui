package controller

import (
	"bytes"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jjonline/serve-swagger-ui/app/service"
	"github.com/jjonline/serve-swagger-ui/conf"
	"github.com/jjonline/serve-swagger-ui/render"
	"net/http"
	"os"
	"strings"
	"time"
)

// indexController index controller
type indexController struct{}

var (
	SwaggerPathNotFoundFile = errors.New("the swagger JSON file was not found at the specified path")
	SwaggerFileNotExist     = errors.New("the swagger file does not exist")
)

// Index action for list all swagger-JSON file list
func (s *indexController) Index(ctx *gin.Context) {
	data, err := service.ParseService.ParseWithCache()
	if err != nil {
		render.HtmlFail(ctx, SwaggerPathNotFoundFile)
		return
	}

	// assign values
	tokenInter, _ := ctx.Get("token")
	shares := map[string]interface{}{
		"siteName":        conf.Config.Server.SiteName, // siteName
		"items":           data.Items,                  // swagger file list
		"token":           tokenInter,                  // token structure
		"enableGoogle":    conf.Config.EnableGoogle,    // check if enable google oauth
		"enableMicrosoft": conf.Config.EnableMicrosoft, // check if enable microsoft oauth
		"random":          time.Now().Unix(),           // to redirect random param
	}
	ctx.HTML(http.StatusOK, "list.html", shares)
}

// Detail action for swagger-JSON file detail
func (s *indexController) Detail(ctx *gin.Context) {
	hash := strings.TrimRight(ctx.Param("path"), ".html")
	data, err := service.ParseService.ParseWithCache()
	if err != nil {
		render.HtmlFail(ctx, SwaggerPathNotFoundFile)
		return
	}
	if _, exist := data.Table[hash]; !exist {
		render.HtmlFail(ctx, SwaggerFileNotExist)
		return
	}

	// get swagger title
	detailTitle := conf.Config.Server.SiteName
	for _, items := range data.Items {
		for _, item := range items {
			if item.Hash == hash {
				detailTitle = item.Title + " - " + conf.Config.Server.SiteName
				break
			}
		}
	}

	share := map[string]string{
		"url":      "/json/" + hash + ".json",
		"siteName": detailTitle,
	}
	ctx.HTML(http.StatusOK, "detail.html", share)
}

// Json serve JSON file
func (s *indexController) Json(ctx *gin.Context) {
	hash := strings.TrimRight(ctx.Param("path"), ".json")
	data, err := service.ParseService.ParseWithCache()
	if err != nil {
		render.F(ctx, SwaggerPathNotFoundFile)
		return
	}
	path, exist := data.Table[hash]
	if !exist {
		render.F(ctx, SwaggerFileNotExist)
		return
	}
	stream, err := os.ReadFile(path)
	if err != nil {
		render.F(ctx, err)
		return
	}

	ctx.DataFromReader(http.StatusOK, int64(len(stream)), "application/json", bytes.NewReader(stream), nil)
}
