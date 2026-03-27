package setnet

import (
	"encoding/json"
	"log"
	"os"

	"xyrTools/netSetService/config"
	"xyrTools/netSetService/handledata"

	"myMod/notify"

	"github.com/Microsoft/go-winio"
)

func SetNet() {
	dir, _ := os.Getwd()
	logFile, err := os.OpenFile(dir+"/xiaoyulog.txt", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		notify.NotifyError(err, "打开日志文件失败")
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	// 管道描述符
	securityDescriptor := "D:P(A;;GA;;;S-1-5-32-544)(A;;GRGW;;;S-1-5-32-545)"
	// 配置命名管道
	netCfgPipeCfg := &winio.PipeConfig{
		SecurityDescriptor: securityDescriptor,
	}
	pipePath := `\\.\pipe\netCfgPipe`
	ln, err := winio.ListenPipe(pipePath, netCfgPipeCfg)
	if err != nil {
		log.Println("Error listening on pipe:", err)
		os.Exit(1)
	}
	defer ln.Close()
	log.Println("Waiting for connection on", pipePath)
	// 循环等待连接
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}
		// 读取数据
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			log.Println("Error reading from connection:", err)
			continue
		}
		log.Printf("Received %d bytes: %s\n", n, string(buf[:n]))
		// 解析数据
		unpackateData, isUnpackage, unpackageErr := handledata.UnpackageData(buf[:n], "#")
		if !isUnpackage {
			log.Println("Error unpackage data:", unpackageErr)
			continue
		}
		log.Printf("Data: %s\n", unpackateData)
		result := config.ParseConfigAndConfigure(unpackateData)
		// 将结果序列化为 JSON 并发送回客户端
		resultJSON, err := json.Marshal(result)
		if err != nil {
			log.Println("Error marshalling result:", err)
			continue
		}
		//打包结果
		resultStr, packageErr := handledata.PackageData(string(resultJSON), "#")
		if packageErr != nil {
			log.Println("Error package data:", packageErr)
			continue
		}
		_, err = conn.Write(resultStr)
		if err != nil {
			log.Println("Error writing result to connection:", err)
			continue
		}

	}
}
