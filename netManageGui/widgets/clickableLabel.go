package widgets

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type CustomButton struct {
	Clickable    widget.Clickable
	Text         string
	Width        unit.Dp
	Height       unit.Dp
	MaxWidth     unit.Dp
	MaxHeight    unit.Dp
	CornerRadius int
	BgColor      color.NRGBA
	TextColor    color.NRGBA
	TextSize     unit.Sp
	BorderColor  color.NRGBA
	BorderWidth  unit.Dp
	Padding      unit.Dp
	SizeAuto     bool // 是否自适应内容大小

	OnClick func()
}

// 构造函数
func NewCustomButton(text string) *CustomButton {
	return &CustomButton{
		Text:         text,
		Width:        unit.Dp(100),
		Height:       unit.Dp(30),
		MaxWidth:     unit.Dp(300), // 0，表示不限制最大宽度
		MaxHeight:    unit.Dp(30),  // 0，表示不限制最大高度
		CornerRadius: 4,
		BgColor:      color.NRGBA{R: 0x33, G: 0x99, B: 0xFF, A: 255},
		TextColor:    color.NRGBA{A: 255},
		TextSize:     unit.Sp(14),
		BorderColor:  color.NRGBA{R: 0, G: 0, B: 0, A: 255},
		BorderWidth:  unit.Dp(1),
		Padding:      unit.Dp(8),
		SizeAuto:     false, // 默认自适应
	}
}

// Layout 渲染函数
func (b *CustomButton) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	// 点击事件
	clicked := b.Clickable.Clicked(gtx)
	if clicked && b.OnClick != nil {
		b.OnClick()
	}

	// 最大尺寸限制
	if b.MaxWidth > 0 {
		max := gtx.Dp(b.MaxWidth)
		if gtx.Constraints.Max.X > max {
			gtx.Constraints.Max.X = max
		}
	}

	label := material.Body1(th, b.Text)
	label.TextSize = b.TextSize
	label.Color = b.TextColor

	// 自动布局
	if b.SizeAuto {
		return b.Clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
			return layout.UniformInset(b.Padding).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				dims := label.Layout(gtx)
				b.drawBackground(gtx, dims.Size)
				return layout.Center.Layout(gtx, label.Layout)
			})
		})
	}

	// 固定尺寸布局
	size := image.Pt(gtx.Dp(b.Width), gtx.Dp(b.Height))
	gtx.Constraints = layout.Exact(size)

	return b.Clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		b.drawBackground(gtx, size)
		return layout.Center.Layout(gtx, label.Layout)
	})
}

// 绘制背景 + 边框
func (b *CustomButton) drawBackground(gtx layout.Context, size image.Point) {
	rrect := clip.RRect{
		Rect: image.Rectangle{Max: size},
		NE:   b.CornerRadius,
		NW:   b.CornerRadius,
		SE:   b.CornerRadius,
		SW:   b.CornerRadius,
	}
	clipOp := rrect.Push(gtx.Ops)
	paint.Fill(gtx.Ops, b.BgColor)
	clipOp.Pop()

	if b.BorderWidth > 0 {
		stroke := clip.Stroke{
			Path:  rrect.Path(gtx.Ops),
			Width: float32(gtx.Dp(b.BorderWidth)),
		}.Op().Push(gtx.Ops)
		paint.Fill(gtx.Ops, b.BorderColor)
		stroke.Pop()
	}
}

// 一堆 Set 方法，用于链式调用
func (b *CustomButton) SetSize(w, h unit.Dp) *CustomButton {
	b.Width = w
	b.Height = h
	return b
}

func (b *CustomButton) SetCornerRadius(r int) *CustomButton {
	b.CornerRadius = r
	return b
}

func (b *CustomButton) SetBackgroundColor(c color.NRGBA) *CustomButton {
	b.BgColor = c
	return b
}

func (b *CustomButton) SetBorder(c color.NRGBA, width unit.Dp) *CustomButton {
	b.BorderColor = c
	b.BorderWidth = width
	return b
}

func (b *CustomButton) SetTextSize(s unit.Sp) *CustomButton {
	b.TextSize = s
	return b
}

func (b *CustomButton) SetTextColor(c color.NRGBA) *CustomButton {
	b.TextColor = c
	return b
}

func (b *CustomButton) SetPadding(p unit.Dp) *CustomButton {
	b.Padding = p
	return b
}

func (b *CustomButton) SetSizeAuto(auto bool) *CustomButton {
	b.SizeAuto = auto
	return b
}

func (b *CustomButton) SetOnClick(f func()) *CustomButton {
	b.OnClick = f
	return b
}
