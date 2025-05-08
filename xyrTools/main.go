package main

import (
	"fmt"
	//"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"xyrTools/xyrTools/core"
	"xyrTools/xyrTools/extendFunc"
	initSys "xyrTools/xyrTools/init"
)

func main() {
	// 检查是否已存在锁文件
	if extendFunc.CheckLockFile() {
		extendFunc.MessageBox("提示", "程序已在运行！")
		return
	}

	// 创建锁文件
	err := extendFunc.CreateLockFile()
	if err != nil {
		fmt.Printf("创建锁文件失败: %v\n", err)
		return
	}
	// 确保退出时删除锁文件
	defer extendFunc.RemoveLockFile()
	// 日志函数
	logFunc := func(level, msg string) {
		fmt.Printf("[%s] %s\n", level, msg)
	}

	// 创建核心引擎
	engine := core.NewCoreEngine(logFunc)
	//获取当前项目的路径
	configPath, err := os.Getwd()
	if err != nil {
		logFunc("fatal", err.Error())
		return
	}
	// 路径加相对路径拼接出配置文件路径
	configPath = configPath + "/config/config.yaml"

	//初始化系统
	initSys.InitSys(engine, configPath)
	// 隐藏控制台窗口

	// 启动所有模块
	engine.StartAll()

	// 捕捉退出信号
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	// 停止所有模块
	engine.StopAll()

	logFunc("info", "程序已退出")
}
