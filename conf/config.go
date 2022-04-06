package conf

import (
	"github.com/jjonline/serve-swagger-ui/define"
	"github.com/jjonline/serve-swagger-ui/utils/cfg"
	"net/url"
	"strings"
)

// 项目config配置定义

// 暴露给全局使用的配置变量
var (
	Config config
	Cmd    cmdConfig
)

// config 项目配置上层结构
type config struct {
	Server          server    `json:"server"`    // server config
	Log             log       `json:"log"`       // log config
	Google          google    `json:"google"`    // google oauth config
	Microsoft       microsoft `json:"microsoft"` // microsoft oauth config
	Swagger         swagger   `json:"swagger"`   // swagger json file config
	Account         account   `json:"account"`   // google account set
	ConfigFile      string    `json:"-"`         // record config file path
	ShouldLogin     bool      `json:"-"`         // record should log in at first
	EnableGoogle    bool      `json:"-"`         // record is set google enable
	EnableMicrosoft bool      `json:"-"`         // record is set microsoft enable
}

// cmdConfig command line args
type cmdConfig struct {
	ConfigFile  string // set config file path
	Host        string // set web server host ip
	Port        int    // set web server ports
	SwaggerPath string // set swagger file path
	LogLevel    string // set logger level
	LogPath     string // set logger path
	OpenBrowser bool   // set if auto open browser
}

// parseAfterLoad Unified processing flow logic after the configuration item is loaded
func (c config) parseAfterLoad() {
	// if google or microsoft oauth is enabled, check needed JwtKey
	if Config.EnableGoogle || Config.EnableMicrosoft {
		if Config.Server.BaseURL == "" {
			panic("Enable authentication must be set Server.BaseURL")
		}
		// try parse baseURL is valid
		if _, err := url.Parse(Config.Server.BaseURL); err != nil {
			panic("Server.BaseURL value is invalid")
		}
		if Config.Server.JwtKey == "" || Config.Server.JwtExpiredTime <= 0 {
			panic("Enable authentication must be set Server.JwtKey and Server.JwtExpiredTime")
		}
	}

	// parse BaseURL suffix slash, add corrected slash
	// Make sure the config ends up being slashed with or without a trailing slash
	if Config.Server.BaseURL != "" {
		Config.Server.BaseURL = strings.TrimRight(Config.Server.BaseURL, "/") + "/"
	}

}

// region 初始化

// Init 初始化
func Init() {
	// ① set framework version
	Config.Server.Version = define.Version

	// set config file path
	if Cmd.ConfigFile == "" {
		Cmd.ConfigFile = define.DefaultConfig
	}
	Config.ConfigFile = Cmd.ConfigFile

	// read config always
	var cfgLoader cfg.IFace
	cfgLoader = cfg.Viper{}
	if err := cfgLoader.Parse(Cmd.ConfigFile, "toml", &Config); err != nil {
		// do not set config file params, reset to empty value
		Config.ConfigFile = ""
	}

	// ② command line args first
	if Cmd.Host != "" {
		Config.Server.Host = Cmd.Host
	}
	if Cmd.Port != 0 {
		Config.Server.Port = Cmd.Port
	}
	if Cmd.SwaggerPath != "" {
		Config.Swagger.Path = Cmd.SwaggerPath
	}
	if Cmd.LogPath != "" {
		Config.Log.Path = Cmd.LogPath
	}
	if Cmd.LogLevel != "" {
		Config.Log.Level = Cmd.LogLevel
	}

	// ③ set default value
	if Config.Server.SiteName == "" {
		Config.Server.SiteName = define.DefaultSiteName
	}
	if Config.Server.Host == "" {
		Config.Server.Host = define.DefaultHost
	}
	if Config.Server.Port == 0 {
		Config.Server.Port = define.DefaultPort
	}
	if Config.Log.Path == "" {
		Config.Log.Path = define.DefaultLogPath
	}
	if Config.Log.Level == "" {
		Config.Log.Level = define.DefaultLogLevel
	}
	if Config.Swagger.Path == "" {
		Config.Swagger.Path = define.DefaultPath
	}

	// ④ set EnableGoogle value
	if Config.Google.ClientID != "" && Config.Google.ClientSecret != "" {
		Config.ShouldLogin = true
		Config.EnableGoogle = true
	}

	// ⑤ set EnableMicrosoft value
	if Config.Microsoft.ClientID != "" && Config.Microsoft.ClientSecret != "" && Config.Microsoft.Tenant != "" {
		Config.ShouldLogin = true
		Config.EnableMicrosoft = true
	}

	// 配置加载并解析映射成功后统一处理逻辑：譬如Url统一处理后缀斜杠
	Config.parseAfterLoad()
}

// endregion

// region 热重载

// Reload 热重载
// 当配置变更时又无需重新启动进程触发，监听调用
func Reload() {

}

// endregion
