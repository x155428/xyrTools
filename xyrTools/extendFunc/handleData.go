package extendFunc

import (
	"encoding/binary"
	"fmt"
)

// 封装数据：将数据与结尾符号拼接并返回
func PackageData(data string, delimiter string) ([]byte, error) {
	// 计算数据长度
	dataLength := uint32(len(data))

	// 创建缓冲区，首先是长度（4字节），然后是数据内容，最后是结尾符号
	result := make([]byte, 4+len(data)+len(delimiter))

	// 将数据长度转换为大端字节序并放入缓冲区
	binary.BigEndian.PutUint32(result[:4], dataLength)

	// 将数据内容放入缓冲区
	copy(result[4:], data)

	// 将结尾符号放入缓冲区
	copy(result[4+len(data):], delimiter)

	return result, nil
}

// 解析数据：从接收到的字节流中解析出数据并校验结尾符
func UnpackageData(buf []byte, delimiter string) (string, bool, error) {
	// 确保接收到的字节流足够长
	if len(buf) < 4 {
		return "", false, fmt.Errorf("buffer too short")
	}

	// 获取数据长度字段（前4个字节）
	dataLength := binary.BigEndian.Uint32(buf[:4])

	// 确保数据流包含数据长度和结尾符号
	expectedLength := 4 + int(dataLength) + len(delimiter)
	if len(buf) < expectedLength {
		return "", false, fmt.Errorf("incomplete data received")
	}

	// 提取数据内容
	data := string(buf[4 : 4+dataLength])

	// 校验结尾符号
	receivedDelimiter := string(buf[4+dataLength:])
	isValid := receivedDelimiter == delimiter

	return data, isValid, nil
}
