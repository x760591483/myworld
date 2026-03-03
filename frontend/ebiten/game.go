package ebiten

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/x760591483/myworld/assets"
	"github.com/x760591483/myworld/core"
)

var treeImage1 *ebiten.Image

func loadTreeImage(in []byte) *ebiten.Image {
	img, _, err := image.Decode(bytes.NewReader(in))
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
	return ebiten.NewImageFromImage(white)
}

func init() {
	treeImage1 = loadTreeImage(assets.Tree1PNG)
}

const avatarSize = 36 // 头像区域边长（像素）

// Game 实现 ebiten.Game 接口，把 core.World 显示出来
type Game struct {
	World *core.World

	// 选中的实体（互斥，至多一个非 nil）
	selectedCreature *core.Creature
	selectedPlant    *core.Plant
	// 信息板在屏幕上的固定位置
	panelX, panelY int
	// 实时头像离屏图（每帧重绘）
	avatarImg *ebiten.Image
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

	// ── 鼠标点击检测 ────────────────────────────────
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		// 屏幕坐标 → 世界坐标（目前 1:1 映射，无 Camera 偏移）
		wx, wy := float64(mx), float64(my)
		c, p := g.World.FindEntityAt(wx, wy)
		if c != nil {
			g.selectedCreature = c
			g.selectedPlant = nil
			g.panelX, g.panelY = mx+10, my+10 // 信息板显示在点击位置右下方
		} else if p != nil {
			g.selectedCreature = nil
			g.selectedPlant = p
			g.panelX, g.panelY = mx+10, my+10
		} else {
			// 点击空白处，取消选中
			g.selectedCreature = nil
			g.selectedPlant = nil
		}
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

	for _, p := range g.World.AllPlants() {
		if p == nil {
			continue
		}
		// 使用与生物相同的坐标映射，直接用世界坐标
		sx := p.X
		sy := p.Y
		size := p.Radius * 2
		scale := size / 128.0
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(sx-size/2, sy-size/2)
		// 固定染成绿色（R=0, G=1, B=0），参数范围 0.0~1.0
		op.ColorScale.Scale(0, 1, 0, 1)
		screen.DrawImage(treeImage1, op)
	}
	for _, c := range g.World.AllCreatures() {
		DrawCreature(screen, c)
	}

	// ── 绘制信息板（固定在屏幕位置，不跟随世界移动）──────
	if g.selectedCreature != nil || g.selectedPlant != nil {
		// 每帧实时绘制头像
		if g.avatarImg == nil {
			g.avatarImg = ebiten.NewImage(avatarSize, avatarSize)
		}
		g.avatarImg.Clear()

		if g.selectedCreature != nil {
			DrawCreatureAvatar(g.avatarImg, g.selectedCreature)
			DrawCreatureInfoPanel(screen, g.selectedCreature, g.panelX, g.panelY, g.avatarImg)
		} else {
			DrawPlantAvatar(g.avatarImg, g.selectedPlant)
			DrawPlantInfoPanel(screen, g.selectedPlant, g.panelX, g.panelY, g.avatarImg)
		}
	}
}

// Layout 指定逻辑屏幕尺寸
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return int(g.World.GetWidth()), int(g.World.GetHeight())
}
