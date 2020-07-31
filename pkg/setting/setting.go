package setting

import (
	"log"
	"time"

	"github.com/go-ini/ini"
)

type App struct {
	JwtSecret string
	PageSize  int
	PrefixUrl string

	RuntimeRootPath string

	ImageSavePath  string
	ImageMaxSize   int
	ImageAllowExts []string

	ExportSavePath string
	QrCodeSavePath string
	FontSavePath   string

	LogSavePath string
	LogSaveName string
	LogFileExt  string
	TimeFormat  string
}

var AppPath string //程序的绝对路径

var AppSetting = &App{}


type Server struct {
	RunMode      string
	HttpPort     int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

var ServerSetting = &Server{}

type Database struct {
	Type        string
	User        string
	Password    string
	Host        string
	Db_Name        string
	Db_Charset string
	Db_Port int
	TablePrefix string
	BackupFlag int
	WorkDir string
}

var DatabaseSetting = &Database{}

type BackupSpec struct {
	Spec string
}

var BackupSetting = &BackupSpec{}

type Redis struct {
	Host        string
	Password    string
	MaxIdle     int
	MaxActive   int
	IdleTimeout time.Duration
}

var RedisSetting = &Redis{}

var cfg *ini.File

// Setup initialize the configuration instance
func Setup() {
	var err error
	cfg, err = ini.Load("conf/app.ini")
	if err != nil {
		log.Fatalf("setting.Setup, fail to parse 'conf/app.ini': %v", err)
	}

	mapTo("app", AppSetting)
	mapTo("server", ServerSetting)
	mapTo("database", DatabaseSetting)
	//mapTo("spec", BackupSetting)
	mapTo("redis", RedisSetting)

	BackupSetting.Spec= cfg.Section("spec").Key("spec").String()


/*
	AppSetting.PageSize=10
	AppSetting.JwtSecret="233"
	AppSetting.PrefixUrl="http://127.0.0.1:8887"
	AppSetting.RuntimeRootPath="runtime/"
	AppSetting.ImageSavePath="upload/images/"
	AppSetting.ImageMaxSize=5
	AppSetting.ImageAllowExts=[]string{".jpg",".jpeg",".png"}
	AppSetting.ExportSavePath="export/"
	AppSetting.QrCodeSavePath = "qrcode/"
	AppSetting.FontSavePath = "fonts/"
	AppSetting.LogSavePath = "logs/"
	AppSetting.LogSaveName = "log"
	AppSetting.LogFileExt = "log"
	AppSetting.TimeFormat = "20060102"

 */
	AppSetting.ImageMaxSize = AppSetting.ImageMaxSize * 1024 * 1024

/*
		ServerSetting.RunMode = "release"
		ServerSetting.HttpPort = 8887
		ServerSetting.ReadTimeout=60
		ServerSetting.ReadTimeout=60

 */

	ServerSetting.ReadTimeout = ServerSetting.ReadTimeout * time.Second
	ServerSetting.WriteTimeout = ServerSetting.WriteTimeout * time.Second
	RedisSetting.IdleTimeout = RedisSetting.IdleTimeout * time.Second
}

// mapTo map section
func mapTo(section string, v interface{}) {
	err := cfg.Section(section).MapTo(v)
	if err != nil {
		log.Fatalf("Cfg.MapTo %s err: %v", section, err)
	}
}
