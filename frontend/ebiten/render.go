package ebiten

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/x760591483/myworld/core"
)

// drawFilledCircle 在 dst 上以 (cx,cy) 为圆心、r 为半径画一个实心圆
func drawFilledCircle(dst *ebiten.Image, cx, cy, r float32, clr color.RGBA) {
	// dst为目标图像，cx/cy为圆心坐标，r为半径，clr为颜色，最后一个参数true表示填充
	vector.DrawFilledCircle(dst, cx, cy, r, clr, true)
}

// DrawCreature 把一个生物绘制到屏幕上
// offsetX/offsetY 是世界坐标到屏幕坐标的平移量（摄像机偏移）
func DrawCreature(screen *ebiten.Image, c *core.Creature) {
	if c == nil {
		return
	}

	// 屏幕坐标（圆心）
	cx := float32(c.X)
	cy := float32(c.Y)
	r := float32(c.Radius)
	// fmt.Printf("Drawing creature ID=%d at (%.2f, %.2f) with radius %.2f\n", c.ID, cx, cy, r)
	// 1. 画身体（实心圆）
	bodyColor := color.RGBA{
		R: c.Color.R,
		G: c.Color.G,
		B: c.Color.B,
		A: 255,
	}
	drawFilledCircle(screen, cx, cy, r, bodyColor)

	// 2. 画两只眼睛
	// 两眼分别在朝向两侧，以 Direction 为中轴，各偏转 EyeAngle/2
	eyeAngles := [2]float64{
		c.Direction + c.EyeAngle/2, // 左眼
		c.Direction - c.EyeAngle/2, // 右眼
	}
	eyeR := float32(c.EyeRadius)
	pupilR := float32(c.PupilRadius)

	for _, angle := range eyeAngles {
		// 眼睛圆心位置
		ex := cx + float32(math.Cos(angle))*float32(c.EyeOffset)
		ey := cy - float32(math.Sin(angle))*float32(c.EyeOffset) // 屏幕Y轴向下，取反

		// 眼白（白色）
		drawFilledCircle(screen, ex, ey, eyeR, color.RGBA{R: 255, G: 255, B: 255, A: 255})

		// 瞳孔（黑色），瞳孔略偏向朝向方向，使眼睛有"注视感"
		px := ex + float32(math.Cos(angle))*pupilR*0.4
		py := ey - float32(math.Sin(angle))*pupilR*0.4
		drawFilledCircle(screen, px, py, pupilR, color.RGBA{R: 0, G: 0, B: 0, A: 255})
	}
}
