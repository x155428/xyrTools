package netManage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os/exec"
	"strconv"
	"time"
	"xyrTools/xyrTools/extendFunc"

	"github.com/Microsoft/go-winio"
)

func ApplyNetConfig1(cfg NetConfig) error {
	// 1. 检查配置的有效性
	if err := validateNetConfig(cfg); err != nil {
		return err
	}

	// 2. 应用 DHCP 设置
	if cfg.DHCP {
		if err := applyDHCP(cfg.Adapter); err != nil {
			return err
		}

		// 如果需要通过 DHCP 获取 DNS
		if cfg.DNSdhcp {
			if err := applyDNSDHCP(cfg.Adapter); err != nil {
				return err
			}
		} else if len(cfg.DNS) > 0 {
			// 设置静态 DNS
			if err := applyStaticDNS(cfg.Adapter, cfg.DNS); err != nil {
				return err
			}
		}
	} else {
		// 配置静态 IP
		if err := applyStaticIP(cfg.Adapter, cfg.IP, cfg.Netmask, cfg.Gateway); err != nil {
			return err
		}

		// 配置静态 DNS
		if len(cfg.DNS) > 0 {
			if err := applyStaticDNS(cfg.Adapter, cfg.DNS); err != nil {
				return err
			}
		}
	}

	// 3. 应用 MTU 配置
	if cfg.MTU > 0 {
		if err := applyMTU(cfg.Adapter, cfg.MTU); err != nil {
			return err
		}
	}

	// 4. 应用 Metric 配置
	if cfg.Metric > 0 {
		if err := applyMetric(cfg.Adapter, cfg.Metric); err != nil {
			return err
		}
	}

	// 5. 刷新 DNS 缓存
	if cfg.FlushDNS {
		if err := flushDNSCache(); err != nil {
			return err
		}
	}

	return nil
}
func ApplyNetConfig(cfg NetConfig) error {
	// 创建一次性命名管道
	pipePath := `\\.\pipe\netCfgPipe`
	timeout := time.Second * 10
	// 尝试连接命名管道（最多等待10秒）
	conn, err := winio.DialPipe(pipePath, &timeout)
	if err != nil {
		extendFunc.MessageBox("提示", "Failed to connect to pipe:"+err.Error())
		return err
	}
	defer conn.Close()
	// 将 cfg 序列化为 JSON 字节切片
	cfgData, err := json.Marshal(cfg)
	if err != nil {
		extendFunc.MessageBox("提示", "Failed to marshal config:"+err.Error())
		return err
	}
	packageData, packageErr := extendFunc.PackageData(string(cfgData), "#")
	if packageErr != nil {
		//fmt.Println("Failed to package config:", packageErr)
		extendFunc.MessageBox("提示", "Failed to package config:"+packageErr.Error())
		return packageErr
	}

	// 发送数据到命名管道
	_, err = conn.Write(packageData)
	if err != nil {
		//fmt.Println("Failed to write to pipe:", err)
		extendFunc.MessageBox("提示", "Failed to write to pipe:"+err.Error())
		return err
	}
	fmt.Println("Sent:", string(packageData))
	//extendFunc.MessageBox("提示", string(packageData))
	// 接收服务端回应
	buf := make([]byte, 2048)
	n, err := conn.Read(buf)
	if err != nil && err != io.EOF {
		fmt.Println("Read error:", err)
		return err
	}
	if n > 0 {
		unpackateData, isUnpackage, unpackageErr := extendFunc.UnpackageData(buf[:n], "#")
		if !isUnpackage {
			fmt.Println("Error unpackage data:", unpackageErr)
			fmt.Println("Received:", string(buf[:n]))
			return unpackageErr
		}
		//fmt.Println("Received:", unpackateData)
		extendFunc.MessageBox("提示", unpackateData)
	} else {
		//fmt.Println("No response received.")
		extendFunc.MessageBox("提示", "No response received.")
	}

	return nil
}

// 验证配置的合法性
func validateNetConfig(cfg NetConfig) error {
	printNetworkInterfaces()
	// 检查配置中指定的网卡是否存在
	if err := checkAdapterExistence(cfg.Adapter); err != nil {
		return err
	}
	if !cfg.DHCP && cfg.DNSdhcp {
		return errors.New("非法配置：静态 IP 模式下不能使用 DNS DHCP")
	}

	if net.ParseIP(cfg.IP) == nil && !cfg.DHCP {
		return fmt.Errorf("无效 IP 地址: %s", cfg.IP)
	}

	if net.ParseIP(cfg.Netmask) == nil && !cfg.DHCP {
		return fmt.Errorf("无效子网掩码: %s", cfg.Netmask)
	}

	if net.ParseIP(cfg.Gateway) == nil && !cfg.DHCP {
		return fmt.Errorf("无效网关: %s", cfg.Gateway)
	}

	for _, dns := range cfg.DNS {
		if net.ParseIP(dns) == nil {
			return fmt.Errorf("无效 DNS: %s", dns)
		}
	}

	if cfg.MTU < 576 || cfg.MTU > 9000 {
		return fmt.Errorf("MTU 不在合理范围: %d", cfg.MTU)
	}

	return nil
}

