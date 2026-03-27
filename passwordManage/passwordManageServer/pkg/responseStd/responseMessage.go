/*
 * @Author: 小鱼
 * @Date: 2025-09-29 14:00:00
 * @LastEditors: 小鱼
 * @LastEditTime: 2025-09-29 14:00:00
 * @FilePath: \passwordManageServer\pkg\responseStd\responseMessage.go
 * @Description: 标准响应格式定义，提供统一的API响应结构和返回方法
 *
 * Copyright (c) 2025 by 小鱼, All Rights Reserved.
 */
package responseStd

import (
	"encoding/json"
	"net/http"
)

// CommonResponse 定义了标准的响应格式
// @description: 通用响应结构体
// @property {int} Code - HTTP状态码
// @property {string} Message - 响应消息
// @property {interface{}} Data - 响应数据
// @property {bool} Success - 成功状态标志
type CommonResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Success bool        `json:"success"`
}

// SendSuccessResponse 发送成功响应
// @description: 发送标准的成功响应
// @param {http.ResponseWriter} w - HTTP响应写入器
// @param {string} message - 成功消息
// @param {interface{}} data - 响应数据
func SendSuccessResponse(w http.ResponseWriter, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	response := CommonResponse{
		Code:    http.StatusOK,
		Message: message,
		Data:    data,
		Success: true,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// SendErrorResponse 发送错误响应
// @description: 发送标准的错误响应
// @param {http.ResponseWriter} w - HTTP响应写入器
// @param {int} statusCode - HTTP状态码
// @param {string} message - 错误消息
func SendErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	response := CommonResponse{
		Code:    statusCode,
		Message: message,
		Data:    nil,
		Success: false,
	}
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
