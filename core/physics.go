package core

import (
	"math"
)

// 边界反弹/钳制（单个生物，O(n)）
func resolveWorldBounds(c *Creature, w *World) {
	if c.X < 0 {
		c.X = 0
		c.VelocityX = -c.VelocityX // 反弹
	} else if c.X > w.GetWidth() {
		c.X = w.GetWidth()
		c.VelocityX = -c.VelocityX // 反弹
	}

	if c.Y < 0 {
		c.Y = 0
		c.VelocityY = -c.VelocityY // 反弹
	} else if c.Y > w.GetHeight() {
		c.Y = w.GetHeight()
		c.VelocityY = -c.VelocityY // 反弹
	}
}

// resolveCollisions 纯物理碰撞修正：阻止实体重叠，不处理任何游戏逻辑。
//
// 规则：
//  1. 生物 vs 生物：两者都反弹，沿碰撞轴分离。
//  2. 生物 vs 植物：只有生物反弹，植物静止不动。
//
// 其他逻辑（吃掉植物、受伤等）请在 rules.go 中处理，不要放在此处。
func resolveCollisions(w *World) {
	sh := w.SpatialIndex

	// ── 1. 生物 vs 生物 ──────────────────────────────────────
	// ForEachCreaturePair 保证每对只处理一次（a.ID < b.ID）
	sh.ForEachCreaturePair(w.Creatures, func(a, b *Creature) {
		dx := b.X - a.X
		dy := b.Y - a.Y
		distSq := dx*dx + dy*dy
		minDist := a.Radius + b.Radius

		if distSq >= minDist*minDist || distSq == 0 {
			return // 未重叠，跳过
		}

		dist := math.Sqrt(distSq)
		// 碰撞法线（从 a 指向 b 的单位向量）
		nx := dx / dist
		ny := dy / dist

		// dvn = (Vb - Va) · n，n 从 a 指向 b
		// dvn < 0：b 相对 a 朝法线负方向运动，即两者相互靠近
		// dvn > 0：两者已在分离，不需要处理
		dvn := (b.VelocityX-a.VelocityX)*nx + (b.VelocityY-a.VelocityY)*ny

		// 只在相互靠近时才反弹，已在分离则不处理
		if dvn < 0 {
			// 弹性碰撞：交换法线方向速度分量（质量相等简化版）
			// 若日后需要质量加权，替换此处即可
			a.VelocityX += dvn * nx
			a.VelocityY += dvn * ny
			b.VelocityX -= dvn * nx
			b.VelocityY -= dvn * ny
		}

		// 位置分离：把重叠量各推一半，避免下一帧仍然重叠
		overlap := minDist - dist
		a.X -= nx * overlap * 0.5
		a.Y -= ny * overlap * 0.5
		b.X += nx * overlap * 0.5
		b.Y += ny * overlap * 0.5
	})

	// ── 2. 生物 vs 植物 ──────────────────────────────────────
	// 植物不动，只修正生物的速度和位置
	for _, c := range w.Creatures {
		if c == nil {
			continue
		}
		// 查询生物周围可能接触的植物（查询半径 = 生物半径 + 最大植物半径，保守估计）
		nearby := sh.QueryPlantsWithin(c.X, c.Y, c.Radius+DefaultPlantMaxRadius, 0)
		for _, p := range nearby {
			dx := p.X - c.X
			dy := p.Y - c.Y
			distSq := dx*dx + dy*dy
			minDist := c.Radius + p.Radius

			if distSq >= minDist*minDist || distSq == 0 {
				continue // 未重叠
			}
			dist := math.Sqrt(distSq)
			// 法线方向：从生物指向植物
			nx := dx / dist
			ny := dy / dist

			// 生物在法线方向的速度分量
			vn := c.VelocityX*nx + c.VelocityY*ny

			// 只在生物朝植物方向运动时反弹（弹性反弹：反转法线速度分量）
			if vn > 0 {
				c.VelocityX -= 2 * vn * nx
				c.VelocityY -= 2 * vn * ny
			}

			// 位置分离：把生物推出重叠区域（植物不动）
			overlap := minDist - dist
			c.X -= nx * overlap
			c.Y -= ny * overlap
		}
	}
}