// 应用 DHCP 设置
func applyDHCP(adapter string) error {
	cmd := exec.Command("netsh", "interface", "ip", "set", "address", "name="+adapter, "source=dhcp")
	err := cmd.Run()

	if err != nil {
		exec.Command("powershell", "-Command", "Disable-NetAdapter -Name '"+adapter+"' -Confirm:$false; Start-Sleep -Seconds 1; Enable-NetAdapter -Name '"+adapter+"' -Confirm:$false").Run()
		exec.Command("ipconfig", "/release").Run()
		exec.Command("ipconfig", "/renew").Run()
		return err
	}
	exec.Command("powershell", "-Command", "Disable-NetAdapter -Name '"+adapter+"' -Confirm:$false; Start-Sleep -Seconds 1; Enable-NetAdapter -Name '"+adapter+"' -Confirm:$false").Run()
	exec.Command("ipconfig", "/release").Run()
	exec.Command("ipconfig", "/renew").Run()
	return nil
}

// 应用 DHCP DNS 设置
func applyDNSDHCP(adapter string) error {
	cmd := exec.Command("netsh", "interface", "ip", "set", "dns", "name="+adapter, "source=dhcp")
	return cmd.Run()
}

// 应用静态 DNS 设置
func applyStaticDNS(adapter string, dns []string) error {
	cmd := exec.Command("netsh", "interface", "ip", "set", "dns", "name="+adapter, "static", dns[0])
	err := cmd.Run()
	if err != nil {
		return err
	}
	for i := 1; i < len(dns); i++ {
		cmd := exec.Command("netsh", "interface", "ip", "add", "dns", "name="+adapter, dns[i], "index="+strconv.Itoa(i+1))
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}

// 应用静态 IP 配置
func applyStaticIP(adapter, ip, netmask, gateway string) error {
	cmd := exec.Command("netsh", "interface", "ip", "set", "address", "name="+adapter, "static", ip, netmask, gateway)
	return cmd.Run()
}

// 应用 MTU 设置
func applyMTU(adapter string, mtu int) error {
	cmd := exec.Command("netsh", "interface", "ip", "set", "interface", "name="+adapter, "mtu="+strconv.Itoa(mtu))
	return cmd.Run()
}

// 应用 Metric 设置
func applyMetric(adapter string, metric int) error {
	cmd := exec.Command("netsh", "interface", "ip", "set", "interface", "name="+adapter, "metric="+strconv.Itoa(metric))
	return cmd.Run()
}

// 刷新 DNS 缓存
func flushDNSCache() error {
	cmd := exec.Command("ipconfig", "/flushdns")
	return cmd.Run()
}

//	辅助函数
//
// 获取本机网卡列表并返回
func getNetworkInterfaces() ([]net.Interface, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	return interfaces, nil
}

// 打印网卡列表，用于调试
func printNetworkInterfaces() {
	interfaces, err := getNetworkInterfaces()
	if err != nil {
		fmt.Println("获取网卡列表失败:", err)
		return
	}

	fmt.Println("本机网卡列表:")
	for _, iface := range interfaces {
		fmt.Printf("网卡名称: %s, 网卡状态: %s\n", iface.Name, iface.HardwareAddr)
	}
}

// 检查网卡是否存在
func checkAdapterExistence(adapter string) error {
	interfaces, err := getNetworkInterfaces()
	if err != nil {
		return fmt.Errorf("获取网卡列表失败: %v", err)
	}

	// 检查指定网卡是否在列表中
	for _, iface := range interfaces {
		if iface.Name == adapter {
			return nil
		}
	}

	return fmt.Errorf("网卡 %s 不存在", adapter)
}

// 获取指定网卡当前配置信息
func getAdapterConfig(adapter string) (NetConfig, error) {
	// 检查网卡是否存在
	if err := checkAdapterExistence(adapter); err != nil {
		return NetConfig{}, err
	}
	var cfg NetConfig
	var err error

	return cfg, err
}
