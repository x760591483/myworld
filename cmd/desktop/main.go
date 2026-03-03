package main

// 桌面程序：先不用 Ebiten，只在终端中测试 core.World 和 Tick

import (
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/x760591483/myworld/core"
	frontend "github.com/x760591483/myworld/frontend/ebiten"
)

func main() {
	// 创建世界
	w := core.NewWorld()

	// 创建一个生物，放在世界中心，赋予向右的速度
	// id := w.NextID()
	// creature := core.NewCreature(id, 0, 0, 10, nil)
	// creature.VelocityX = 50
	// creature.VelocityY = 0
	// w.AddCreature(creature)

	// // 创建一些植物，随机分布在世界中
	// for i := 0; i < 10; i++ {
	// 	id := w.NextID()
	// 	plant := core.NewPlant(id, float64(i*30-150), float64(i*20-100), 15)
	// 	w.AddPlant(plant)
	// }

	// 设置 Ebiten 窗口参数
	ebiten.SetWindowSize(int(w.GetWidth()), int(w.GetHeight()))
	ebiten.SetWindowTitle("myworld - simple creature")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled) // 允许调整窗口大小
	ebiten.SetTPS(60)

	// 运行游戏
	game := frontend.NewGame(w)
	if err := ebiten.RunGame(game); err != nil && !isRegularTermination(err) {
		log.Fatal(err)
	}
}

// isRegularTermination 用于以后扩展优雅退出，目前简单返回 false
func isRegularTermination(err error) bool {
	_ = math.NaN() // 暂时避免未使用导入，后续可删除 math 相关代码
	return false
}
