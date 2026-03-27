package config

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// NetworkConfig 用于解析传入的网络配置
type NetworkConfig struct {
	Name     string   `json:"Name"`
	Adapter  string   `json:"Adapter"`
	DHCP     bool     `json:"DHCP"`
	DNSdhcp  bool     `json:"DNSdhcp"`
	IP       string   `json:"IP"`
	Netmask  string   `json:"Netmask"`
	Gateway  string   `json:"Gateway"`
	DNS      []string `json:"DNS"`
	MTU      int      `json:"MTU"`
	Metric   int      `json:"Metric"`
	FlushDNS bool     `json:"FlushDNS"`
}

// ExecutionResult 封装结果信息
type ResultMessage struct {
	Success bool   `json:"Success"` //成功失败标识
	Details string `json:"Details"` //详细信息
	Other   string `json:"Other"`   //其他信息
}

func runCommand(cmd *exec.Cmd) ResultMessage {
	// 设置命令行字符集为 UTF-8
	cmd = exec.Command("cmd", "/C", "chcp 65001 && "+strings.Join(cmd.Args, " "))
	// 执行命令并获取输出
	output, err := cmd.CombinedOutput()
	// 如果执行失败，标记为失败
	if err != nil {
		return ResultMessage{Success: false, Details: err.Error(), Other: string(output)}
	}
	// 如果执行成功，标记为成功
	return ResultMessage{Success: true, Details: "命令执行成功", Other: string(output)}
}

func ConfigureNetwork(config NetworkConfig) ResultMessage {
	// 网卡名称
	interfaceName := config.Adapter
	// TODO:检查网卡是否存在

	// 配置 DHCP
	if config.DHCP {
		// dhcp模式
		cmd := exec.Command("netsh", "interface", "ip", "set", "address", "name="+interfaceName, "source=dhcp")
		// 命令配置为dhcp
		result := runCommand(cmd)
		if !result.Success {
			// 配置dhcp异常
			// 检查 Other 字段是否包含 "DHCP is already enabled on this interface"，已配置过dhcp，重新启用网卡
			if strings.Contains(result.Other, "DHCP is already enabled on this interface") {
				// 禁用再启用网卡
				cmd := exec.Command("netsh", "interface", "set", "interface", "name="+interfaceName, "admin=disable")
				result := runCommand(cmd)
				if !result.Success {
					// 禁用网卡失败
					return ResultMessage{Success: false, Details: "禁用网卡失败，请手动检查！", Other: result.Other}
				}

				// 启用网卡
				cmd = exec.Command("netsh", "interface", "set", "interface", "name="+interfaceName, "admin=enable")
				result = runCommand(cmd)
				if !result.Success {
					return ResultMessage{Success: false, Details: "启用网卡失败，请手动检查！", Other: result.Other}
				}

				// dhcp配置成功，检查dns配置模式，手动/自动获取
				if !config.DNSdhcp {
					// 手动配置 DNS
					for _, dns := range config.DNS {
						cmd = exec.Command("netsh", "interface", "ip", "set", "dns", "name="+interfaceName, "static", dns, "primary")
						result := runCommand(cmd)
						if !result.Success {
							return ResultMessage{Success: false, Details: "配置 DNS 失败", Other: result.Other}
						}
					}
				} else {
					//配置dns自动获取
					cmd = exec.Command("netsh", "interface", "ip", "set", "dns", "name="+interfaceName, "source=dhcp")
					result := runCommand(cmd)
					if !result.Success {
						return ResultMessage{Success: false, Details: "配置 DNS 自动获取失败", Other: result.Other}
					}
				}
			} else {
				// 其他失败原因，返回失败信息
				return ResultMessage{Success: false, Details: "配置 DHCP 失败", Other: result.Other}
			}
		}

	} else {
		// 设置静态 IP 地址
		cmd := exec.Command("netsh", "interface", "ip", "set", "address", "name="+interfaceName, "static", config.IP, config.Netmask, config.Gateway)
		result := runCommand(cmd)
		if !result.Success {
			return ResultMessage{Success: false, Details: "配置静态 IP 失败", Other: result.Other}
		}

		// 配置 DNS
		if !config.DNSdhcp {
			for _, dns := range config.DNS {
				cmd = exec.Command("netsh", "interface", "ip", "set", "dns", "name="+interfaceName, "static", dns, "primary")
				result := runCommand(cmd)
				if !result.Success {
					return ResultMessage{Success: false, Details: "配置 DNS 失败", Other: result.Other}
				}
			}
		}
	}

	// 配置 MTU
	cmd := exec.Command("netsh", "interface", "ipv4", "set", "subinterface", interfaceName, "mtu="+fmt.Sprint(config.MTU), "store=persistent")
	result := runCommand(cmd)
	if !result.Success {
		return ResultMessage{Success: false, Details: "配置 MTU 失败", Other: result.Other}
	}

	// 配置 Metric
	cmd = exec.Command("netsh", "interface", "ipv4", "set", "subinterface", interfaceName, "metric="+fmt.Sprint(config.Metric), "store=persistent")
	result = runCommand(cmd)
	if !result.Success {
		return ResultMessage{Success: false, Details: "配置 Metric 失败", Other: result.Other}
	}

	// 清除 DNS 缓存
	if config.FlushDNS {
		cmd = exec.Command("ipconfig", "/flushdns")
		result := runCommand(cmd)
		if !result.Success {
			return ResultMessage{Success: false, Details: "清除 DNS 缓存失败", Other: result.Other}
		}
	}

	return ResultMessage{Success: true, Details: "配置成功"}
}

func ParseConfigAndConfigure(jsonStr string) ResultMessage {
	// 解析 JSON 配置
	var config NetworkConfig
	err := json.Unmarshal([]byte(jsonStr), &config)
	if err != nil {
		return ResultMessage{Success: false, Details: "配置解析失败", Other: err.Error()}
	}

	// 执行配置
	return ConfigureNetwork(config)
}
