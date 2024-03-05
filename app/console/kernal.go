package console

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jjonline/serve-swagger-ui/app/service"
	"github.com/jjonline/serve-swagger-ui/client"
	"github.com/jjonline/serve-swagger-ui/conf"
	"github.com/jjonline/serve-swagger-ui/route"
	"github.com/toqueteos/webbrowser"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

// region Service Boot Launcher

// BootStrap boot
func BootStrap() {
	// output base info
	client.Logger.Info(fmt.Sprintf("Golang Version      : %s", runtime.Version()))
	client.Logger.Info(fmt.Sprintf("MAX Cpu Num         : %d", runtime.GOMAXPROCS(-1)))
	client.Logger.Info(fmt.Sprintf("Command Args        : %s", os.Args))
	client.Logger.Info(fmt.Sprintf("Config file Path    : %s", conf.Config.ConfigFile))
	client.Logger.Info(fmt.Sprintf("Enable Google login : %t", conf.Config.EnableGoogle))
	client.Logger.Info(fmt.Sprintf("Server HOST         : %s", conf.Config.Server.Host))
	client.Logger.Info(fmt.Sprintf("Server PORT         : %d", conf.Config.Server.Port))
	client.Logger.Info(fmt.Sprintf("Swagger JSON Path   : %s", conf.Config.Swagger.Path))
	client.Logger.Info(fmt.Sprintf("Log Path            : %s", conf.Config.Log.Path))
	client.Logger.Info(fmt.Sprintf("Log Level           : %s", conf.Config.Log.Level))

	_, signalChan := quitCtx()

	startHTTPApp(signalChan)
}

// endregion

// region HttpServer启动

// startHTTPApp http api服务启动
func startHTTPApp(signalChan chan os.Signal) {
	// Set gin startup mode
	gin.SetMode(gin.ReleaseMode)

	server := &http.Server{
		Addr:           fmt.Sprintf("%s:%d", conf.Config.Server.Host, conf.Config.Server.Port),
		Handler:        route.Bootstrap(),
		ReadTimeout:    time.Duration(conf.Config.Server.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(conf.Config.Server.WriteTimeout) * time.Second,
		MaxHeaderBytes: 1 << 20, //1MB
	}

	// The main process blocks the channel
	idleCloser := make(chan struct{})

	// start file change watcher
	go service.ParseService.StartFileWatcher()

	// http serv handle exit signal
	go func() {
		afterStartHTTPOk() // after http start ok, then do something
		<-signalChan

		// 超时context
		timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer timeoutCancel()
		// We received an interrupt signal, shut down.
		if err := server.Shutdown(timeoutCtx); err != nil {
			// Error from closing listeners, or context timeout:
			client.Logger.Error("Http service violent stop：" + err.Error())
		} else {
			// time.Sleep(1 * time.Second)
			// successful shutdown process ok
			client.Logger.Info("Http service stops gracefully")
		}

		// closer quit chan
		close(idleCloser)
	}()

	// continue serv http service
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		client.Logger.Error("Http service exception：" + err.Error())
		close(idleCloser)
	}

	// wait for stop main process
	<-idleCloser
	client.Logger.Info("Process exited: The service was shut down")
}

func afterStartHTTPOk() {
	// wait server error if http start failed
	time.Sleep(300 * time.Millisecond)

	// display serve URL
	fmt.Printf("serve-swagger-ui at: http://%s:%d\n", conf.Config.Server.Host, conf.Config.Server.Port)

	// if --open is true, open first doc auto
	if conf.Cmd.OpenBrowser {
		hash, err := service.ParseService.FirstDoc()
		if err != nil {
			panic(err)
		}
		visit := fmt.Sprintf("http://%s:%d/doc/%s.html", conf.Config.Server.Host, conf.Config.Server.Port, hash)
		if err := webbrowser.Open(visit); err != nil {
			fmt.Printf("Failed to open the browser automatically, please open it manually: %s", visit)
		}
	}
}

// endregion

// region Global process control signal capture

// quitCtx global exit signal
func quitCtx() (context.Context, chan os.Signal) {
	ctx, cancel := context.WithCancel(context.Background())
	quitChan := make(chan os.Signal)
	signal.Notify(quitChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)

	go func() {
		<-quitChan // wait quit signal

		signal.Stop(quitChan)
		cancel()
		close(quitChan)
	}()

	return ctx, quitChan
}

// endregion
