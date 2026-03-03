package core

// World 表示一个二维世界，内部管理所有实体
type World struct {
	// 简单起步：先只用切片存放生物和植物
	Creatures []*Creature
	Plants    []*Plant

	nextID uint64 // 简单的自增 ID 生成器

	// 空间索引
	SpatialIndex *SpatialHash

	// 世界边界
	Width  float64
	Height float64
}

// NewWorld 创建一个空的世界
// cellSize: 空间哈希格子大小，建议设为生物最大交互半径
func NewWorld() *World {
	return &World{
		// 动态层桶数 = MaxEntitys（生物），静态层桶数 = MaxEntitys（植物）
		SpatialIndex: NewSpatialHash(50.0, MaxEntitys),
		Width:        WorldWidth,
		Height:       WorldHeight,
	}
}

// 返回世界宽度
func (w *World) GetWidth() float64 {
	return w.Width
}

// 返回世界高度
func (w *World) GetHeight() float64 {
	return w.Height
}

// NextID 返回一个新的全局唯一 ID（在当前 world 内）
func (w *World) NextID() uint64 {
	w.nextID++
	return w.nextID
}

// AddCreature 向世界中添加一个生物
func (w *World) AddCreature(c *Creature) {
	if c == nil {
		return
	}
	w.Creatures = append(w.Creatures, c)
}

// AddPlant 向世界中添加一个植物（同时插入静态空间索引，无需每帧重建）
func (w *World) AddPlant(p *Plant) {
	if p == nil {
		return
	}
	w.Plants = append(w.Plants, p)
	w.SpatialIndex.InsertPlant(p) // 植物不动，插入一次即可
}

// RemovePlant 移除一个植物（同时从静态空间索引中删除）
func (w *World) RemovePlant(p *Plant) {
	if p == nil {
		return
	}
	// 从切片中移除
	for i, plant := range w.Plants {
		if plant.ID == p.ID {
			w.Plants[i] = w.Plants[len(w.Plants)-1]
			w.Plants[len(w.Plants)-1] = nil
			w.Plants = w.Plants[:len(w.Plants)-1]
			break
		}
	}
	// 从静态空间索引中移除
	w.SpatialIndex.RemovePlant(p)
}

// AllCreatures 返回世界中的所有生物（只读视角）
func (w *World) AllCreatures() []*Creature {
	return w.Creatures
}

// AllPlants 返回世界中的所有植物（只读视角）
func (w *World) AllPlants() []*Plant {
	return w.Plants
}
