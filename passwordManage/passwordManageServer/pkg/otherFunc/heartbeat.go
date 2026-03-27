/*
 * @Author: 小鱼
 * @Date: 2025-07-03 15:27:13
 * @LastEditors: 小鱼
 * @LastEditTime: 2025-09-29 13:51:18
 * @FilePath: \passwordManageServer\pkg\otherFunc\heartbeat.go
 * @Description: 心跳检测模块，负责处理客户端心跳请求，维持会话活跃
 *
 * Copyright (c) 2025 by 小鱼, All Rights Reserved.
 */
package otherFunc

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func HeartBeat(addr, uuid, name string) {
	const interval = 20 * time.Second // 心跳间隔
	const maxMissed = 2               // 最大允许丢包次数

	client := &http.Client{
		Timeout: 5 * time.Second, // 每次心跳请求的超时时间
	}

	body := fmt.Sprintf(`{"uuid": "%s", "name": "%s"}`, uuid, name)
	missed := 0 // 连续未回应计数
	addr = "http://" + addr + "/heartbeat"

	for {
		req, err := http.NewRequest("POST", addr, strings.NewReader(body))
		if err != nil {
			log.Printf("构造请求失败: %v", err)
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			missed++
			log.Printf("发送失败 (%d/%d): %v", missed, maxMissed, err)
		} else {
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				missed++
				log.Printf("非正常响应 (%d/%d): %s", missed, maxMissed, resp.Status)
			} else {
				missed = 0
			}
		}

		if missed >= maxMissed {
			log.Println("主控已失联，自毁！")
			os.Exit(1)
		}

		time.Sleep(interval)
	}
}
