package core

// Tick 推进世界时间 dt 秒
// 这一版只做一件简单的事：
// 根据生物的速度 (VelocityX, VelocityY) 更新它们的位置 (X, Y)
// 不考虑碰撞、不考虑边界、不考虑 AI
func (w *World) Tick(dt float64) {
	if w == nil {
		return
	}
	if dt <= 0 {
		return
	}

	for _, c := range w.Creatures {
		if c == nil {
			continue
		}

		// x = x + v * t
		c.X += c.VelocityX * dt
		c.Y += c.VelocityY * dt
	}
}
