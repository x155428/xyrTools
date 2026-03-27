package memopt

// --- 内存优化模块 ---
import (
	"fmt"
	"strings"
	"sync"
	"syscall"
	"time"
	"xyrTools/xyrTools/extendFunc"
	"xyrTools/xyrTools/modInterfaces"

	"golang.org/x/sys/windows"
)

// --- 接口依赖 ---
type MemOptModule struct {
	status modInterfaces.ModuleStatus // 模块状态
	ctx    modInterfaces.Context      // 模块上下文
	stopCh chan struct{}              // 停止信号通道，stop时通知模块退出
	wg     sync.WaitGroup             // 等待组、确保模块退出时所有 goroutine 都已退出
}

func New() modInterfaces.Module {
	return &MemOptModule{
		stopCh: make(chan struct{}),
	}
}

func (m *MemOptModule) ID() string          { return "memopt" }
func (m *MemOptModule) Name() string        { return "内存优化模块" }
func (m *MemOptModule) Description() string { return "定期优化系统内存使用" }
func (m *MemOptModule) Version() string     { return "1.0.0" }
func (m *MemOptModule) Author() string      { return "小鱼" }

func (m *MemOptModule) Init(ctx modInterfaces.Context) error {
	m.ctx = ctx
	m.ctx.Log("info", "内存优化模块已初始化")
	// 订阅事件，优化所有进程
	m.ctx.Events.Subscribe("memory:optimized", func(evt modInterfaces.Event) {
		m.ctx.Log("info", fmt.Sprintf("收到事件 memory:optimized => %v", evt.Data))
		message, ok := evt.Data.(string)
		// 如果收到的信息是内存优化，则启动优化
		if ok && message == "运行内存优化任务" {
			m.ctx.Log("info", "收到内存优化任务，开始优化")
			m.Optimize()
		}
	})
	// 订阅事件，优化指定进程
	m.ctx.Events.Subscribe("memory:optimizeByNames", func(evt modInterfaces.Event) {
		m.ctx.Log("info", fmt.Sprintf("收到事件 memory:optimizeByNames => %v", evt.Data))
		names, ok := evt.Data.([]string)
		if ok && len(names) > 0 {
			m.ctx.Log("info", "收到指定进程名称列表，开始优化")
			m.OptimizeByNames(names...)
		}
	})
	return nil
}

func (m *MemOptModule) Start() error {
	m.status.Running = true
	m.status.StartTime = time.Now()

	interval := 30 // 默认 30 秒执行一次
	if val, ok := m.ctx.Config["interval"].(int); ok {
		interval = val
	}
	m.wg.Add(1)

	go func() {
		defer m.wg.Done()
		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		defer ticker.Stop()
		m.ctx.Log("info", "内存优化模块启动，间隔秒数: "+fmt.Sprint(interval))
		for {
			select {
			case <-ticker.C:
				m.Optimize()
				m.ctx.Events.Publish("memory:optimized", "内存优化完成。")
			case <-m.stopCh:
				m.ctx.Log("info", "内存优化模块停止")
				return
			}
		}
	}()
	return nil
}

func (m *MemOptModule) Stop() error {
	close(m.stopCh)
	m.wg.Wait()
	m.status.Running = false
	return nil
}

func (m *MemOptModule) Status() modInterfaces.ModuleStatus {
	return m.status
}

func (m *MemOptModule) Reload() error {
	m.ctx.Log("info", "内存优化模块重新加载配置")
	_ = m.Stop()
	m.stopCh = make(chan struct{})
	return m.Start()
}

// --- 内存优化核心逻辑 ---

var (
	modpsapi    = syscall.NewLazyDLL("psapi.dll")
	procEmptyWS = modpsapi.NewProc("EmptyWorkingSet")
)

func emptyWorkingSet(handle windows.Handle) {
	_, _, _ = procEmptyWS.Call(uintptr(handle))
}

func (m *MemOptModule) Optimize() {
	pids := make([]uint32, 1024)
	var needed uint32

	if err := windows.EnumProcesses(pids, &needed); err != nil {
		m.ctx.Log("error", "无法枚举进程: "+err.Error())
		extendFunc.MessageBox("提示", "无法枚举进程: "+err.Error())
		return
	}

	numProcs := needed / 4
	for i := 0; i < int(numProcs); i++ {
		pid := pids[i]
		if pid == 0 {
			continue
		}

		hProcess, err := windows.OpenProcess(windows.PROCESS_QUERY_INFORMATION|windows.PROCESS_SET_QUOTA, false, pid)
		if err != nil {
			continue
		}
		emptyWorkingSet(hProcess)
		_ = windows.CloseHandle(hProcess)
	}
	m.ctx.Log("info", "内存优化完成")
	extendFunc.MessageBox("提示", "内存优化完成")
}

// 优化指定进程内存
func (m *MemOptModule) OptimizeByNames(procNames ...string) {
	if len(procNames) == 0 {
		m.ctx.Log("warn", "未指定进程名称")
		extendFunc.MessageBox("提示", "未指定进程名称")
		return
	}

	nameSet := make(map[string]struct{})
	for _, name := range procNames {
		nameSet[strings.ToLower(name)] = struct{}{}
	}

	pids := make([]uint32, 1024)
	var needed uint32
	if err := windows.EnumProcesses(pids, &needed); err != nil {
		m.ctx.Log("error", "无法枚举进程: "+err.Error())
		return
	}

	numProcs := int(needed / 4)
	for i := 0; i < numProcs; i++ {
		pid := pids[i]
		if pid == 0 {
			continue
		}

		hProcess, err := windows.OpenProcess(windows.PROCESS_QUERY_INFORMATION|windows.PROCESS_VM_READ|windows.PROCESS_SET_QUOTA, false, pid)
		if err != nil {
			continue
		}

		var exeName [windows.MAX_PATH]uint16
		err = windows.GetModuleBaseName(hProcess, 0, &exeName[0], uint32(len(exeName)))
		if err != nil {
			windows.CloseHandle(hProcess)
			continue
		}
		processName := strings.ToLower(windows.UTF16ToString(exeName[:]))
		// 如果processName在procNames中，则优化
		if _, ok := nameSet[processName]; !ok {
			windows.CloseHandle(hProcess)
			continue
		}
		emptyWorkingSet(hProcess)
		//m.ctx.Log("info", fmt.Sprintf("已优化进程: %s (PID %d)", processName, pid))
		extendFunc.MessageBox("提示", fmt.Sprintf("已优化进程: %s (PID %d)", processName, pid))
		windows.CloseHandle(hProcess)
	}
}
