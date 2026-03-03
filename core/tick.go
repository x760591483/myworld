package core

// Tick 推进世界一帧，分三个阶段：
//
//  1. Think  — 每个生物根据周边感知决策，写入 VelocityX/Y（rules.go）
//  2. Move   — 用速度积分，算出期望位置
//  3. Resolve — 修正非法状态：碰撞分离 + 边界反弹（physics.go）
//
// 阶段之间严格隔离：Think 只写速度，Move 只写位置，Resolve 只做修正。
func (w *World) Tick(dt float64) {
	if w == nil {
		return
	}
	if dt <= 0 {
		return
	}

	// ── 阶段 0：重建动态空间索引（基于上一帧末尾的位置）─────
	w.SpatialIndex.RebuildCreatures(w.Creatures)

	// ── 阶段 1：Think —— 决策，更新每个生物的速度意图 ────────
	for _, c := range w.Creatures {
		if c == nil {
			continue
		}
		updateIntention(c, w, dt)
	}

	// ── 阶段 2：Move —— 积分，算出期望位置 ───────────────────
	for _, c := range w.Creatures {
		if c == nil {
			continue
		}
		c.X += c.VelocityX * dt
		c.Y += c.VelocityY * dt
	}

	// ── 阶段 3：Resolve —— 物理修正，消除重叠和越界 ──────────
	resolveCollisions(w)
	for _, c := range w.Creatures {
		if c == nil {
			continue
		}
		resolveWorldBounds(c, w)
		updateEyeDirection(c)
	}

	//

	// ── 维护种群数量（生物/植物不足时自动补充）───────────────
	w.maintainPopulation()
}
