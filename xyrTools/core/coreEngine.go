// core/engine.go
package core

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"xyrTools/xyrTools/modInterfaces"

	"github.com/google/uuid"
	"gopkg.in/yaml.v2"
)

type CoreEngine struct {
	modules   map[string]*modInterfaces.ModuleInstance
	eventBus  modInterfaces.EventBus
	globalCfg map[string]interface{}
	log       func(string, string)
	lock      sync.Mutex
}

func NewCoreEngine(logFunc func(string, string)) *CoreEngine {
	return &CoreEngine{
		modules:   make(map[string]*modInterfaces.ModuleInstance),
		eventBus:  modInterfaces.NewEventBus(),
		log:       logFunc,
		globalCfg: make(map[string]interface{}),
	}
}

// 加载配置
func (e *CoreEngine) LoadConfig(path string) error {
	cfgPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("获取配置文件路径出错: %w", err)
	}
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return fmt.Errorf("读取配置文件出错: %w", err)
	}
	cfg := make(map[string]interface{})
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("yaml格式错误: %w", err)
	}
	e.globalCfg = cfg
	return nil
}

// Register 方法用于向核心引擎注册一个模块。
// 参数 id 为模块的唯一标识符。
// 参数 factory 是一个工厂函数，用于创建模块实例。
// 返回值为错误信息，若注册过程中出现问题则返回相应错误，注册成功返回 nil。
func (e *CoreEngine) Register(id string, factory func() modInterfaces.Module) error {
	// 加锁，确保在注册模块时不会有其他并发操作修改模块列表
	e.lock.Lock()
	// 函数结束时解锁，保证锁的正确释放
	defer e.lock.Unlock()

	// 检查模块是否已经注册
	if _, exists := e.modules[id]; exists {
		// 若模块已存在，返回错误信息
		return fmt.Errorf("module %s already registered", id)
	}

	// 从全局配置中获取对应模块的配置信息
	cfgRaw, ok := e.globalCfg[id]
	if !ok {
		// 若配置中不存在该模块的配置，返回错误信息
		return fmt.Errorf("module %s not found in config", id)
	}

	// 尝试将模块配置转换为 map[interface{}]interface{} 类型
	rawMap, ok := cfgRaw.(map[interface{}]interface{})
	if !ok {
		// 若转换失败，说明模块配置不是有效的 map 类型，返回错误信息
		return fmt.Errorf("module %s config is not a valid map", id)
	}

	// 将 map[interface{}]interface{} 类型的配置转换为 map[string]interface{} 类型
	cfg := convertMap(rawMap)

	// 使用工厂函数创建模块实例
	mod := factory()
	uuidStr := uuid.NewString()
	heartbeatCfg := e.globalCfg["heartbeat"]
	heartbeatCfgTmp, OK := heartbeatCfg.(map[interface{}]interface{})
	if !OK {
		return fmt.Errorf("heartbeat config is not a valid map")
	}
	heartbeatCfgMap := convertMap(heartbeatCfgTmp)
	heartbeatAddr := heartbeatCfgMap["heartbeatAddress"].(string)

	// 构建模块上下文配置
	ctx := modInterfaces.Context{

		Config:        cfg,           // 模块配置信息
		Log:           e.log,         // 日志函数
		Events:        e.eventBus,    // 事件总线
		Uuid:          uuidStr,       // 模块实例ID
		Name:          id,            // 模块名称
		HeartbeatAddr: heartbeatAddr, // 心跳地址
	}

	// 向模块注入上下文信息并初始化模块
	if err := mod.Init(ctx); err != nil {
		// 若初始化失败，返回错误信息
		return err
	}

	// 将模块实例及其上下文信息添加到核心引擎的模块列表中
	e.modules[id] = &modInterfaces.ModuleInstance{

		Impl:       mod,     // 模块实例
		Ctx:        ctx,     // 模块上下文
		InstanceID: uuidStr, // 模块实例ID
	}

	// 注册成功，返回 nil
	return nil
}

