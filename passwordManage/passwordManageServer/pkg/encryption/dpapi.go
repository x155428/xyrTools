/*
 * @Author: 小鱼
 * @Date: 2025-10-20 10:41:27
 * @LastEditors: 小鱼
 * @LastEditTime: 2025-10-20 11:04:18
 * @FilePath: \passwordManageServer\pkg\encryption\dpapi.go
 * @Description:
 *
 * Copyright (c) 2025 by 小鱼, All Rights Reserved.
 */
package encryption

import (
	"errors"
	"fmt"
	"runtime"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	DpapiUserScope    = 0
	DpapiMachineScope = windows.CRYPTPROTECT_LOCAL_MACHINE
)

// 实现封装dpapi的加解密
// 给定数据，使用dpapi加密，返回加密后的数据
// flags: 指定加密选项，默认为0（当前用户范围，本地机器范围等）
func DpapiEncrypt(data []byte, flags uint32) ([]byte, error) {
	// 检查数据是否为空
	if len(data) == 0 {
		return nil, errors.New("no data to encrypt")
	}

	// 检查是否为Windows系统
	if runtime.GOOS == "windows" {
		// 使用Windows DPAPI加密
		in := windows.DataBlob{
			Size: uint32(len(data)),
			Data: &data[0],
		}

		var out windows.DataBlob

		// 调用Windows API进行加密，使用指定的flags
		err := windows.CryptProtectData(&in, nil, nil, 0, nil, flags, &out)
		if err != nil {
			return nil, fmt.Errorf("CryptProtectData failed: %w", err)
		}

		// 确保释放由CryptProtectData分配的内存
		defer windows.LocalFree(windows.Handle(unsafe.Pointer(out.Data)))

		// 复制加密数据到Go字节切片
		//enc := make([]byte, out.Size)
		enc := unsafe.Slice((*byte)(unsafe.Pointer(out.Data)), out.Size)
		encCopy := make([]byte, out.Size)
		copy(encCopy, enc)

		return encCopy, nil
	} else {
		// 非Windows系统不支持DPAPI
		return nil, fmt.Errorf("DPAPI is only supported on Windows, current OS: %s", runtime.GOOS)
	}
}

// 给定数据，解密数据，返回解密后的数据
// flags: 指定解密选项，默认为0
func DpapiDecrypt(data []byte, flags uint32) ([]byte, error) {
	// 检查数据是否为空
	if len(data) == 0 {
		return nil, errors.New("no data to decrypt")
	}

	// 检查是否为Windows系统
	if runtime.GOOS == "windows" {
		// 使用Windows DPAPI解密
		in := windows.DataBlob{
			Size: uint32(len(data)),
		}
		in.Data = &data[0]

		var out windows.DataBlob

		// 调用Windows API进行解密，使用指定的flags
		err := windows.CryptUnprotectData(&in, nil, nil, 0, nil, flags, &out)
		if err != nil {
			if errno, ok := err.(syscall.Errno); ok && errno == windows.ERROR_INVALID_DATA {
				return nil, fmt.Errorf("DPAPI decryption failed: key invalid or user context changed")
			}
			return nil, fmt.Errorf("CryptUnprotectData failed: %w", err)
		}

		// 确保释放由CryptUnprotectData分配的内存
		defer windows.LocalFree(windows.Handle(unsafe.Pointer(out.Data)))

		// 复制解密数据到Go字节切片
		plain := unsafe.Slice((*byte)(unsafe.Pointer(out.Data)), out.Size)
		plainCopy := make([]byte, out.Size)
		copy(plainCopy, plain)

		return plainCopy, nil
	} else {
		// 非Windows系统不支持DPAPI
		return nil, fmt.Errorf("DPAPI is only supported on Windows, current OS: %s", runtime.GOOS)
	}
}
