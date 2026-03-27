package passwordmanage

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"time"
	"xyrTools/xyrTools/modInterfaces"

	"github.com/spf13/cast"
)

type PasswordManageModule struct {
	status modInterfaces.ModuleStatus // 模块状态
	ctx    modInterfaces.Context      // 模块上下文
	stopCh chan struct{}              // 停止信号通道，stop时通知模块退出
	wg     sync.WaitGroup             // 等待组、确保模块退出时所有 goroutine 都已退出

	// 扩展字段
	pid    int          // 进程ID，用于管理进程
	server *http.Server // 服务器实例，用于管理web服务
}

func New() modInterfaces.Module {
	return &PasswordManageModule{
		stopCh: make(chan struct{}),
	}
}
func (m *PasswordManageModule) ID() string          { return "passwordmanage" }
func (m *PasswordManageModule) Name() string        { return "密码管理模块" }
func (m *PasswordManageModule) Description() string { return "管理和生成密码" }
func (m *PasswordManageModule) Version() string     { return "1.0.0" }
func (m *PasswordManageModule) Author() string      { return "小鱼" }

func (m *PasswordManageModule) Init(ctx modInterfaces.Context) error {
	m.ctx = ctx
	m.ctx.Log("info", "密码管理模块已初始化")
	return nil

}
func (m *PasswordManageModule) Start() error {
	m.ctx.Log("info", "密码管理模块已启动")
	//go m.StartHeartbeatLoop()
	// 获取当前目录
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %v", err)
	}
	// 获取配置信息
	config := m.ctx.Config
	isHeartBeat := cast.ToString(config["isHeartbeat"])

	// 获取uuid和名字
	uuid := m.ctx.Uuid
	name := m.ctx.Name
	heartbeatAddr := m.ctx.HeartbeatAddr
	// 从配置中获取密码管理模块的配置信息
	fmt.Println(config)
	serverPath, ok := config["serverPath"].(string)
	if !ok {
		return fmt.Errorf("serverPath not found in config")
	}
	// 命令行启动服务端
	serverPath = filepath.Join(dir, serverPath)
	// 传参 uuid 和 name
	m.pid, err = StartExeByPath(serverPath, []string{"--uuid", uuid, "--name", name, "--isHeartbeat", isHeartBeat, "--heartbeatAddress", heartbeatAddr})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("启动成功，PID=", m.pid)

	// 启动web端
	webPath, ok := config["clientPath"].(string)
	if !ok {
		return fmt.Errorf("clientPath not found in config")
	}

	domain, ok := config["domain"].(string)
	if !ok {
		return fmt.Errorf("domain not found in config")
	}
	port, ok := config["clientPort"].(int)
	if !ok {
		return fmt.Errorf("port not found in config")
	}
	// 构建完整的web路径
	webPath = filepath.Join(dir, webPath)
	// 拼接web地址
	portTmp := strconv.Itoa(port)
	webAddress := domain + ":" + portTmp
	// 设置Tls证书路径
	certPath, ok := config["certPath"].(string)
	if !ok {
		return fmt.Errorf("certPath not found in config")
	}
	keyPath, ok := config["keyPath"].(string)
	if !ok {
		return fmt.Errorf("keyPath not found in config")
	}
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return fmt.Errorf("failed to load certificate: %v", err)
	}
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	// 启动web端
	fileServer := http.FileServer(http.Dir(webPath))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		requestPath := filepath.Join(webPath, r.URL.Path)
		if stat, err := os.Stat(requestPath); err == nil && !stat.IsDir() {
			fileServer.ServeHTTP(w, r)
		} else {
			http.ServeFile(w, r, filepath.Join(webPath, "index.html"))
		}
	})

	m.server = &http.Server{
		Addr:      webAddress,
		Handler:   nil,
		TLSConfig: tlsConfig,
	}
	log.Printf("HTTPS 服务启动中: https://%s\n", webAddress)
	go func() {
		if err := m.server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			m.ctx.Log("error", fmt.Sprintf("HTTPS server error: %v", err))
		}
	}()

	return nil
}
func (m *PasswordManageModule) Stop() error {
	// 关闭web服务
	if m.server != nil {
		if err := m.server.Close(); err != nil {
			m.ctx.Log("error", fmt.Sprintf("HTTPS server close error: %v", err))
		}
	}
	// 关闭进程
	if m.pid != 0 {
		process, err := os.FindProcess(m.pid)
		if err != nil {
			m.ctx.Log("error", fmt.Sprintf("Failed to find process: %v", err))
		}
		if err := process.Kill(); err != nil {
			m.ctx.Log("error", fmt.Sprintf("Failed to kill process: %v", err))
		}
		m.pid = 0
	}
	return nil
}
func (m *PasswordManageModule) Status() modInterfaces.ModuleStatus {
	return m.status
}
func (m *PasswordManageModule) Reload() error {
	if err := m.Stop(); err != nil {
		return fmt.Errorf("failed to stop module: %v", err)
	}
	if err := m.Start(); err != nil {
		return fmt.Errorf("failed to start module: %v", err)
	}
	m.ctx.Log("info", "密码管理模块已重新加载")
	return nil
}

func StartExeByPath(path string, args []string) (int, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return 0, fmt.Errorf("无法解析路径: %v", err)
	}
	if _, err := os.Stat(absPath); err != nil {
		return 0, fmt.Errorf("文件不存在: %v", err)
	}

	cmd := exec.Command(absPath, args...)
	cmd.Dir = filepath.Dir(absPath) // 设置工作目录为exe所在目录
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("启动失败: %v", err)
	}

	pid := cmd.Process.Pid
	log.Printf("已启动 %s [PID=%d]", absPath, pid)

	go func() {
		err := cmd.Wait()
		if err != nil {
			log.Printf("进程 %d 异常退出: %v", pid, err)
		} else {
			log.Printf("进程 %d 正常退出", pid)
		}
	}()

	return pid, nil
}

func (m *PasswordManageModule) sendHeartbeat() bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://127.0.0.1:6000/heartbeat")
	if err != nil || resp.StatusCode != 200 {
		return false
	}
	return true
}

func (m *PasswordManageModule) StartHeartbeatLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	const maxFailures = 2 // 最大允许连续失败次数
	failCount := 0        // 当前连续失败次数

	for {
		select {
		case <-ticker.C:
			if m.sendHeartbeat() {
				failCount = 0 // 成功则清零
			} else {
				failCount++
				fmt.Printf("心跳失败 %d 次\n", failCount)
				if failCount >= maxFailures {
					fmt.Println("主控连续失联，准备自毁")
					// 自毁逻辑
					m.Stop()
					return
				}
			}
		case <-m.stopCh:
			return
		}
	}
}
