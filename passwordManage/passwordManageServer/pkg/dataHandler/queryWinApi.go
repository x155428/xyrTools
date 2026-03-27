/**
 * @Author: 小鱼
 * @Date: 2024-04-18 13:22:35
 * @LastEditors: 小鱼
 * @LastEditTime: 2024-04-18 13:22:35
 * @FilePath: \passwordManageServer\pkg\dataHandler\queryWinApi.go
 * @Description: WinAPI查询相关数据结构定义
 *
 * Copyright (c) 2024 by 小鱼, All Rights Reserved.
 */
package dataHandler

type WinApiIndex struct {
	Module       string `json:"module"`
	FunctionName string `json:"functionName"`
}
