package consoleutil

import (
	"fmt"
	"myMod/notify"
	"os"
	"sync"
	"syscall"
)

var (
	kernel32             = syscall.NewLazyDLL("kernel32.dll")
	procAllocConsole     = kernel32.NewProc("AllocConsole")
	procFreeConsole      = kernel32.NewProc("FreeConsole")
	procSetCtrlHandler   = kernel32.NewProc("SetConsoleCtrlHandler")
	procGetConsoleWindow = kernel32.NewProc("GetConsoleWindow")

	consoleLock sync.Mutex
	consoleIn   *os.File
	consoleOut  *os.File
	handlerPtr  uintptr
	handlerSet  bool
)

// 判断控制台是否已附加
func isConsoleAttached() bool {
	ret, _, _ := procGetConsoleWindow.Call()
	return ret != 0
}

// 创建控制台并重定向输入输出
func CreateConsole() error {
	consoleLock.Lock()
	defer consoleLock.Unlock()

	if isConsoleAttached() {
		return nil
	}

	ret, _, err := procAllocConsole.Call()
	if ret == 0 {
		return fmt.Errorf("AllocConsole failed: %v", getLastError(err))
	}

	if err := redirectStdIO(); err != nil {
		return fmt.Errorf("redirectStdIO failed: %v", err)
	}

	// 注册关闭事件处理器
	handlerPtr = syscall.NewCallback(func(ctrlType uint32) uintptr {
		if ctrlType == 2 { // CTRL_CLOSE_EVENT
			//notify.NotifyInfo("事件2")
			go func() {
				_ = ReleaseConsole()
			}()
			//notify.NotifyInfo("返回1")
			return 1
		}
		return 0
	})
	procSetCtrlHandler.Call(handlerPtr, 1)
	handlerSet = true

	return nil
}

// 释放控制台并清理资源
func ReleaseConsole() error {
	consoleLock.Lock()
	defer consoleLock.Unlock()

	if !isConsoleAttached() {
		notify.NotifyInfo("当前没有控制台窗口")
		return nil
	}

	// 清理 handler
	if handlerSet {
		procSetCtrlHandler.Call(handlerPtr, 0)
		handlerSet = false
	}

	// 重定向至 NUL，防止非法输出
	_ = redirectToNull()

	// 释放控制台
	//notify.NotifyInfo("释放控制台窗口...")
	ret, _, err := procFreeConsole.Call()
	if ret == 0 {
		notify.NotifyError(err, "释放控制台窗口失败")
		return fmt.Errorf("FreeConsole failed: %v", getLastError(err))
	}

	// 关闭句柄
	if consoleIn != nil {
		consoleIn.Close()
		consoleIn = nil
	}
	if consoleOut != nil {
		consoleOut.Close()
		consoleOut = nil
	}

	notify.NotifyInfo("控制台窗口已释放")
	return nil
}

// 将标准输入输出指向控制台窗口
func redirectStdIO() error {
	out, err := os.OpenFile("CONOUT$", os.O_RDWR, 0)
	if err != nil {
		return err
	}
	in, err := os.OpenFile("CONIN$", os.O_RDWR, 0)
	if err != nil {
		out.Close()
		return err
	}

	// 关闭原句柄
	if consoleOut != nil {
		consoleOut.Close()
	}
	if consoleIn != nil {
		consoleIn.Close()
	}

	consoleOut = out
	consoleIn = in

	os.Stdout = out
	os.Stderr = out
	os.Stdin = in

	return nil
}

// 重定向至 NUL
func redirectToNull() error {
	nullFile, err := os.OpenFile("NUL", os.O_RDWR, 0)
	if err != nil {
		return err
	}
	os.Stdout = nullFile
	os.Stderr = nullFile
	os.Stdin = nullFile
	return nil
}

// 获取更有意义的错误信息
func getLastError(e error) error {
	if errno, ok := e.(syscall.Errno); ok {
		return fmt.Errorf("%s (code: %d)", errno.Error(), int(errno))
	}
	return e
}
