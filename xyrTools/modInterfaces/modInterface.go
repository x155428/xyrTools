package modInterfaces

import (
	"fmt"
	"sync"
	"time"
)

// --- 模块封装体（注册后管理状态） ---
type ModuleInstance struct {
	Impl   Module       // 模块实体，存储各种模块实体
	Ctx    Context      // 初始化上下文
	Status ModuleStatus // 当前运行状态

	Mutex sync.Mutex // 多线程安全
}

type Module interface {
	ID() string          // 唯一标识符（建议小写英文）
	Name() string        // 可显示名称
	Description() string // 简短描述
	Version() string     // 版本号
	Author() string      // 作者信息

	Init(ctx Context) error // 初始化模块，注入上下文
	Start() error           // 启动模块运行逻辑
	Stop() error            // 停止模块运行逻辑
	Status() ModuleStatus   // 获取当前状态

	Reload() error // 重新加载
}

// 模块上下文信息（启动时注入）
type Context struct {
	Config map[string]interface{}         // 模块独立配置
	Log    func(level string, msg string) // 日志函数
	Events EventBus                       // 模块间通信事件总线
}

// --- 模块运行状态 ---
type ModuleStatus struct {
	Running   bool
	LastError error
	StartTime time.Time
	EndTime   time.Time
}

// --- 事件结构体 ---
type Event struct {
	Name string
	Data interface{}
}

// --- 事件总线接口 ---
type EventBus interface {
	Subscribe(event string, handler func(Event))   // 订阅事件
	Unsubscribe(event string, handler func(Event)) // 取消订阅
	Publish(event string, data interface{})        // 发布事件
}

// --- 默认事件总线实现 ---
type defaultEventBus struct {
	subscribers map[string][]func(Event)
	lock        sync.RWMutex
}

func NewEventBus() EventBus {
	return &defaultEventBus{
		subscribers: make(map[string][]func(Event)),
	}
}

func (bus *defaultEventBus) Subscribe(event string, handler func(Event)) {
	bus.lock.Lock()
	defer bus.lock.Unlock()
	bus.subscribers[event] = append(bus.subscribers[event], handler)
}

func (bus *defaultEventBus) Unsubscribe(event string, handler func(Event)) {
	bus.lock.Lock()
	defer bus.lock.Unlock()
	handlers := bus.subscribers[event]
	for i, h := range handlers {
		if &h == &handler {
			// 从订阅者列表中移除指定的 handler
			bus.subscribers[event] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}
}

// Publish 方法用于发布一个事件，会触发所有订阅该事件的处理函数。
// event 为要发布的事件名称。
// data 为事件携带的数据，可传递任意类型的数据。
func (bus *defaultEventBus) Publish(event string, data interface{}) {
	// 使用读锁，允许多个 goroutine 同时读取订阅者列表，提高并发性能
	bus.lock.RLock()
	// 获取订阅该事件的所有处理函数
	handlers := bus.subscribers[event]
	// 释放读锁
	bus.lock.RUnlock()

	// 遍历所有订阅该事件的处理函数
	for _, handler := range handlers {
		// 为每个处理函数启动一个新的 goroutine 来执行，避免阻塞当前 goroutine
		go func(h func(Event)) {
			// 使用 defer 和 recover 捕获可能出现的 panic，防止一个处理函数的 panic 影响其他处理函数
			defer func() {
				if r := recover(); r != nil {
					// log打印错误信息
					logFunc := func(level, msg string) {
						fmt.Printf("[%s] %s\n", level, msg)
					}
					logFunc("error", fmt.Sprintf("Event handler panic: %v", r))
				}
			}()
			// 调用处理函数，传入包含事件名称和数据的 Event 结构体
			h(Event{Name: event, Data: data})
		}(handler)
	}
}
