package extendFunc

import (
	"github.com/shirou/gopsutil/v3/net"
)

// 单个网卡当前的网络流量计数
type NetIO struct {
	Name      string //网卡名称
	BytesSent uint64 //发送的字节数
	BytesRecv uint64 //接收的字节数
}

// 获取当前所有网卡的网络流量计数
// 返回值: 每个网卡的网络流量计数，错误信息
func GetNetIOCounters() ([]NetIO, error) {
	counters, err := net.IOCounters(true)
	if err != nil {
		return nil, err
	}

	var result []NetIO
	for _, c := range counters {
		result = append(result, NetIO{
			Name:      c.Name,
			BytesSent: c.BytesSent,
			BytesRecv: c.BytesRecv,
		})
	}

	return result, nil
}

// 单个网卡在时间段内的上下行速率（KB/s）
type NetIOSpeed struct {
	Name     string
	RecvKBps float64
	SendKBps float64
}

// 计算速率
// before: 前一次的网络流量计数
// after: 当前的网络流量计数
// intervalSec: 时间间隔（秒）
// 返回值: 每个网卡的上下行速率
func CalcNetIOSpeed(before, after []NetIO, intervalSec float64) []NetIOSpeed {
	beforeMap := make(map[string]NetIO)
	for _, stat := range before {
		beforeMap[stat.Name] = stat
	}

	var speeds []NetIOSpeed
	for _, a := range after {
		b, ok := beforeMap[a.Name]
		if !ok {
			continue // 网卡在前一次不存在，跳过
		}
		recv := float64(a.BytesRecv-b.BytesRecv) / 1024.0 / intervalSec
		send := float64(a.BytesSent-b.BytesSent) / 1024.0 / intervalSec

		speeds = append(speeds, NetIOSpeed{
			Name:     a.Name,
			RecvKBps: recv,
			SendKBps: send,
		})
	}

	return speeds
}
