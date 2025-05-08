package modules

import (
	"fmt"
	//"test/modInterfaces"
)

// **Linux/macOS 文件监控模块**
type LinuxFileMonitor struct {
	config map[string]interface{}
}

func (m *LinuxFileMonitor) Start() error {
	fmt.Println("[LinuxFileMonitor] Linux/macOS 版文件监控启动！")
	return nil
}

func (m *LinuxFileMonitor) Stop() error {
	fmt.Println("[LinuxFileMonitor] Linux/macOS 版文件监控停止！")
	return nil
}

func (m *LinuxFileMonitor) Name() string {
	return "file_monitor"
}

// **补充 Config 方法**
func (m *LinuxFileMonitor) Config() interface{} {
	return m.config
}

func (m *LinuxFileMonitor) SetConfig(config interface{}) error {
	conf, ok := config.(map[string]interface{})
	if !ok {
		return fmt.Errorf("配置格式错误")
	}
	m.config = conf
	return nil
}
