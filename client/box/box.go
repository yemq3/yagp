package box

import (
	"image"
)

// RelativeBox 神经网络返回的原始框数据，其中坐标都是0到1的float
type RelativeBox struct {
	X1   float64
	Y1   float64
	X2   float64
	Y2   float64
	Conf float64
	Name string
}

// AbsoluteBox 坐标用整数表示的AbsoluteBox
type AbsoluteBox struct {
	Rect image.Rectangle
	Conf float64
	Name string
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

func NewRelativeBox(rect image.Rectangle, width, height int) RelativeBox {
	return RelativeBox{}
}

func (b *RelativeBox) ToRetangle(width, height int) image.Rectangle {
	x0 := max(int(b.X1*float64(width)), 0)
	y0 := max(int(b.Y1*float64(height)), 0)
	x1 := min(int(b.X2*float64(width)), width)
	y1 := min(int(b.Y2*float64(height)), height)
	rect := image.Rect(x0, y0, x1, y1)

	return rect
}

func (b *RelativeBox) ToAbsoluteBox(width, height int) AbsoluteBox {
	rect := b.ToRetangle(width, height)

	return AbsoluteBox{
		Rect: rect,
		Conf: b.Conf,
		Name: b.Name,
	}
}

func (b *AbsoluteBox) ToRelativeBox(width, height int) RelativeBox {
	box := RelativeBox{
		X1:   float64(b.Rect.Min.X) / float64(width),
		Y1:   float64(b.Rect.Min.Y) / float64(height),
		X2:   float64(b.Rect.Max.X) / float64(width),
		Y2:   float64(b.Rect.Max.Y) / float64(height),
		Conf: b.Conf,
		Name: b.Name,
	}
	return box
}