// 启动所有模块
func (e *CoreEngine) StartAll() {
	for id, inst := range e.modules {
		e.log("info", fmt.Sprintf("Starting module %s", id))
		enabled, ok := inst.Ctx.Config["enabled"].(bool)
		// 模块配置为关闭
		if !ok || !enabled {
			e.log("info", fmt.Sprintf("Module %s is disabled", id))
			continue
		} else {
			// 模块配置为开启
			if err := inst.Impl.Start(); err != nil {
				// 模块启动失败
				e.log("error", fmt.Sprintf("Module %s failed to start: %v", id, err))
				inst.Status.LastError = err
			} else {
				inst.Status.StartTime = time.Now()
				inst.Status.Running = true
			}
		}

	}
}

// 停止所有模块
func (e *CoreEngine) StopAll() {
	// 遍历所有模块
	for id, inst := range e.modules {
		// 模块未启动，跳过
		if !inst.Status.Running {
			e.log("info", fmt.Sprintf("Module %s is not running, skipping stop", id))
			continue
		}
		if err := inst.Impl.Stop(); err != nil {
			e.log("error", fmt.Sprintf("Module %s failed to stop: %v", id, err))
		} else {
			e.log("info", fmt.Sprintf("Stopping module %s", id))
			inst.Status.Running = false
			inst.Status.EndTime = time.Now()
		}
	}
}

// 获取事件总线
func (e *CoreEngine) GetEventBus() modInterfaces.EventBus {
	return e.eventBus
}

// 日志函数
func (e *CoreEngine) Log(level, msg string) {
	e.log(level, msg)
}

// 获取全局配置
func (e *CoreEngine) GetConfig() map[string]interface{} {
	return e.globalCfg
}

// 转换 map[interface{}]interface{} 到 map[string]interface{}
func convertMap(input map[interface{}]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range input {
		strKey := fmt.Sprintf("%v", k)
		result[strKey] = v
	}
	return result
}

// 启动指定模块
func (e *CoreEngine) StartModule(id string) error {
	inst, exists := e.modules[id]
	if !exists {
		return fmt.Errorf("module %s not found", id)
	}
	if inst.Status.Running {
		return fmt.Errorf("module %s is already running", id)
	}
	if err := inst.Impl.Start(); err != nil {
		return err
	}
	inst.Status.StartTime = time.Now()
	inst.Status.Running = true
	return nil

}

// 停止指定模块
func (e *CoreEngine) StopModule(id string) error {
	inst, exists := e.modules[id]
	if !exists {
		return fmt.Errorf("module %s not found", id)
	}
	if !inst.Status.Running {
		return fmt.Errorf("module %s is not running", id)
	}
	if err := inst.Impl.Stop(); err != nil {
		return err
	}
	inst.Status.EndTime = time.Now()
	inst.Status.Running = false
	return nil

}

// 重新加载指定模块
func (e *CoreEngine) ReloadModule(id string) error {
	inst, exists := e.modules[id]
	if !exists {
		return fmt.Errorf("module %s not found", id)
	}
	if err := inst.Impl.Reload(); err != nil {
		return err
	}
	return nil

}

// 获取模块状态
func (e *CoreEngine) GetModuleStatus(id string) (modInterfaces.ModuleStatus, error) {
	inst, exists := e.modules[id]
	if !exists {
		return modInterfaces.ModuleStatus{}, fmt.Errorf("module %s not found", id)
	}
	return inst.Status, nil

}

// 心跳检测
func (e *CoreEngine) StartHeartBeatServer(addr string) error {
	http.HandleFunc("/heartbeat", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// 启动 HTTP 服务，监听心跳请求
	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			e.log("error", fmt.Sprintf("心跳服务启动失败: %v", err))
		}
	}()

	e.log("info", fmt.Sprintf("心跳服务已启动，监听地址: %s", addr))
	return nil
}
