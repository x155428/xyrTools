# xyrTools

xyrTools 是一个模块化的系统工具集合，采用发布订阅制架构，通过消息传递实现模块间的解耦交互。项目旨在提供一套可扩展、可拆卸的系统工具解决方案，涵盖内存优化、网络管理、密码管理、文件监控等多种功能。

## 功能特性

### 核心功能
- **内存优化**：调用系统 API 实现内存优化，支持自动优化和手动触发
- **网络管理**：提供网络适配器配置编辑、应用功能，包含高权限服务支持
- **密码管理**：完整的密码管理系统，支持加密存储、安全访问
- **文件监控**：监控文件变动，支持异常后缀变动告警（勒索病毒检测）【暂未实现】
- **系统托盘**：提供系统托盘界面，快速访问常用功能

### 监控与告警【暂未实现，属于TODO】
- **系统监控**：采集系统常用数据、等保检查项数据
- **性能监控**：CPU、内存突然暴涨告警（挖矿检测），达到阈值自动执行内存优化
- **文件监控**：大量异常后缀变动告警（勒索病毒）、文件夹内容变动提醒（共享文件夹变动）
- **处置模块**：根据告警和预定义动作，对告警进行处置，集成少量高危特征库

## 系统架构

项目采用发布订阅制架构，各模块通过消息传递实现解耦交互。核心引擎负责消息的分发与处理，各功能模块可独立开发和部署。

### 架构层次
1. **核心层**：核心引擎、消息总线、模块管理
2. **功能模块层**：内存优化、网络管理、密码管理、文件监控等
3. **服务层**：网络设置服务、密码管理服务
4. **界面层**：系统托盘、网络管理GUI、密码管理Web界面

## 模块说明

### 1. 主项目（xyrTools）
- **core/**：核心引擎，负责消息分发与处理
- **modules/**：功能模块集合
  - **memoryOptimizer/**：内存优化模块
  - **netManage/**：网络管理模块
  - **passwordManage/**：密码管理模块
  - **fileMonitor/**：文件监控模块【暂未实现】
  - **tray/**：系统托盘模块
- **extendFunc/**：扩展功能，包含日志、网络工具等
- **config/**：配置文件目录

### 2. 网络管理（netManageGui & netSetService）
- **netManageGui/**：网络管理GUI界面，基于Fyne框架
- **netSetService/**：网络设置服务，提供高权限网络配置功能，配置网卡需要高权限，因此写了个服务处理请求，GUI界面通过消息传递与服务交互。

### 3. 密码管理（passwordManageServer & passwordManageWeb）
- **passwordManageServer/**：密码管理服务器，提供安全的密码存储与管理
- **passwordManageWeb/**：密码管理Web界面，基于Vue框架

### 4. 基础模块（myMod）
- **consolemgr/**：控制台管理【暂未实现，无法解决控制台输出问题已取消】
- **consoleutil/**：控制台工具【暂未实现】
- **notify/**：系统通知

## 快速开始

### 环境要求
- Go 1.23+
- Node.js 14+（用于密码管理Web界面编译，vite）
- Windows 10+（主要支持平台，其他平台未测试）

### 安装与运行

#### 1. 编译主项目
```bash
cd xyrTools
go build -o xyrTools.exe main.go
```

#### 2. 运行网络设置服务
```bash
cd netSetService
go build -o netSetService.exe main.go 
# 以管理员权限运行
```

#### 3. 编译密码管理服务器
```bash
cd passwordManage/passwordManageServer
go build -o passwordManageServer.exe main.go
# 可单独使用，也可作为主程序的模块使用，因为是开发后硬接进来的，最好首次运行服务端交互界面初始化数据库
```

#### 4. 启动主程序
```bash
# 在xyrTools目录运行
./xyrTools.exe
```

## 配置说明

### 主配置文件
- **xyrTools/config/config.yaml**：主程序配置
- **xyrTools/config/netConfig.yaml**：网络管理配置

### 密码管理配置
- **passwordManage/passwordManageServer/conf/sysConf.toml**：密码管理服务器配置

### 网络管理配置
- **netManageGui/config/netConfig.yaml**：网络管理GUI配置

## 开发指南

### 模块开发
1. 在 `xyrTools/modules/` 目录下创建新模块目录
2. 实现 `modInterfaces/modInterface.go` 中定义的接口
3. 在 `core/coreEngine.go` 中注册新模块
4. 编译并测试模块功能

### 消息传递
- 发布消息：使用 `core.PublishMessage()` 发布消息
- 订阅消息：使用 `core.SubscribeMessage()` 订阅消息

## 技术栈

### 后端
- Go 1.23+
- SQLite（密码管理数据库）
- AES/RSA/ECDSA（加密算法）

### 前端
- Vue 3
- Fyne（Go GUI框架）
- Element Plus（UI组件库）

## 贡献
欢迎给出建议来改进这个项目。