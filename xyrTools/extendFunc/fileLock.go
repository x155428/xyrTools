package extendFunc

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

var (
	lockFilePath = "./program.lock"

	user32         = syscall.NewLazyDLL("user32.dll")
	procMessageBox = user32.NewProc("MessageBoxW")
)

// 弹出消息框
func MessageBox(title, text string) {
	titlePtr, _ := syscall.UTF16PtrFromString(title)
	textPtr, _ := syscall.UTF16PtrFromString(text)
	procMessageBox.Call(0, uintptr(unsafe.Pointer(textPtr)), uintptr(unsafe.Pointer(titlePtr)), 0)
}

func CheckLockFile() bool {
	// 检查锁文件是否已经存在
	if _, err := os.Stat(lockFilePath); err == nil {
		// 如果文件存在，表示程序已启动
		return true
	}
	return false
}

func CreateLockFile() error {
	// 创建锁文件，如果文件存在就表示程序已启动
	lockFile, err := os.Create(lockFilePath)
	if err != nil {
		return err
	}
	defer lockFile.Close()

	// 将当前进程ID写入锁文件中，标记程序正在运行
	_, err = lockFile.WriteString(fmt.Sprintf("%d", os.Getpid()))
	return err
}

func RemoveLockFile() error {
	// 删除锁文件
	return os.Remove(lockFilePath)
}
