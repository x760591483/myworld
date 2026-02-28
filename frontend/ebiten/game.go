package ebiten

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/x760591483/myworld/core"
)

// Game 实现 ebiten.Game 接口，把 core.World 显示出来
type Game struct {
	World *core.World
}

// NewGame 创建一个绑定到给定世界的游戏对象
func NewGame(w *core.World) *Game {
	return &Game{World: w}
}

// Update 每帧调用，用于推进世界时间
func (g *Game) Update() error {
	if g.World == nil {
		return nil
	}
	// 这里先用固定 dt，后面可以改成根据实际帧间隔
	const dt = 1.0 / 60.0
	g.World.Tick(dt)
	return nil
}

// Draw 把世界画到屏幕
func (g *Game) Draw(screen *ebiten.Image) {
	if g.World == nil {
		return
	}

	screen.Fill(color.RGBA{R: 0, G: 0, B: 0, A: 255})

	for _, c := range g.World.AllCreatures() {
		if c == nil {
			continue
		}
		// 暂时用一个小矩形代表圆形身体，后面再细化真正画圆
		sx := float64(320) + c.X // 简单把世界坐标平移到屏幕中心附近
		sy := float64(240) + c.Y
		size := c.Radius * 2
		ebitenutil.DrawRect(screen, sx-size/2, sy-size/2, size, size, color.RGBA{
			R: c.Color.R,
			G: c.Color.G,
			B: c.Color.B,
			A: 255,
		})
	}
}

// Layout 指定逻辑屏幕尺寸
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 640, 480
}
