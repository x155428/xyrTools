package netManage

import (
	"os"

	"gopkg.in/yaml.v2"
)

type ConfigFile struct {
	Configs []NetConfig `yaml:"configs"`
}

type NetConfig struct {
	Name    string `yaml:"name"`    // 配置名称 自定义
	Desc    string `yaml:"desc"`    // 配置描述 自定义
	Adapter string `yaml:"adapter"` // 网卡名称
	DHCP    bool   `yaml:"dhcp"`    // 是否使用DHCP
	DNSdhcp bool   `yaml:"dnsdhcp"` // 是否使用DHCP获取DNS
	IP      string `yaml:"ip"`      // IP地址
	Netmask string `yaml:"netmask"` // 子网掩码
	Gateway string `yaml:"gateway"` // 网关

	DNS      []string `yaml:"dns"`      // DNS服务器
	MTU      int      `yaml:"mtu"`      // MTU大小
	Metric   int      `yaml:"metric"`   // 跃点数
	FlushDNS bool     `yaml:"flushDNS"` // 是否刷新DNS缓存
}

func LoadConfigFromFile(path string) ([]NetConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg ConfigFile
	err = yaml.Unmarshal(data, &cfg)
	return cfg.Configs, err
}
