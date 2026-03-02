package ebiten

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/x760591483/myworld/assets"
	"github.com/x760591483/myworld/core"
)

var treeImage *ebiten.Image

func init() {
	img, _, err := image.Decode(bytes.NewReader(assets.Tree3PNG))
	if err != nil {
		panic(err)
	}
	// 将图片所有不透明像素转为白色，使 ColorScale 乘色能正确生效
	// （iconfont 下载的图标通常是黑色像素，黑×颜色=黑，必须先转白）
	bounds := img.Bounds()
	white := image.NewRGBA(bounds)
	draw.Draw(white, bounds, img, bounds.Min, draw.Src) // 先复制 alpha
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, a := img.At(x, y).RGBA()
			if a > 0 {
				white.SetRGBA(x, y, color.RGBA{R: 255, G: 255, B: 255, A: uint8(a >> 8)})
			}
		}
	}
	treeImage = ebiten.NewImageFromImage(white)
}

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

	screen.Fill(color.RGBA{R: 125, G: 125, B: 125, A: 255})

	for _, c := range g.World.AllCreatures() {
		if c == nil {
			continue
		}
		sx := float64(320) + c.X
		sy := float64(240) + c.Y
		size := c.Radius * 2
		scale := size / 128.0

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(sx-size/2, sy-size/2)
		// 固定染成绿色（R=0, G=1, B=0），参数范围 0.0~1.0
		op.ColorScale.Scale(0, 1, 0, 1)
		screen.DrawImage(treeImage, op)
	}
}

// Layout 指定逻辑屏幕尺寸
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 640, 480
}
