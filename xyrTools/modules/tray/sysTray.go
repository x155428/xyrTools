// 托盘模块，整个系统的托盘模块，用于管理系统托盘图标
package sysTray

import (
	"bytes"
	"context"
	"crypto/sha256"
	_ "embed"
	"fmt"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"time"
	"xyrTools/xyrTools/extendFunc"
	"xyrTools/xyrTools/modInterfaces"
	"xyrTools/xyrTools/modules/netManage"

	"github.com/gen2brain/beeep"

	"myMod/notify"

	"fyne.io/systray"
	ico "github.com/Kodeworks/golang-image-ico"
	"github.com/fsnotify/fsnotify"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

// 网卡配置数据结构
type NetworkInterface struct {
	Name      string
	IPAddress string
	Status    string
}

// 系统托盘模块
type SysTrayModule struct {
	status         modInterfaces.ModuleStatus
	ctx            modInterfaces.Context
	stopCh         chan struct{}
	wg             sync.WaitGroup
	netConfigItems []*systray.MenuItem
	cancelFuncs    []context.CancelFunc
}

// 图标路径、网络配置文件路径、子菜单处理函数映射
var (
	iconPath    string
	netCfgPath  string
	subHandlers = make(map[string]func(modInterfaces.Event))
)

func New() modInterfaces.Module {
	return &SysTrayModule{
		stopCh: make(chan struct{}),
	}
}

func (s *SysTrayModule) ID() string          { return "systray" }
func (s *SysTrayModule) Name() string        { return "系统托盘模块" }
func (s *SysTrayModule) Description() string { return "托盘模块" }
func (s *SysTrayModule) Version() string     { return "1.0.0" }
func (s *SysTrayModule) Author() string      { return "小鱼" }

// 初始化 SysTray 模块
func (s *SysTrayModule) Init(ctx modInterfaces.Context) error {
	s.ctx = ctx
	s.ctx.Log("info", "系统托盘模块已初始化")
	return nil
}

// 启动系统托盘
func (s *SysTrayModule) Start() error {
	s.ctx.Log("info", "SysTray 模块启动中")

	projectDir, err := os.Getwd()
	if err != nil {
		return err
	}
	iconPath = filepath.Join(projectDir, "icon", "cat.png")
	netCfgPath = filepath.Join(projectDir, "config", "netConfig.yaml")

	go func() {
		runtime.LockOSThread()
		s.ctx.Log("info", "启动系统托盘模块")
		s.startSysTray()
		runtime.UnlockOSThread()
	}()
	return nil
}

// 停止系统托盘
func (s *SysTrayModule) Stop() error {
	s.status.Running = false
	close(s.stopCh)
	s.wg.Wait()
	return nil
}

// 获取模块当前状态
func (s *SysTrayModule) Status() modInterfaces.ModuleStatus {
	return s.status
}

// 重载模块
func (s *SysTrayModule) Reload() error {
	s.ctx.Log("info", "系统托盘模块重新加载")
	_ = s.Stop()
	s.stopCh = make(chan struct{})
	return s.Start()
}

// 启动托盘菜单
func (s *SysTrayModule) startSysTray() {
	// 初始化托盘图标
	systray.Run(s.onReady, s.onExit)
}

// 托盘图标就绪时的回调
func (s *SysTrayModule) onReady() {
	systray.SetIcon(readIcon(iconPath))
	systray.SetTitle("系统工具")
	systray.SetTooltip("系统工具托盘模块")

	// 主菜单项
	netMenu := systray.AddMenuItem("网络配置", "打开网络配置面板")
	localNetMenu := systray.AddMenuItem("适配器管理", "本地适配器设置")
	netSwitchMenu := systray.AddMenuItem("切换配置", "应用预设网络配置")
	memoptThisMenu := systray.AddMenuItem("优化本进程内存", "运行内存优化任务")
	systray.AddSeparator()
	memOptMenu := systray.AddMenuItem("内存优化", "释放内存资源")
	systray.AddSeparator()
	infoMenu := systray.AddMenuItem("系统信息", "查看系统状态")
	systray.AddSeparator()
	openConsole := systray.AddMenuItem("控制台", "打开控制台")
	exitSys := systray.AddMenuItem("退出系统", "关闭系统")

	// 绑定事件处理
	s.bindMenuEvents(netMenu, localNetMenu, infoMenu, memOptMenu, openConsole, exitSys, memoptThisMenu)
	// 动态加载网络配置子菜单
	s.loadNetConfigs(netSwitchMenu)

	// 订阅配置更新事件
	s.subscribeNetCfgChange(netSwitchMenu)

	// 监听网卡配置文件
	projectDir, err := os.Getwd()
	if err != nil {
		s.ctx.Log("error", err.Error())
	}
	netCfgPath := filepath.Join(projectDir, "config", "netConfig.yaml")
	go s.watchConfigFile([]string{netCfgPath}, func(path string) {
		s.ctx.Log("info", fmt.Sprintf("配置文件 %s 变动，触发菜单更新", path))
		s.ctx.Events.Publish("sysTray:netCfgChanged", "netCfgChanged")
	})
}

func readIcon(iconPath string) []byte {
	// 读取图标文件
	return convertPNGToICO(iconPath)

}

// 托盘图标退出时的回调
func (s *SysTrayModule) onExit() {
	s.ctx.Log("info", "退出系统托盘模块")
}

// 显示网络配置子菜单
func (s *SysTrayModule) loadNetConfigs(parent *systray.MenuItem) {
	var mu sync.Mutex
	mu.Lock()
	defer mu.Unlock()
	s.clearSubMenuItems()
	configs, err := netManage.LoadConfigFromFile(netCfgPath)
	if err != nil {
		s.ctx.Log("error", "加载网络配置失败: "+err.Error())
		return
	}

	for _, cfg := range configs {
		cfg := cfg
		item := parent.AddSubMenuItem(cfg.Name, fmt.Sprintf("应用到 %s", cfg.Adapter))
		s.netConfigItems = append(s.netConfigItems, item)

		ctx, cancel := context.WithCancel(context.Background())
		s.cancelFuncs = append(s.cancelFuncs, cancel)
		go func(c netManage.NetConfig, m *systray.MenuItem, ctx context.Context) {
			for {
				select {
				case <-ctx.Done():
					return
				case <-m.ClickedCh:
					s.ctx.Log("info", "应用配置: "+c.Name)
					err := netManage.ApplyNetConfig(cfg)
					if err != nil {
						s.ctx.Log("error", "应用配置失败: "+err.Error())
					}
				}
			}
		}(cfg, item, ctx)
	}
}

// 清空子菜单
func (s *SysTrayModule) clearSubMenuItems() {
	for _, cancel := range s.cancelFuncs {
		cancel()
	}
	s.cancelFuncs = nil
	for _, item := range s.netConfigItems {
		item.Remove()
	}
	s.netConfigItems = nil

}

func (s *SysTrayModule) subscribeNetCfgChange(parent *systray.MenuItem) {
	handler := func(evt modInterfaces.Event) {
		s.ctx.Log("info", fmt.Sprintf("配置变更事件: %v", evt.Data))
		s.loadNetConfigs(parent)
		s.ctx.Events.Unsubscribe("sysTray:netCfgChanged", subHandlers["netCfg"])
	}
	subHandlers["netCfg"] = handler
	s.ctx.Events.Subscribe("sysTray:netCfgChanged", handler)
}

func (s *SysTrayModule) bindMenuEvents(net, local, info, mem, openConsole, exitOs, memoptThis *systray.MenuItem) {
	go func() {
		for {
			select {
			case <-net.ClickedCh:
				s.openNetworkConfigWindow()
			case <-local.ClickedCh:
				s.openNcpa()
			case <-info.ClickedCh:
				s.showSystemInfo()
			case <-mem.ClickedCh:
				s.ctx.Events.Publish("memory:optimized", "运行内存优化任务")
			case <-openConsole.ClickedCh:
				//consoleutil.CreateConsole()
				notify.NotifyInfo("控制台待开发！")
			case <-exitOs.ClickedCh:
				extendFunc.RemoveLockFile()
				os.Exit(0)
			case <-memoptThis.ClickedCh:
				processes := []string{"xyrTools.exe"}
				s.ctx.Events.Publish("memory:optimizeByNames", processes)
			}
		}
	}()
}

// 打开网络连接属性窗口
func (s *SysTrayModule) openNcpa() {
	var cmd *exec.Cmd
	// 尝试 rundll32.exe，速度最快
	cmd = exec.Command("rundll32.exe", "shell32.dll,Control_RunDLL", "ncpa.cpl")
	//err := cmd.Start()
	err := cmd.Run()
	if err == nil {
		return
	}
	s.ctx.Log("error", fmt.Sprintf("rundll32 启动失败: %v", err))

	// 尝试 control.exe，兼容性强
	cmd = exec.Command("control.exe", "ncpa.cpl")
	err = cmd.Start()
	if err == nil {
		go func() {
			_ = cmd.Wait() // 回收子进程资源
		}()
		return
	}
	s.ctx.Log("error", fmt.Sprintf("control.exe 启动失败: %v", err))

	// 最后尝试 cmd /C，最慢，但成功率高
	cmd = exec.Command("cmd", "/C", "ncpa.cpl")
	err = cmd.Run()
	if err != nil {
		go func() {
			_ = cmd.Wait() // 回收子进程资源
		}()
		s.ctx.Log("error", fmt.Sprintf("cmd /C 启动失败: %v", err))
		return
	}
	go func() {
		_ = cmd.Wait() // 回收子进程资源
	}()
	return
}

// 显示系统基本信息
func (s *SysTrayModule) showSystemInfo() {
	go func() {
		s.ctx.Log("info", "显示系统基本信息")

		// 收集数据
		cpuPercent, err1 := cpu.Percent(time.Second, false)
		vmStat, err2 := mem.VirtualMemory()
		hostInfo, err3 := host.Info()

		if err1 != nil || err2 != nil || err3 != nil {
			s.ctx.Log("error", "获取系统信息失败")
			return
		}

		bootTime := time.Unix(int64(hostInfo.BootTime), 0).Format("2006-01-02 15:04:05")

		// 日志详细记录
		s.ctx.Log("info", fmt.Sprintf(
			"系统信息:\n"+
				" - 主机名: %s\n"+
				" - 系统: %s %s (%s)\n"+
				" - 开机时间: %v\n"+
				" - CPU 使用率: %.2f%%\n"+
				" - 总内存: %.2f GB\n"+
				" - 内存使用率: %.2f%%",
			hostInfo.Hostname,
			hostInfo.Platform, hostInfo.PlatformVersion, hostInfo.KernelArch,
			bootTime,
			cpuPercent[0],
			float64(vmStat.Total)/1e9,
			vmStat.UsedPercent,
		))

		// 提炼通知内容
		notifyTitle := "系统状态更新"
		notifyBody := fmt.Sprintf(
			"CPU使用率: %.1f%%\n内存使用率: %.1f%%\n开机: %s",
			cpuPercent[0],
			vmStat.UsedPercent,
			bootTime,
		)

		// 推送通知
		err := beeep.Notify(notifyTitle, notifyBody, "assets/information.png")
		if err != nil {
			s.ctx.Log("error", "发送系统通知失败: "+err.Error())
		}
	}()
}

// 打开网络管理窗口
func (s *SysTrayModule) openNetworkConfigWindow() {
	showNetManageGui()

}

// 辅助函数：将 PNG 图片转换为 ICO 格式
func convertPNGToICO(pngPath string) []byte {
	file, err := os.Open(pngPath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		panic(err)
	}

	buf := new(bytes.Buffer)
	err = ico.Encode(buf, img)
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}

// 配置文件监听，当配置文件变化时重构菜单
func (s *SysTrayModule) watchConfigFile(filePaths []string, onChange func(path string)) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	// 记录每个文件的最新 hash
	lastHashes := make(map[string][32]byte)

	// 初始化 hash
	for _, file := range filePaths {
		hash, _ := fileHash(file)
		lastHashes[file] = hash

		// 监听该文件所在目录
		dir := filepath.Dir(file)
		err := watcher.Add(dir)
		if err != nil {
			return err
		}
	}

	go func() {
		defer watcher.Close()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) != 0 {
					changedFile := event.Name
					for _, watchFile := range filePaths {
						if filepath.Clean(changedFile) == filepath.Clean(watchFile) {
							newHash, _ := fileHash(watchFile)
							if newHash != lastHashes[watchFile] {
								lastHashes[watchFile] = newHash
								onChange(watchFile)
							}
						}
					}
				}
			case err := <-watcher.Errors:
				_ = err // 可记录日志
			}
		}
	}()

	return nil
}

