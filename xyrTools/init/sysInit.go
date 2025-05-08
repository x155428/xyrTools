package initSys

import (
	"fmt"
	"strings"
	"xyrTools/xyrTools/core"
	modInterfaces "xyrTools/xyrTools/modInterfaces"
	memopt "xyrTools/xyrTools/modules/memoryOptimizer"
	sysTray "xyrTools/xyrTools/modules/tray"
)

// InitSys 初始化系统，加载配置并注册配置文件中指定的模块。
// 参数 coreEngine 是核心引擎实例，用于配置加载、日志记录和模块注册。
// 参数 confPath 是配置文件的路径。
func InitSys(coreEngine *core.CoreEngine, confPath string) {
	// 模块工厂映射，键为模块名称，值为创建模块实例的工厂函数。
	// 当有新模块添加或现有模块移除时，需要修改此映射。
	moduleFactoriesMap := map[string]func() modInterfaces.Module{
		"memopt":  memopt.New,
		"sysTray": sysTray.New,
	}

	// 若加载失败，记录致命错误日志并终止初始化流程。
	if err := coreEngine.LoadConfig(confPath); err != nil {
		coreEngine.Log("fatal", err.Error())

		return
	}

	// 从核心引擎获取全局配置信息。
	globalCfg := coreEngine.GetConfig()
	// 尝试从配置中读取模块列表。
	modulesList, ok := globalCfg["modules"]
	if !ok {
		// 若配置中不存在模块列表，记录致命错误日志并终止初始化流程。
		coreEngine.Log("fatal", "No modules found in config")
		return
	}
	// 将模块列表从字符串类型转换为字符串切片。
	modulesSlice := strings.Fields(modulesList.(string))
	if len(modulesSlice) == 0 {
		// 若转换后的模块列表为空，记录致命错误日志并终止初始化流程。
		coreEngine.Log("fatal", "No valid modules specified")
		return
	}
	// 遍历模块列表，对每个模块进行处理。
	for _, mod := range modulesSlice {
		// 根据模块名从模块工厂映射中获取对应的工厂函数。
		factory, ok := moduleFactoriesMap[mod]
		if !ok {
			// 若未找到对应的工厂函数，记录错误日志并跳过该模块。
			coreEngine.Log("error", fmt.Sprintf("No module factory found for module %s", mod))
			continue
		}
		// 使用核心引擎注册模块，传入模块名和对应的工厂函数。
		if err := coreEngine.Register(mod, factory); err != nil {
			// 若注册失败，记录错误日志。
			coreEngine.Log("error", fmt.Sprintf("Failed to register module %s: %v", mod, err))
		}
	}
}
