package main

import (
	"fmt"
	"path/filepath"

	//"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"xyrTools/xyrTools/core"
	"xyrTools/xyrTools/extendFunc"

	//"xyrTools/xyrTools/extendFunc"
	"xyrTools/xyrTools/extendFunc/log"
	initSys "xyrTools/xyrTools/init"

	"golang.org/x/sys/windows"

	"github.com/spf13/cast"
)

func main() {
	mutexName, _ := windows.UTF16PtrFromString("Global\\MyAppLock")
	mutex, err := windows.CreateMutex(nil, false, mutexName)
	if err != nil && err != windows.ERROR_ALREADY_EXISTS {
		panic("创建互斥体失败: " + err.Error())
	}
	defer windows.CloseHandle(mutex)

	// 检查是否已有实例运行
	if err == windows.ERROR_ALREADY_EXISTS {
		extendFunc.MessageBox("提示", "程序已在运行！")
		panic("程序已在运行！")
	}

	//获取路径
	configPath, err := filepath.Abs("./config/config.yaml")
	if err != nil {
		panic(err)
	}
	// 初始化日志器
	if err := logx.InitFromYaml(configPath); err != nil {
		panic(err)
	}

	// 简单打印函数，注入引擎
	logFunc := func(level, msg string) {
		fmt.Printf("[%s] %s\n", level, msg)
	}

	// 创建核心引擎
	engine := core.NewCoreEngine(logFunc)

	//初始化系统
	initSys.InitSys(engine, configPath)
	// 隐藏控制台窗口

	// 启动所有模块
	engine.StartAll()

	// 开启心跳检测
	cfgtmp := engine.GetConfig()
	heartbeatCfg := cfgtmp["heartbeat"]
	heartbeatCfgTmp, ok := heartbeatCfg.(map[interface{}]interface{})
	if !ok {
		fmt.Print("heartbeatCfg is not a map[interface{}]interface{}")
	}
	converted := convertMap(heartbeatCfgTmp)
	heartbeatAddr := cast.ToString(converted["heartbeatAddress"])
	isHeartbeat := cast.ToBool(converted["enabled"])
	if isHeartbeat {
		go engine.StartHeartBeatServer(heartbeatAddr)
	}

	// 捕捉退出信号
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	// 停止所有模块
	engine.StopAll()
	logx.Out("程序已退出", "info", false)
}

func convertMap(input map[interface{}]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range input {
		strKey := fmt.Sprintf("%v", k)
		result[strKey] = v
	}
	return result
}
