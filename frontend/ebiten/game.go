package ebiten

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/x760591483/myworld/assets"
	"github.com/x760591483/myworld/core"
)

// treeSprite 保存一棵树的分层精灵：轮廓 + 填充遮罩 + 填充区域范围
type treeSprite struct {
	outline  *ebiten.Image // 白色轮廓图（中间空心），用 ColorScale 染色
	fillMask *ebiten.Image // 白色填充遮罩（仅中间空心区域有像素），用 SubImage 裁剪高度来表达不同填充比例
	fillMinY int           // 填充区域在图片中的最小 Y（顶部）
	fillMaxY int           // 填充区域在图片中的最大 Y（底部）
}

var treeSpr1 *treeSprite

// loadTreeSprite 从 PNG 字节加载树精灵，分离出轮廓图和填充遮罩。
// 轮廓图：所有不透明像素转白色，中间空心区域保持透明。
// 填充遮罩：仅中间空心区域为白色像素，其余透明。
// 绘制时：先画轮廓（染色），再画填充遮罩的 SubImage（按比例裁剪高度，染另一种色）。
func loadTreeSprite(in []byte) *treeSprite {
	img, _, err := image.Decode(bytes.NewReader(in))
	if err != nil {
		panic(err)
	}

	bounds := img.Bounds()
	outlineRGBA := image.NewRGBA(bounds)
	fillRGBA := image.NewRGBA(bounds)

	// 先把原图复制到 outline（保留 alpha）
	draw.Draw(outlineRGBA, bounds, img, bounds.Min, draw.Src)

	xCenter := (bounds.Min.X + bounds.Max.X) / 2
	fillMinY := bounds.Max.Y // 初始化为最大值，后续取 min
	fillMaxY := bounds.Min.Y // 初始化为最小值，后续取 max

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		step := 0
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, _, _, a := img.At(x, y).RGBA()

			// 轮廓图：所有不透明像素转白色
			if a > 0 {
				outlineRGBA.SetRGBA(x, y, color.RGBA{R: 255, G: 255, B: 255, A: uint8(a >> 8)})
			}

			// 状态机检测中间空心区域（与原逻辑一致）
			if r > 0 && step == 0 && x < xCenter {
				step = 1 // 遇到左边缘
			}
			if r == 0 && step == 1 && x <= xCenter {
				step = 2 // 进入中间空心
			}
			if step == 2 && r > 0 {
				step = 3 // 离开中间空心（右边缘）
			}

			if step == 2 {
				// 空心区域：轮廓图保持透明，填充遮罩设为白色
				outlineRGBA.SetRGBA(x, y, color.RGBA{R: 0, G: 0, B: 0, A: 0})
				fillRGBA.SetRGBA(x, y, color.RGBA{R: 255, G: 255, B: 255, A: 255})

				if y < fillMinY {
					fillMinY = y
				}
				if y > fillMaxY {
					fillMaxY = y
				}
			}
		}
	}

	// fillMaxY 是最后一行的 Y，实际范围应为 [fillMinY, fillMaxY+1)
	if fillMaxY >= fillMinY {
		fillMaxY++
	}

	fmt.Printf("tree sprite loaded: fill region Y=[%d, %d)\n", fillMinY, fillMaxY)

	return &treeSprite{
		outline:  ebiten.NewImageFromImage(outlineRGBA),
		fillMask: ebiten.NewImageFromImage(fillRGBA),
		fillMinY: fillMinY,
		fillMaxY: fillMaxY,
	}
}

func init() {
	treeSpr1 = loadTreeSprite(assets.Tree1PNG)
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

		// ── 1. 绘制轮廓图（用植物颜色染色）──
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(sx-size/2, sy-size/2)
		op.ColorScale.Scale(
			float32(p.Color.R)/255.0,
			float32(p.Color.G)/255.0,
			float32(p.Color.B)/255.0,
			1,
		)
		screen.DrawImage(treeSpr1.outline, op)

		// ── 2. 绘制填充遮罩（按 Energy/MaxEnergy 比例裁剪高度）──
		fillRatio := 0.0
		if p.MaxEnergy > 0 {
			fillRatio = p.Energy / p.MaxEnergy
			if fillRatio > 1 {
				fillRatio = 1
			}
			if fillRatio < 0 {
				fillRatio = 0
			}
		}
		if fillRatio > 0 {
			fillH := treeSpr1.fillMaxY - treeSpr1.fillMinY          // 填充区域总高度（像素）
			visibleH := int(float64(fillH) * fillRatio)              // 本次可见高度
			cropY := treeSpr1.fillMaxY - visibleH                    // 从底部向上填充
			sub := treeSpr1.fillMask.SubImage(image.Rect(
				0, cropY,
				treeSpr1.fillMask.Bounds().Dx(), treeSpr1.fillMaxY,
			)).(*ebiten.Image)

			opFill := &ebiten.DrawImageOptions{}
			opFill.GeoM.Scale(scale, scale)
			// SubImage 的 Bounds().Min 会变成 (0, cropY)，DrawImage 会自动从那个偏移开始绘制
			opFill.GeoM.Translate(sx-size/2, sy-size/2)
			// 填充颜色：比轮廓暗一些（乘以 0.5）或者用不同颜色表示饱满度
			opFill.ColorScale.Scale(
				float32(p.Color.R)/255.0*0.6,
				float32(p.Color.G)/255.0*0.6,
				float32(p.Color.B)/255.0*0.6,
				1,
			)
			screen.DrawImage(sub, opFill)
		}
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
