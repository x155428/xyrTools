// core/engine.go
package core

import (
	"fmt"
	"os"
	"sync"
	"time"

	"xyrTools/xyrTools/modInterfaces"

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
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}
	cfg := make(map[string]interface{})
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("invalid yaml format: %w", err)
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

	// 构建模块上下文配置
	ctx := modInterfaces.Context{

		Config: cfg,        // 模块配置信息
		Log:    e.log,      // 日志函数
		Events: e.eventBus, // 事件总线
	}

	// 向模块注入上下文信息并初始化模块
	if err := mod.Init(ctx); err != nil {
		// 若初始化失败，返回错误信息
		return err
	}

	// 将模块实例及其上下文信息添加到核心引擎的模块列表中
	e.modules[id] = &modInterfaces.ModuleInstance{

		Impl: mod, // 模块实例
		Ctx:  ctx, // 模块上下文
	}

	// 注册成功，返回 nil
	return nil
}

// 根据模块配置启动所有模块
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