func (s *SysTrayModule) openConsole() error {
	// 创建并执行 cmd 命令，使用 /k 保持命令行窗口打开
	cmd := exec.Command("cmd", "/k", "echo This is an interactive console window! && echo Type commands here... && set")

	// 设置窗口为可交互模式
	cmd.Stdout = os.Stdout // 重定向输出到当前程序的标准输出
	cmd.Stderr = os.Stderr // 重定向错误输出
	cmd.Stdin = os.Stdin   // 使用户输入能够传递给新命令行窗口

	// 启动命令行窗口
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start interactive console: %v", err)
	}

	// 等待命令执行
	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("error occurred while running the command: %v", err)
	}

	return nil
}

// 计算文件的 hash
func fileHash(path string) ([32]byte, error) {
	var empty [32]byte
	data, err := os.ReadFile(path)
	if err != nil {
		return empty, err
	}
	return sha256.Sum256(data), nil
}

var (
	kernel32  = syscall.NewLazyDLL("kernel32.dll")
	procAlloc = kernel32.NewProc("AllocConsole")
	procFree  = kernel32.NewProc("FreeConsole")
)

func createConsole() error {
	// 分配控制台
	ret, _, err := procAlloc.Call()
	if ret == 0 {
		return fmt.Errorf("AllocConsole failed: %v", err)
	}

	// 重定向 STDOUT
	out, err := os.OpenFile("CONOUT$", os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("open CONOUT$ failed: %v", err)
	}
	os.Stdout = out
	os.Stderr = out

	// 重定向 STDIN
	in, err := os.OpenFile("CONIN$", os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("open CONIN$ failed: %v", err)
	}
	os.Stdin = in

	fmt.Println("控制台已创建，开始输出…")
	return nil
}

func closeConsole() error {
	ret, _, err := procFree.Call()
	if ret == 0 {
		return fmt.Errorf("FreeConsole failed: %v", err)
	}
	return nil
}
