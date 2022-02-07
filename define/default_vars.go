package define

// 定义一些全局的常量、变量默认值

// 定义一些导出的全局默认值变量
//  ① 项目内所有默认变量定义在此处，便于对默认值进行统一管理，IDE也可以快速追踪何处使用了默认值
//  ② 尽量以Default作为变量名前缀
var (
	Version         = "0.0.1"     // framework version
	DefaultHost     = "0.0.0.0"   // default host
	DefaultPort     = 9080        // default port
	DefaultConfig   = "conf.toml" // default config file path
	DefaultLogPath  = "stdout"    // default log path
	DefaultLogLevel = "debug"     // default log level
)
