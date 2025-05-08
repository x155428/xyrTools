package widgets

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// Input 组件
type Input struct {
	Text         string           // 文本内容
	Clickable    widget.Clickable // 点击事件
	Editor       widget.Editor    // 内部输入框
	Placeholder  string           // 占位符文本，提示词
	Width        unit.Dp          // 宽度
	Height       unit.Dp          // 高度
	CornerRadius int              // 圆角半径
	BgColor      color.NRGBA      // 背景颜色
	BorderColor  color.NRGBA      // 边框颜色
	BorderWidth  unit.Dp          // 边框宽度
	TextSize     unit.Sp          // 文本大小

}

// 构造函数：初始化 Input 组件
func NewInput(placeholder string, width, height unit.Dp, autoResize bool) *Input {
	return &Input{
		Editor: widget.Editor{
			SingleLine: true,
		}, // 输入框
		Placeholder:  placeholder,  // 占位符文本
		Width:        unit.Dp(300), // 宽度
		Height:       unit.Dp(20),  // 高度
		CornerRadius: 0,
		BgColor:      color.NRGBA{R: 0x33, G: 0x99, B: 0xFF, A: 255},
		BorderColor:  color.NRGBA{R: 0, G: 0, B: 0, A: 255},
		BorderWidth:  unit.Dp(1),
		TextSize:     unit.Sp(12),
	}
}

// Layout 渲染输入框
// 布局
func (i *Input) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	// 固定整体尺寸
	gtx.Constraints = layout.Exact(image.Pt(gtx.Dp(i.Width), gtx.Dp(i.Height)))

	var editorDim layout.Dimensions

	// 先布局 editor，获取实际尺寸
	macro := op.Record(gtx.Ops)
	editor := material.Editor(th, &i.Editor, i.Placeholder)
	editor.TextSize = i.TextSize
	editor.Font.Weight = 12
	editorDim = layout.UniformInset(unit.Dp(1)).Layout(gtx, editor.Layout)
	call := macro.Stop()

	return layout.Stack{}.Layout(gtx,
		// 背景 + 边框，尺寸用 editorDim
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			i.drawBackground(gtx, editorDim.Size)
			return editorDim
		}),
		// 输入框
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			call.Add(gtx.Ops)
			return editorDim
		}),
	)
}

// 背景 + 边框
func (i *Input) drawBackground(gtx layout.Context, size image.Point) {
	rrect := clip.RRect{
		Rect: image.Rectangle{Max: size},
		SE:   i.CornerRadius, SW: i.CornerRadius,
		NE: i.CornerRadius, NW: i.CornerRadius,
	}

	// 背景
	bgClip := rrect.Push(gtx.Ops)
	paint.Fill(gtx.Ops, i.BgColor)
	bgClip.Pop()

	// 边框
	if i.BorderWidth > 0 {
		stroke := clip.Stroke{
			Path:  rrect.Path(gtx.Ops),
			Width: float32(gtx.Dp(i.BorderWidth)),
		}
		borderOp := stroke.Op().Push(gtx.Ops)
		paint.Fill(gtx.Ops, i.BorderColor)
		borderOp.Pop()
	}
}

// 可选配置
func (i *Input) SetSize(w, h unit.Dp) *Input {
	i.Width = w
	i.Height = h
	return i
}

func (i *Input) SetTextSize(s unit.Sp) *Input {
	i.TextSize = s
	return i
}

func (i *Input) SetBackgroundColor(c color.NRGBA) *Input {
	i.BgColor = c
	return i
}

func (i *Input) SetBorder(c color.NRGBA, width unit.Dp) *Input {
	i.BorderColor = c
	i.BorderWidth = width
	return i
}

func (i *Input) SetCornerRadius(r int) *Input {
	i.CornerRadius = r
	return i
}
