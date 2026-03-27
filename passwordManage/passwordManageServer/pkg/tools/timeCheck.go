package tools

import (
	"errors"
	"strconv"			
	"time"
)

// 时间戳校验

// CheckTimestamp 校验时间戳是否在规定时间窗口内，allowTime默认5分钟，单位秒
// 支持多种时间戳格式：整数类型的秒/毫秒/微秒/纳秒，以及字符串类型的时间戳
// 时间戳必须在当前时间的 ±allowTime 范围内才视为有效
func CheckTimestamp(timestamp interface{}, allowTime int64) bool {
	// 校验时间戳是否在规定时间内，allowTime默认5分钟
	if allowTime == 0 {
		allowTime = 5 * 60
	}
	
	// 解析时间戳到time.Time
	t, err := parseTimestamp(timestamp)
	if err != nil {
		return false // 无法解析的时间戳视为无效
	}
	
	// 获取当前时间
	now := time.Now()
	
	// 检查时间差，确保时间戳在允许的时间窗口内（过去或未来）
	diff := now.Sub(t)
	if diff < 0 {
		diff = -diff
	}
	
	if diff > time.Duration(allowTime)*time.Second {
		return false
	}
	return true
}

// parseTimestamp 解析多种格式的时间戳到time.Time
func parseTimestamp(timestamp interface{}) (time.Time, error) {
	switch v := timestamp.(type) {
	case int64:
		// 整数类型时间戳，自动判断精度
		return parseIntegerTimestamp(v), nil
	case int:
		// int类型转换为int64再处理
		return parseIntegerTimestamp(int64(v)), nil
	case int32:
		// int32类型转换为int64再处理
		return parseIntegerTimestamp(int64(v)), nil
	case string:
		// 字符串类型时间戳
		return parseStringTimestamp(v)
	default:
		// 不支持的类型
		return time.Time{}, errors.New("unsupported timestamp type")
	}
}

// parseIntegerTimestamp 解析整数类型的时间戳，自动判断精度
func parseIntegerTimestamp(ts int64) time.Time {
	// 根据时间戳长度判断精度
	// 秒级: 10位数字，毫秒级: 13位数字，微秒级: 16位数字，纳秒级: 19位数字
	tsStr := strconv.FormatInt(ts, 10)
	length := len(tsStr)
	
	// 处理负时间戳
	if ts < 0 {
		length-- // 减去负号的长度
	}
	
	switch {
	case length <= 10:
		// 秒级时间戳
		return time.Unix(ts, 0)
	case length <= 13:
		// 毫秒级时间戳
		return time.Unix(ts/1000, (ts%1000)*1000000)
	case length <= 16:
		// 微秒级时间戳
		return time.Unix(ts/1000000, (ts%1000000)*1000)
	default:
		// 纳秒级时间戳或其他高精度
		return time.Unix(ts/1000000000, ts%1000000000)
	}
}

// parseStringTimestamp 解析字符串类型的时间戳
func parseStringTimestamp(tsStr string) (time.Time, error) {
	// 尝试直接转换为整数
	if ts, err := strconv.ParseInt(tsStr, 10, 64); err == nil {
		return parseIntegerTimestamp(ts), nil
	}
	
	// 尝试解析为RFC3339格式时间字符串
	if t, err := time.Parse(time.RFC3339, tsStr); err == nil {
		return t, nil
	}
	
	// 尝试解析为常见的日期时间格式
	timeFormats := []string{
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05",
		"2006/01/02 15:04:05",
		"2006-01-02",
		"2006/01/02",
	}
	
	for _, format := range timeFormats {
		if t, err := time.Parse(format, tsStr); err == nil {
			return t, nil
		}
	}
	
	return time.Time{}, errors.New("invalid timestamp string format")
}
