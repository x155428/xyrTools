package main

import (
	"net"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type NetConfig struct {
	Name     string   `yaml:"name"`
	Desc     string   `yaml:"desc"`
	Adapter  string   `yaml:"adapter"`
	DHCP     bool     `yaml:"dhcp"`
	DNSDHCP  bool     `yaml:"dnsdhcp"`
	IP       string   `yaml:"ip"`
	Netmask  string   `yaml:"netmask"`
	Gateway  string   `yaml:"gateway"`
	DNS      []string `yaml:"dns"`
	MTU      int      `yaml:"mtu"`
	Metric   int      `yaml:"metric"`
	FlushDNS bool     `yaml:"flushDNS"`
}

type ConfigFile struct {
	Configs []NetConfig `yaml:"configs"`
}

// 配置表单控件结构体
type ConfigForm struct {
	CfgName       *widget.Entry
	DescEntry     *widget.Entry
	AdapterSelect *widget.Select
	DhcpCheck     *widget.Check
	DnsdhcpCheck  *widget.Check
	IpEntry       *widget.Entry
	MaskEntry     *widget.Entry
	GwEntry       *widget.Entry
	DnsEntry      *widget.Entry
	MtuEntry      *widget.Entry
	MetricEntry   *widget.Entry
	FlushCheck    *widget.Check
}

var selected *NetConfig   // 选中的配置
var selectedIndex int = 0 // 选中的配置索引
var path string           // 配置文件路径

func main() {
	// 获取当前项目路径
	//打包用此
	dir, err := os.Getwd()
	if err != nil {
		//fmt.Println("获取当前项目路径失败:", err)
		return
	}
	path = strings.Join([]string{dir, "config", "netConfig.yaml"}, "/")
	//path = "D:/go/workdir/src/xyrTools/test/config/netConfig.yaml"
	cfg, err := loadConfig(path)
	if err != nil {
		//fmt.Println("配置加载失败:", err)
		return
	}
	myApp := app.New()
	myWin := myApp.NewWindow("网卡配置管理器")
	myWin.Resize(fyne.NewSize(800, 600))

	//左侧配置列表控件
	cfgNameList := widget.NewList(
		func() int { return len(cfg.Configs) },                  // 列表项数量
		func() fyne.CanvasObject { return widget.NewLabel("") }, // 创建列表项
		func(i widget.ListItemID, o fyne.CanvasObject) { // 列表项内容
			o.(*widget.Label).SetText(cfg.Configs[i].Name)
		},
	)

	// 配置表单控件结构体
	cfgDetailsForm := NewConfigForm()
	// 提交时刷新
	// cfgDetailsForm.CfgName.OnSubmitted = func(s string) {
	// 	selected.Name = s
	// 	cfgNameList.Refresh()
	// }

	// 配置表单布局
	form := container.NewVBox(
		widget.NewLabel("配置名："), cfgDetailsForm.CfgName,
		widget.NewLabel("描述："), cfgDetailsForm.DescEntry,
		widget.NewLabel("网卡："), cfgDetailsForm.AdapterSelect,
		cfgDetailsForm.DhcpCheck,
		cfgDetailsForm.DnsdhcpCheck,
		widget.NewLabel("IP 地址："), cfgDetailsForm.IpEntry,
		widget.NewLabel("子网掩码："), cfgDetailsForm.MaskEntry,
		widget.NewLabel("网关："), cfgDetailsForm.GwEntry,
		widget.NewLabel("DNS（逗号分隔）："), cfgDetailsForm.DnsEntry,
		widget.NewLabel("MTU："), cfgDetailsForm.MtuEntry,
		widget.NewLabel("Metric："), cfgDetailsForm.MetricEntry,
		cfgDetailsForm.FlushCheck,
	)

	//######################################################################
	//#############################  控件监听事件  ##########################
	//配置名选中事件处理函数
	cfgNameList.OnSelected = func(id widget.ListItemID) {
		// 应用配置
		applyChanges(cfgDetailsForm)
		// 更新表单数据
		updateForm(&cfg.Configs[id], cfgDetailsForm)
		selectedIndex = id
		//fmt.Print("选中的配置索引：", selectedIndex)
		cfgNameList.Refresh()

	}
	cfgNameList.OnUnselected = func(id widget.ListItemID) {
		// 失焦时、内容写到数据结构中
		//fmt.Print("失焦的配置索引：", id)
		if id < 0 || id >= len(cfg.Configs) {
			// 如果配置项已删除，不操作
			//fmt.Print("已删，不操作")
			return
		}
		// 应用配置
		applyChanges(cfgDetailsForm)
		cfg.Configs[id] = *selected

	}

	cfgDetailsForm.DhcpCheck.OnChanged = func(b bool) {
		//fmt.Println("DHCP 状态改变：", b)
		if b {
			cfgDetailsForm.IpEntry.Disable()
			cfgDetailsForm.MaskEntry.Disable()
			cfgDetailsForm.GwEntry.Disable()
		} else {
			cfgDetailsForm.IpEntry.Enable()
			cfgDetailsForm.MaskEntry.Enable()
			cfgDetailsForm.GwEntry.Enable()
		}

	}

	cfgDetailsForm.DnsdhcpCheck.OnChanged = func(b bool) {
		//fmt.Println("DNS DHCP 状态改变：", b)
		if b {
			cfgDetailsForm.DnsEntry.Disable()
		} else {
			cfgDetailsForm.DnsEntry.Enable()
		}
	}

	// 右侧配置区域容器
	cfgDetails := container.NewBorder(
		widget.NewLabel("配置详情"), nil, nil, nil,
		container.NewVScroll(form),
	)
	// 左侧配置编辑按钮区域

	cfgManageBtnContainer := container.NewHBox(
		widget.NewButton("增加", func() { addCfgBtnClick(cfg, cfgNameList, cfgDetailsForm) }),
		widget.NewButton("删除", func() { delCfgBtnClick(cfg, cfgNameList, cfgDetailsForm) }),
		widget.NewButton("向前插入", func() { addCfgBtnBeforeClick(cfg, cfgNameList, cfgDetailsForm) }),
	)
	fixedArea := container.NewVBox(cfgManageBtnContainer)
	// 右侧配置按钮区域
	cfgDetailsBtnContainer := container.NewHBox(
		layout.NewSpacer(),
		widget.NewButton("保存", func() { saveCfgBtnClick(cfg, path) }),
		//widget.NewButton("取消", cancelCfgBtnClick),
	)
	// 左侧组合容器
	left := container.NewBorder(fixedArea, nil, nil, nil, cfgNameList)
	// 右侧组合容器
	right := container.NewBorder(nil, cfgDetailsBtnContainer, nil, nil, cfgDetails)

	mainSplit := container.NewHSplit(left, right)
	mainSplit.Offset = 0.2

	// 如果配置长度大于0，选中第一个配置
	if len(cfg.Configs) > 0 {
		cfgNameList.Select(selectedIndex) // 选中第一个配置项
	} else {
		// 如果配置长度为0，添加一个默认配置
		addCfgBtnClick(cfg, cfgNameList, cfgDetailsForm)
	}

	cfgNameList.Refresh()
	myWin.SetContent(mainSplit)
	myWin.ShowAndRun()
}

// 获取表单控件的初始化函数
func NewConfigForm() *ConfigForm {
	interfaceList := getInterfaces() // 获取网卡列表

	return &ConfigForm{
		CfgName:       widget.NewEntry(),                    // 配置名输入框
		DescEntry:     widget.NewEntry(),                    // 描述输入框
		AdapterSelect: widget.NewSelect(interfaceList, nil), // 网卡选择框
		DhcpCheck:     widget.NewCheck("DHCP", nil),         // DHCP 复选框
		DnsdhcpCheck:  widget.NewCheck("DNS DHCP", nil),     // DNS DHCP 复选框
		IpEntry:       widget.NewEntry(),                    // IP 地址输入框
		MaskEntry:     widget.NewEntry(),                    // 子网掩码输入框
		GwEntry:       widget.NewEntry(),                    // 网关输入框
		DnsEntry:      widget.NewEntry(),                    // DNS 输入框（逗号分隔）
		MtuEntry:      widget.NewEntry(),                    // MTU 输入框
		MetricEntry:   widget.NewEntry(),                    // Metric 输入框
		FlushCheck:    widget.NewCheck("Flush DNS", nil),    // Flush DNS 复选框
	}
}

func parseDNS(dnsStr string) []string {
	//fmt.Print("解析DNS：", dnsStr)
	var res []string
	parts := strings.Split(dnsStr, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			res = append(res, p)
		}
	}
	return res
}

