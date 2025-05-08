package widgets

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

type ScrollDemo struct {
	List    layout.List
	Options []string
}

func NewScrollDemo(options []string) *ScrollDemo {
	return &ScrollDemo{
		Options: options,
		List:    layout.List{Axis: layout.Vertical},
	}
}

func (sd *ScrollDemo) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	// 限定最大高度
	maxHeight := gtx.Dp(unit.Dp(300))
	gtx.Constraints.Max.Y = maxHeight
	maxWidth := gtx.Dp(unit.Dp(320))
	gtx.Constraints.Max.X = maxWidth

	// 增加内边距
	return layout.UniformInset(unit.Dp(2)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return sd.List.Layout(gtx, len(sd.Options), func(gtx layout.Context, index int) layout.Dimensions {
			return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return NewCustomButton(sd.Options[index]).Layout(gtx, th)
			})
		})
	})
}
