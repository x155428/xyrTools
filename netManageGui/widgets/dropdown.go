package widgets

//  select 组件封装
import (
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type DropDownInput struct {
	Input         Input              // 输入框
	IconButton    widget.Clickable   // 图标按钮
	Scroll        ScrollDemo         // 滚动组件
	Outside       widget.Clickable   // 用于捕获外部点击事件
	Options       []string           // 选项列表
	OptionClicks  []widget.Clickable // 选项点击事件
	Expanded      bool               // 是否展开下拉框
	SelectedIndex int                // 当前选中的选项索引
	List          layout.List        // 列表布局
	ListHight     int                // 列表高度
}

func NewDropDownInput(options []string) *DropDownInput {
	return &DropDownInput{
		Input:        *NewInput("请选择", 100, 20, true),        // 初始化输入框
		Scroll:       *NewScrollDemo(options),                // 初始化滚动组件
		Options:      options,                                // 初始化选项列表
		OptionClicks: make([]widget.Clickable, len(options)), // 初始化选项点击事件
		Expanded:     false,                                  // 初始状态为收起
		List:         layout.List{Axis: layout.Vertical},     // 初始化列表布局
		ListHight:    200,                                    // 初始列表高度为0
	}
}

func (dd *DropDownInput) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	// 如果按钮被点击，切换展开状态
	if dd.IconButton.Clicked(gtx) {
		//打印点击事件
		println("展开按钮被点击了")
		dd.Expanded = !dd.Expanded
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			// 输入框和图标按钮布局
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return dd.Input.Layout(gtx, th)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return dd.IconButton.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return NewCustomButton("▼").Layout(gtx, th)
					})
				}),
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return dd.Scroll.Layout(gtx, th)
		}),
	)
}
