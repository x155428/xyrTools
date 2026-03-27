package fileMonitor

import (
	"fmt"
	"myMod/notify"
	"sync"
	"xyrTools/xyrTools/modInterfaces"
)

type FileMonitorModule struct {
	status         modInterfaces.ModuleStatus // 模块状态
	ctx            modInterfaces.Context      // 模块上下文
	stopCh         chan struct{}              // 停止信号通道，stop时通知模块退出
	wg             sync.WaitGroup             // 等待组、确保模块退出时所有 goroutine 都已退出
	suffixChecker  *SuffixChecker
	behaviorCache  *BehaviorCache
	alertGenerator AlertGenerator
}

func New() modInterfaces.Module {
	return &FileMonitorModule{
		stopCh: make(chan struct{}),
	}
}
func (fm *FileMonitorModule) ID() string          { return "filemonitor" }
func (fm *FileMonitorModule) Name() string        { return "文件监控模块" }
func (fm *FileMonitorModule) Description() string { return "监控文件系统事件" }
func (fm *FileMonitorModule) Version() string     { return "1.0.0" }
func (fm *FileMonitorModule) Author() string      { return "小鱼" }

func (fm *FileMonitorModule) Init(ctx modInterfaces.Context) error {
	fm.ctx = ctx
	fm.ctx.Log("info", "文件监控模块已初始化")
	return nil

}
func (fm *FileMonitorModule) Start() error {
	fm.ctx.Log("info", "文件监控模块已启动")
	notify.NotifyInfo("文件监控模块已启动")
	// 初始化文件后缀检查器和行为缓存
	return nil
}
func (fm *FileMonitorModule) Stop() error {
	fm.ctx.Log("info", "文件监控模块已停止")
	close(fm.stopCh) // 关闭停止信号通道，通知模块退出
	fm.wg.Wait()     // 等待所有 goroutine 退出
	return nil
}
func (fm *FileMonitorModule) Status() modInterfaces.ModuleStatus {
	return fm.status
}
func (fm *FileMonitorModule) Reload() error {
	return nil
}

// func (fm *FileMonitorModule) Init(suffixChecker *SuffixChecker, behaviorCache *BehaviorCache, alertGenerator AlertGenerator) {
// 	fm.suffixChecker = suffixChecker
// 	fm.behaviorCache = behaviorCache
// 	fm.alertGenerator = alertGenerator
// }

func (fm *FileMonitorModule) handleFileEvent(event FileEvent) {
	suffixStatus := fm.suffixChecker.Check(event.Path)

	var alertType AlertType
	var alertMessage string
	switch suffixStatus {
	case SuffixAllowed:
		return
	case SuffixBlocked:
		alertType = AlertHighRisk
		alertMessage = "File matched a blacklisted extension"
		fmt.Println(alertMessage)
	case SuffixSuspicious:
		alertType = AlertSuspicious
		alertMessage = "Suspicious file extension detected"
	}

	if fm.behaviorCache.Record(event.Path) {
		alertJSON := fm.alertGenerator.GenerateAlert(alertType, event.Path)
		fmt.Println(alertJSON)
	}
}
