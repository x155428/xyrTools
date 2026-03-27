package widgets

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type Select struct {
	Options        []string // 所有选项
	Selected       int      // 当前选中项 index
	Expanded       bool     // 是否展开中
	Input          widget.Editor
	Click          widget.Clickable
	OptionList     []widget.Clickable
	Outside        widget.Clickable // 用于点外部关闭
	MaxVisibleOpts int              // 限制最大显示的选项数
}

func (s *Select) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	// 点开收起主按钮
	if s.Click.Clicked(gtx) {
		s.Expanded = !s.Expanded
	}

	// 点外部收起
	if s.Expanded && s.Outside.Clicked(gtx) {
		s.Expanded = false
	}

	for len(s.OptionList) < len(s.Options) {
		s.OptionList = append(s.OptionList, widget.Clickable{})
	}

	return layout.Stack{}.Layout(gtx,
		// 最底层捕获外部点击区域
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			if s.Expanded {
				return s.Outside.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Dimensions{Size: gtx.Constraints.Max}
				})
			}
			return layout.Dimensions{}
		}),
		// 选择框 + 下拉框
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				// 选择框
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					label := s.Options[s.Selected]
					return material.Editor(th, &s.Input, label).Layout(gtx)
					//return material.Button(th, &s.Click, label).Layout(gtx)

				}),
				// 下拉框
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if !s.Expanded {
						return layout.Dimensions{}
					}
					maxHeight := gtx.Metric.Dp(unit.Dp(48 * 4))
					if h := gtx.Constraints.Max.Y; h > maxHeight {
						gtx.Constraints.Max.Y = maxHeight
					}
					var dims layout.Dimensions
					list := layout.List{
						Axis: layout.Vertical,
					}
					dims = list.Layout(gtx, len(s.OptionList), func(gtx layout.Context, i int) layout.Dimensions {
						if s.OptionList[i].Clicked(gtx) {
							s.Selected = i
							s.Expanded = false
						}
						return material.Button(th, &s.OptionList[i], s.Options[i]).Layout(gtx)
					})

					return dims
				}),
			)
		}),
	)
}
