package modules

import (
	"fmt"
	//"test/modInterfaces"
)

// **Windows 文件监控模块**
type WindowsFileMonitor struct {
	config map[string]interface{}
}

func (m *WindowsFileMonitor) Start() error {
	fmt.Println("[WindowsFileMonitor] Windows 版文件监控启动！")
	return nil
}

func (m *WindowsFileMonitor) Stop() error {
	fmt.Println("[WindowsFileMonitor] Windows 版文件监控停止！")
	return nil
}

func (m *WindowsFileMonitor) Name() string {
	return "file_monitor"
}

// **补充 Config 方法**
func (m *WindowsFileMonitor) Config() interface{} {
	return m.config
}

func (m *WindowsFileMonitor) SetConfig(config interface{}) error {
	conf, ok := config.(map[string]interface{})
	if !ok {
		return fmt.Errorf("配置格式错误")
	}
	m.config = conf
	return nil
}
