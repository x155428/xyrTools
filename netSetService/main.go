package main

import (
	"fmt"
	"github.com/kardianos/service"
	"log"
	"os"
	"time"
	"xyrTools/netSetService/setNet"
)

// 配置结构体
type program struct{}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *program) run() {
	setnet.SetNet()

	// 模拟守护进程
	for {
		time.Sleep(10 * time.Second)
	}
}

func (p *program) Stop(s service.Service) error {
	return nil
}

func main() {
	svcConfig := &service.Config{
		Name:        "xiaoyuNetSetService",
		DisplayName: "网卡配置服务",
		Description: "网卡配置需高权限，独立出来以服务形式完成网卡的配置.",
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}

	if len(os.Args) > 1 {
		cmd := os.Args[1]
		switch cmd {
		case "install":
			err = s.Install()
		case "uninstall":
			err = s.Uninstall()
		case "start":
			err = s.Start()
		case "stop":
			err = s.Stop()
		default:
			fmt.Println("Usage: install | uninstall | start | stop")
			return
		}

		if err != nil {
			log.Fatalf("Service command failed: %v", err)
		}

		fmt.Println("Command executed:", cmd)
		return
	}

	err = s.Run()
	if err != nil {
		log.Fatal(err)
	}
}
