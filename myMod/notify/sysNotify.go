package notify

import (
	"github.com/go-toast/toast"
)

// 发送一个 Windows 系统通知
func SendSystemNotification(title, message string) error {
	notification := toast.Notification{
		AppID:   "系统服务提示",
		Title:   title,
		Message: message,
	}
	return notification.Push()
}

// 错误
func NotifyError(err error, context string) {
	if err != nil {
		SendSystemNotification("错误："+context, err.Error())
	}
}

// 提示
func NotifyInfo(info string) {
	SendSystemNotification("提示", info)
}
