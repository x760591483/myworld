package ebiten

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
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

// ── 信息板绘制 ────────────────────────────────────────────

// drawInfoPanel 在屏幕固定位置绘制一个半透明背景面板，逐行输出文本，
// 并将 avatarImg（可为 nil）贴到面板右上角。
func drawInfoPanel(screen *ebiten.Image, lines []string, px, py int, avatarImg *ebiten.Image) {
	const (
		charW        = 6  // ebitenutil.DebugPrint 每字符宽 6px
		lineH        = 16 // 行高
		padX         = 8  // 水平内边距
		padY         = 6  // 垂直内边距
		avatarMargin = 6  // 头像与文字区域的间距
	)

	// 计算文字区域尺寸
	maxLen := 0
	for _, l := range lines {
		if len(l) > maxLen {
			maxLen = len(l)
		}
	}
	textW := maxLen*charW + padX*2

	// 头像区域宽度
	avatarW := 0
	if avatarImg != nil {
		avatarW = avatarImg.Bounds().Dx() + avatarMargin*2
	}
	panelW := textW + avatarW
	panelH := len(lines)*lineH + padY*2
	if avatarImg != nil {
		if minH := avatarImg.Bounds().Dy() + avatarMargin*2 + padY*2; panelH < minH {
			panelH = minH
		}
	}

	// 确保面板不超出屏幕右/下边界
	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()
	if px+panelW > sw {
		px = sw - panelW
	}
	if py+panelH > sh {
		py = sh - panelH
	}
	if px < 0 {
		px = 0
	}
	if py < 0 {
		py = 0
	}

	// 半透明黑色背景
	vector.DrawFilledRect(screen,
		float32(px), float32(py),
		float32(panelW), float32(panelH),
		color.RGBA{R: 0, G: 0, B: 0, A: 180}, true)

	// 逐行输出文字（左侧）
	for i, line := range lines {
		ebitenutil.DebugPrintAt(screen, line, px+padX, py+padY+i*lineH)
	}

	// 头像贴到面板右上角
	if avatarImg != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(px+textW+avatarMargin), float64(py+padY))
		screen.DrawImage(avatarImg, op)
	}
}

// DrawCreatureInfoPanel 绘制生物信息板（右上角显示实时头像）
func DrawCreatureInfoPanel(screen *ebiten.Image, c *core.Creature, px, py int, avatarImg *ebiten.Image) {
	if c == nil {
		return
	}
	lines := []string{
		"=== Creature ===",
		fmt.Sprintf("ID:        %d", c.ID),
		fmt.Sprintf("Pos:       (%.1f, %.1f)", c.X, c.Y),
		fmt.Sprintf("Radius:    %.1f", c.Radius),
		fmt.Sprintf("Health:    %d / %d", c.Health, c.MaxHealth),
		fmt.Sprintf("Mass:      %.1f", c.Mass),
		fmt.Sprintf("Speed:     %.1f", c.Speed),
		fmt.Sprintf("Velocity:  (%.1f, %.1f)", c.VelocityX, c.VelocityY),
		fmt.Sprintf("Direction: %.2f rad", c.Direction),
		fmt.Sprintf("Age:       %d ticks", c.Age),
		fmt.Sprintf("Color:     (%d,%d,%d)", c.Color.R, c.Color.G, c.Color.B),
	}
	drawInfoPanel(screen, lines, px, py, avatarImg)
}

// DrawPlantInfoPanel 绘制植物信息板（右上角显示实时头像）
func DrawPlantInfoPanel(screen *ebiten.Image, p *core.Plant, px, py int, avatarImg *ebiten.Image) {
	if p == nil {
		return
	}
	lines := []string{
		"=== Plant ===",
		fmt.Sprintf("ID:          %d", p.ID),
		fmt.Sprintf("Pos:         (%.1f, %.1f)", p.X, p.Y),
		fmt.Sprintf("Radius:      %.1f", p.Radius),
		fmt.Sprintf("Health:      %d / %d", p.Health, p.MaxHealth),
		fmt.Sprintf("Mass:        %.1f", p.Mass),
		fmt.Sprintf("GrowthStage: %d", p.GrowthStage),
		fmt.Sprintf("Age:         %d ticks", p.Age),
		fmt.Sprintf("Color:       (%d,%d,%d)", p.Color.R, p.Color.G, p.Color.B),
	}
	drawInfoPanel(screen, lines, px, py, avatarImg)
}

// ── 头像绘制（每帧从指针实时读取状态，保证头像动态更新）───────

// DrawCreatureAvatar 在离屏图上绘制生物实时缩略画像（含眼睛朝向）。
func DrawCreatureAvatar(img *ebiten.Image, c *core.Creature) {
	if c == nil {
		return
	}
	size := float32(img.Bounds().Dx())
	maxR := (size - 8) / 2
	scale := maxR / float32(c.Radius)

	cx := size / 2
	cy := size / 2
	r := float32(c.Radius) * scale

	// 身体
	drawFilledCircle(img, cx, cy, r, color.RGBA{R: c.Color.R, G: c.Color.G, B: c.Color.B, A: 255})

	// 眼睛（保持实时 Direction）
	eyeAngles := [2]float64{
		c.Direction + c.EyeAngle/2,
		c.Direction - c.EyeAngle/2,
	}
	eyeR := float32(c.EyeRadius) * scale
	pupilR := float32(c.PupilRadius) * scale
	offset := float32(c.EyeOffset) * scale

	for _, angle := range eyeAngles {
		ex := cx + float32(math.Cos(angle))*offset
		ey := cy - float32(math.Sin(angle))*offset
		drawFilledCircle(img, ex, ey, eyeR, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		px := ex + float32(math.Cos(angle))*pupilR*0.4
		py := ey - float32(math.Sin(angle))*pupilR*0.4
		drawFilledCircle(img, px, py, pupilR, color.RGBA{R: 0, G: 0, B: 0, A: 255})
	}
}

// DrawPlantAvatar 在离屏图上绘制植物实时缩略画像。
func DrawPlantAvatar(img *ebiten.Image, p *core.Plant) {
	if p == nil {
		return
	}
	size := float64(img.Bounds().Dx())
	imgScale := (size - 8) / 128.0
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(imgScale, imgScale)
	op.GeoM.Translate(4, 4)
	op.ColorScale.Scale(
		float32(p.Color.R)/255.0,
		float32(p.Color.G)/255.0,
		float32(p.Color.B)/255.0,
		1,
	)
	img.DrawImage(treeImage1, op)
}
