package initializer

import (
	"github.com/jjonline/go-lib-backend/logger"
	"github.com/jjonline/serve-swagger-ui/conf"
)

//go:noinline
func iniLogger() *logger.Logger {
	return logger.New(conf.Config.Log.Level, conf.Config.Log.Path, "module")
}