// 获取网卡列表
func getInterfaces() []string {
	// 获取网卡列表
	ifaces, err := net.Interfaces()
	if err != nil {
		// 处理错误
		panic(err)
	}

	// 将网卡列表转为字符串数组
	var interfaceNames []string
	for _, iface := range ifaces {
		interfaceNames = append(interfaceNames, iface.Name)
	}
	return interfaceNames
}

// 加载配置文件
func loadConfig(path string) (*ConfigFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg ConfigFile
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func applyChanges(cfgDetailsForm *ConfigForm) {
	if selected == nil {
		return
	}
	selected.Name = cfgDetailsForm.CfgName.Text
	selected.Desc = cfgDetailsForm.DescEntry.Text
	selected.Adapter = cfgDetailsForm.AdapterSelect.Selected
	selected.DHCP = cfgDetailsForm.DhcpCheck.Checked
	selected.DNSDHCP = cfgDetailsForm.DnsdhcpCheck.Checked
	selected.IP = cfgDetailsForm.IpEntry.Text
	selected.Netmask = cfgDetailsForm.MaskEntry.Text
	selected.Gateway = cfgDetailsForm.GwEntry.Text
	selected.DNS = parseDNS(cfgDetailsForm.DnsEntry.Text)
	selected.MTU = parseInt(cfgDetailsForm.MtuEntry.Text, 1500)    // 默认值 1500
	selected.Metric = parseInt(cfgDetailsForm.MetricEntry.Text, 0) // 默认值 0
	selected.FlushDNS = cfgDetailsForm.FlushCheck.Checked
}

// 更新表单数据函数
func updateForm(c *NetConfig, cfgDetailsForm *ConfigForm) {
	selected = c
	cfgDetailsForm.CfgName.SetText(c.Name)
	cfgDetailsForm.DescEntry.SetText(c.Desc)
	cfgDetailsForm.AdapterSelect.SetSelected(c.Adapter)
	cfgDetailsForm.DhcpCheck.SetChecked(c.DHCP)
	cfgDetailsForm.DnsdhcpCheck.SetChecked(c.DNSDHCP)
	cfgDetailsForm.IpEntry.SetText(c.IP)
	cfgDetailsForm.MaskEntry.SetText(c.Netmask)
	cfgDetailsForm.GwEntry.SetText(c.Gateway)
	cfgDetailsForm.DnsEntry.SetText(dnsListToString(c.DNS))
	cfgDetailsForm.MtuEntry.SetText(strconv.Itoa(c.MTU))
	cfgDetailsForm.MetricEntry.SetText(strconv.Itoa(c.Metric))
	cfgDetailsForm.FlushCheck.SetChecked(c.FlushDNS)
}

// 清空表单字段
func clearForm(cfgDetailsForm *ConfigForm) {
	cfgDetailsForm.CfgName.SetText("")
	cfgDetailsForm.DescEntry.SetText("")
	cfgDetailsForm.AdapterSelect.SetSelected("")
	cfgDetailsForm.DhcpCheck.SetChecked(false)
	cfgDetailsForm.DnsdhcpCheck.SetChecked(false)
	cfgDetailsForm.IpEntry.SetText("")
	cfgDetailsForm.MaskEntry.SetText("")
	cfgDetailsForm.GwEntry.SetText("")
	cfgDetailsForm.DnsEntry.SetText("")
	cfgDetailsForm.MtuEntry.SetText("")
	cfgDetailsForm.MetricEntry.SetText("")
	cfgDetailsForm.FlushCheck.SetChecked(false)
}

// 在指定索引前插入一个元素
func InsertBefore[T any](slice *[]T, index int, value T) {
	s := *slice
	if index < 0 || index > len(s) {
		panic("插入索引 " + strconv.Itoa(index) + " 超出范围[0, " + strconv.Itoa(len(s)) + "]")
	}
	if len(s) < cap(s) { // 有剩余空间，直接插入
		s = s[:len(s)+1]
		copy(s[index+1:], s[index:])
		s[index] = value
	} else { // 空间不够，重新申请
		newSlice := make([]T, len(s)+1)
		copy(newSlice, s[:index])
		newSlice[index] = value
		copy(newSlice[index+1:], s[index:])
		s = newSlice
	}
	*slice = s
}

// 在指定索引后插入一个元素
func InsertAfter[T any](slice *[]T, index int, value T) {
	s := *slice
	if index < -1 || index >= len(s) {
		panic("InsertAfter: index " + strconv.Itoa(index) + " out of range [-1, " + strconv.Itoa(len(s)) + ")")
	}

	InsertBefore(slice, index+1, value)
}

// 增加按钮事件处理函数,向后增加
func addCfgBtnClick(cfg *ConfigFile, cfgNameList *widget.List, cfgDetailsForm *ConfigForm) {
	//fmt.Println("增加配置")
	// 清空表单字段

	// 创建一个新的空配置
	newCfg := NetConfig{
		Name:     "新配置",
		Desc:     "",
		Adapter:  "",
		DHCP:     false,
		DNSDHCP:  false,
		IP:       "",
		Netmask:  "",
		Gateway:  "",
		DNS:      nil,
		MTU:      1500,
		Metric:   0,
		FlushDNS: false,
	}
	// 检查selectedIndex是否超出范围
	if selectedIndex < 0 || selectedIndex >= len(cfg.Configs) {
		//fmt.Println("选中索引非法", selectedIndex)
		selectedIndex = -1 // 如果超出范围，将其设置为开头
	}
	// 将新的配置添加到配置列表中
	InsertAfter(&cfg.Configs, selectedIndex, newCfg)
	selectedIndex += 1

	// 更新配置列表显示
	cfgNameList.Refresh()

	// 自动选择新增的配置并更新表单
	cfgNameList.Select(selectedIndex)
}

// 增加按钮事件处理函数,向前增加
func addCfgBtnBeforeClick(cfg *ConfigFile, cfgNameList *widget.List, cfgDetailsForm *ConfigForm) {
	//fmt.Println("向前增加配置")
	// 清空表单字段

	// 创建一个新的空配置
	newCfg := NetConfig{
		Name:     "新配置",
		Desc:     "",
		Adapter:  "",
		DHCP:     false,
		DNSDHCP:  false,
		IP:       "",
		Netmask:  "",
		Gateway:  "",
		DNS:      nil,
		MTU:      1500,
		Metric:   0,
		FlushDNS: false,
	}

	// 将新的配置添加到配置列表中
	InsertBefore(&cfg.Configs, selectedIndex, newCfg)

	// 更新配置列表显示
	cfgNameList.Refresh()

	// 自动选择新增的配置并更新表单
	cfgNameList.Select(selectedIndex)
}

// 删除按钮事件处理函数
func delCfgBtnClick(cfg *ConfigFile, cfgNameList *widget.List, cfgDetailsForm *ConfigForm) {
	//fmt.Println("删除配置")

	// 边界检查
	if selectedIndex < 0 || selectedIndex >= len(cfg.Configs) {
		//fmt.Println("选中索引非法，无法删除")
		return
	}

	// 删除元素
	cfg.Configs = append(cfg.Configs[:selectedIndex], cfg.Configs[selectedIndex+1:]...)

	// 更新选中状态
	if len(cfg.Configs) == 0 {
		selectedIndex = -1
		selected = nil
		clearForm(cfgDetailsForm)
		cfgNameList.UnselectAll()
	} else {
		if selectedIndex >= len(cfg.Configs) {
			selectedIndex = len(cfg.Configs) - 1 // 如果删的是最后一个，选前一个
		}
		selected = &cfg.Configs[selectedIndex] // 更新 selected 指针
		updateForm(selected, cfgDetailsForm)
		cfgNameList.Select(selectedIndex)
	}
	saveCfgBtnClick(cfg, path)
	cfgNameList.Refresh()
}

// 保存按钮事件处理函数
func saveCfgBtnClick(cfg *ConfigFile, path string) {
	err := saveConfig(path, cfg)
	if err != nil {
		//fmt.Println("保存配置失败:", err)
		return
	}
	//fmt.Println("保存配置")
}

// 取消按钮事件处理函数
// func cancelCfgBtnClick() {
// 	//fmt.Println("取消配置")
// }

func saveConfig(path string, cfg *ConfigFile) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func dnsListToString(dnsList []string) string {
	var all []string
	for _, item := range dnsList {
		parts := strings.Fields(item) // 空格拆分
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				all = append(all, p)
			}
		}
	}
	return strings.Join(all, ", ")
}

// parseInt 函数处理字符串到整数的转换，提供默认值
func parseInt(s string, defaultValue int) int {
	// 尝试转换字符串为整数
	if val, err := strconv.Atoi(s); err == nil {
		return val
	}
	// 如果转换失败，则返回默认值
	return defaultValue
}
