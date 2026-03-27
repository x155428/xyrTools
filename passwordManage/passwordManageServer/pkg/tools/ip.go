/*
 * @Author: 小鱼
 * @Date: 2025-10-29 13:46:16
 * @LastEditors: 小鱼
 * @LastEditTime: 2025-10-29 16:28:05
 * @FilePath: \passwordManageServer\pkg\tools\ip.go
 * @Description:
 *
 * Copyright (c) 2025 by 小鱼, All Rights Reserved.
 */
package tools

import (
	"fmt"
	"net"
	"strings"
)

// 如果所有地址格式正确，返回 true，并修改原数组，标准化所有地址
// 如果任何地址格式无效，返回 false，原数组不修改
func IpFormatCheck(ips *[]string) bool {
	for i, ip := range *ips {
		// 检查并标准化每个CIDR地址
		standardizedIP, valid := ipFormatCheckAndStandardize(ip)
		if !valid {
			return false
		}
		// 修改原数组中的地址为标准化后的地址
		(*ips)[i] = standardizedIP
	}
	return true
}

// ipFormatCheckAndStandardize 检查输入的 IP 地址是否都是有效的 IP 地址或标准化的 CIDR 格式
func ipFormatCheckAndStandardize(ip string) (string, bool) {
	// 空字符串直接返回 false
	if ip == "" {
		return "", false
	}

	// 检查是否是CIDR格式（例如：192.168.1.0/24 或 2001:db8::/32）
	if strings.Contains(ip, "/") {
		// 尝试解析为CIDR格式
		_, ipNet, err := net.ParseCIDR(ip)
		if err != nil {
			return "", false
		}

		// 标准化CIDR地址
		normalizedIP := ipNet.IP.Mask(ipNet.Mask)

		// 获取子网掩码的位数
		ones, _ := ipNet.Mask.Size()

		// 重新构建标准化后的CIDR地址
		standardizedCIDR := normalizedIP.String() + "/" + fmt.Sprintf("%d", ones)

		// 返回标准化后的CIDR地址
		return standardizedCIDR, true
	} else {
		// 尝试解析为普通IP地址（IPv4 或 IPv6）
		parsedIP := net.ParseIP(ip)
		if parsedIP != nil {
			return ip, true // 返回原始IP地址
		}
	}

	// 如果不是有效的CIDR或IP地址，返回 false
	return "", false
}

// IsIPInWhitelist 检查给定的 IP 地址是否在白名单内
func IsIPInWhitelist(ip string, whitelist []string) (bool, error) {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false, fmt.Errorf("无效的 IP 地址：%s", ip)
	}

	// 遍历白名单，检查每个条目
	for _, entry := range whitelist {
		// 如果条目包含 "/"，说明它是一个 CIDR 地址
		if strings.Contains(entry, "/") {
			// 解析 CIDR 子网
			_, network, err := net.ParseCIDR(entry)
			if err != nil {
				return false, fmt.Errorf("无效的 CIDR 地址：%s", entry)
			}
			// 检查 IP 是否在 CIDR 子网范围内
			if network.Contains(parsedIP) {
				return true, nil
			}
		} else {
			// 如果条目是单个 IP 地址，直接比较
			if parsedIP.String() == entry {
				return true, nil
			}
		}
	}

	// 如果没有匹配的条目，返回 false
	return false, nil
}
