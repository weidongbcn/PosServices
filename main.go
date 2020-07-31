package main

import (
	"fmt"
	"github.com/kardianos/service"
	"log"
	"os"
	"path/filepath"
	"github.com/weidongbcn/PosServices/pkg/setting"
)

//var appPath string // 程序的绝对路径  =

func init()  {
	//setting.Setup()
	//logging.Setup()
	var err error
	if setting.AppPath, err = GetAppPath(); err != nil {
		log.Fatal(err)
	}
	if err := os.Chdir(filepath.Dir(setting.AppPath)); err != nil {
		log.Fatal(err)
	}
}

func main() {
	svcConfig := &service.Config{
		Name:        "PosServices3",
		DisplayName: "Go Service for POS",
		Description: "This is a TPV Go service.",
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		fmt.Printf("Failed to initialize service: %s \n", err)
	}

	var verb string
	if len(os.Args) > 1 {
		verb = os.Args[1]
	}

	var executionError error
	switch verb {
	case "install":
		executionError = s.Install()
		if executionError != nil {
			fmt.Printf("Failed to install: %s\n", executionError)
			return
		}
		fmt.Printf("Service \"%s\" installed.\n", svcConfig.DisplayName)
	case "uninstall":
		executionError = s.Uninstall()
		if executionError != nil {
			fmt.Printf("Failed to remove: %s\n", executionError)
			return
		}
		fmt.Printf("Service \"%s\" removed.\n", svcConfig.DisplayName)
	case "start":
		executionError = s.Start()
		if executionError != nil {
			fmt.Printf("Failed to start \"%s\" : %s\n", svcConfig.DisplayName, executionError)
			return
		}
		fmt.Printf("Service \"%s\" started. %s \n", svcConfig.DisplayName, executionError)
	case "stop":
		executionError = s.Stop()
		if err != nil {
			fmt.Printf("Failed to stop: %s\n", executionError)
			return
		}
		fmt.Printf("Service \"%s\" stopped.\n", svcConfig.DisplayName)
	default:
		executionError = s.Run()
		if err != nil {
			fmt.Printf("Failed to run: %s\n", executionError)
			return
		}
	}
}