/*
 * @Author: 小鱼
 * @Date: 2025-09-29 14:00:00
 * @LastEditors: 小鱼
 * @LastEditTime: 2025-09-29 14:00:00
 * @FilePath: \passwordManageServer\pkg\otherFunc\jsonFunc.go
 * @Description: JSON处理相关工具函数，提供JSON解析和验证功能
 *
 * Copyright (c) 2025 by 小鱼, All Rights Reserved.
 */
package otherFunc

import (
	"bytes"
	"encoding/json"
	"io"
)

func HasID(data []byte) (bool, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	for {
		t, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return false, err
		}
		if key, ok := t.(string); ok && key == "id" {
			return true, nil // 找到即返回
		}
	}
	return false, nil
}
