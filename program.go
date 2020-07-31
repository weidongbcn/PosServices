package main

import (
	"database/sql"
	"fmt"
	"github.com/JamesStewy/go-mysqldump"
	"github.com/gin-gonic/gin"
	"github.com/kardianos/service"
	"github.com/robfig/cron"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"github.com/weidongbcn/PosServices/pkg/e"
	"github.com/weidongbcn/PosServices/pkg/file"
	"github.com/weidongbcn/PosServices/pkg/logging"
	"github.com/weidongbcn/PosServices/pkg/setting"
	"github.com/weidongbcn/PosServices/pkg/upload"

	"time"
	_ "github.com/go-sql-driver/mysql"
)

var logger service.Logger

type program struct {

}

func GetAppPath() (string, error) {
	prog := os.Args[0]
	p, err := filepath.Abs(prog)
	if err != nil {
		return "", err
	}
	fi, err := os.Stat(p)
	if err == nil {
		if !fi.Mode().IsDir() {
			return p, nil
		}
		err = fmt.Errorf("winsvc.GetAppPath: %s is directory", p)
	}
	if filepath.Ext(p) == "" {
		p += ".exe"
		fi, err := os.Stat(p)
		if err == nil {
			if !fi.Mode().IsDir() {
				return p, nil
			}
			err = fmt.Errorf("winsvc.GetAppPath: %s is directory", p)
		}
	}
	return "", err
}


func (p *program) Start(s service.Service) error {
	// Start should not block. Do the actual work async.
	go p.run()
	return nil
}

func (p *program) run() {

	setting.Setup()
	logging.Setup()
	logging.Info("设置完毕")
	go GinServer()
	//DoMysqlBackup2()
	fmt.Printf("开始定时任务 \n")

	logging.Info("开始定时任务")

	c:= cron.New(cron.WithSeconds())
	c.AddFunc( setting.BackupSetting.Spec, func() {
		logging.Info("开始备份数据库")
		DoMysqlBackup2()
	})

	c.Start()

	t1 := time.NewTimer(time.Second * 10)
	for {
		select {
		case <-t1.C:
			t1.Reset(time.Second * 10)
		}
	}
	/*

	crontab := cron.New()
	crontab.AddFunc("*3 * * * * *", func() {
		fmt.Println("every 3 seconds executing")
	})
	//crontab.AddFunc("0 20 8 * * *", DoMysqlBackup)  // 每天 8:20:00 定时执行 myfunc 函数
	crontab.AddFunc( setting.BackupSetting.Spec, func() {
		logging.Info("开始备份数据库")
		DoMysqlBackup(setting.DatabaseSetting.BackupFlag)
	})
	crontab.Start()
	defer crontab.Stop()
//	select {}  //阻塞主线程停止
i:=0

	for {
		time.Sleep(time.Duration(1) * time.Second)

		i=i+1
		fmt.Printf("%d", i)
		if i == 5 {
			fmt.Printf("开始备份 \n")
		//go	DoMysqlBackup(setting.DatabaseSetting.BackupFlag)
		}
	}

	*/



	//logging.Setup()

}

func (p *program) Stop(s service.Service) error {
	// Stop should not block. Return with a few seconds.
	return nil
}


func GinServer()  {
	logFile, err := os.Create("gin.log")
	if err != nil {
		panic(err)
	}
	gin.DefaultWriter = io.MultiWriter(logFile, os.Stdout)
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.String(200, "您好, 工作一切正常.")
	})
		r.StaticFS("/upload/images", http.Dir(upload.GetImageFullPath()))
		r.POST("/upload", UploadImage)
	r.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"msg":  "测试成功",
			"data": nil,
		})
	})



	r.Run(":8887") // listen and serve on 0.0.0.0:8080


}


func UploadImage(c *gin.Context) {
	code := e.SUCCESS
	data := make(map[string]string)
	file, image, err := c.Request.FormFile("image")
	if err != nil {
		logging.Warn(err)
		//logger.Error(err)
		code = e.ERROR
		c.JSON(http.StatusOK, gin.H{
			"code": code,
			"msg":  e.GetMsg(code),
			"data": data,
		})
	}

	if image == nil {
		code = e.INVALID_PARAMS
	} else {
		imageName := upload.GetImageName(image.Filename)
		fullPath := upload.GetImageFullPath()
		savePath := upload.GetImagePath()

		src := fullPath + imageName
		if ! upload.CheckImageExt(imageName) || ! upload.CheckImageSize(file) {
			code = e.ERROR_UPLOAD_CHECK_IMAGE_FORMAT
		//	logger.Error(e.GetMsg(code))
		logging.Error(e.GetMsg(code))
		} else {
			err := upload.CheckImage(fullPath)
			if err != nil {
				logging.Warn(err)
				code = e.ERROR_UPLOAD_CHECK_IMAGE_FAIL
			} else if err := c.SaveUploadedFile(image, src); err != nil {
			//	logger.Warning(err)
			logging.Warn(err)
				code = e.ERROR_UPLOAD_SAVE_IMAGE_FAIL
			} else {
			logging.Info("图片上传成功")
				data["image_url"] = upload.GetImageFullUrl(imageName)
				data["image_save_url"] = savePath + imageName
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": code,
		"msg":  e.GetMsg(code),
		"data": data,
	})
}

func DoMysqlBackup2()  {
	logging.Info("正再准备开始备份的前行条件" )
	username := setting.DatabaseSetting.User
	password := setting.DatabaseSetting.Password
	hostname := setting.DatabaseSetting.Host
	port := strconv.Itoa(setting.DatabaseSetting.Db_Port)
	dbname := setting.DatabaseSetting.Db_Name

	fmt.Println(fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, hostname, port, dbname))
	fmt.Println("-----------------")

	dumpDir := "MysqlBackup"  // you should create this directory
	dumpFilenameFormat := fmt.Sprintf("%s-20060102T150405", dbname)   // accepts time layout string and add .sql at the end of file

	F, err := file.MustOpen(dumpFilenameFormat, dumpDir)
	if err != nil {
		log.Fatalf("Backup.Setup err: %v", err)
		logging.Fatal("备份目录创建失败 err: %v", err )
	}
	defer F.Close()



	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, hostname, port, dbname))
	if err != nil {
		fmt.Println("Error opening database: ", err)
		logging.Fatal("数据打开失败 err: %v", err )
		return
	}

	// Register database with mysqldump
	dumper, err := mysqldump.Register(db, dumpDir, dumpFilenameFormat)
	if err != nil {
		fmt.Println("Error registering databse:", err)
		logging.Fatal("数据订记失败 err: %v", err )
		return
	}

	// Dump database to file
	resultFilename, err := dumper.Dump()
	if err != nil {
		fmt.Println("Error dumping:", err)
		logging.Fatal("数据备份失败 err: %v", err )
		return
	}
	fmt.Printf("File is saved to %s", resultFilename)
	logging.Info("数据成功备份")
	logging.Info("备份保存在 %s", resultFilename)

	// Close dumper and connected database
	dumper.Close()

}


//cron 备份速度太慢, 抛弃了
/*
func cronInit() {

	crontab := cron.New(cron.WithSeconds())
	//crontab.AddFunc("0 20 8 * * *", DoMysqlBackup)  // 每天 8:20:00 定时执行 myfunc 函数
	crontab.AddFunc( setting.BackupSetting.Spec, func() {
		logging.Info("开始备份数据库")
		DoMysqlBackup(setting.DatabaseSetting.BackupFlag)
	})
	crontab.Start()
	defer crontab.Stop()
	select {

	}
}

 */