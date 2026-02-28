package core

// World 表示一个二维世界，内部管理所有实体
type World struct {
	// 简单起步：先只用切片存放生物和植物
	Creatures []*Creature
	Plants    []*Plant

	nextID uint64 // 简单的自增 ID 生成器
}

// NewWorld 创建一个空的世界
func NewWorld() *World {
	return &World{}
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

// AddPlant 向世界中添加一个植物
func (w *World) AddPlant(p *Plant) {
	if p == nil {
		return
	}
	w.Plants = append(w.Plants, p)
}

// AllCreatures 返回世界中的所有生物（只读视角）
func (w *World) AllCreatures() []*Creature {
	return w.Creatures
}

// AllPlants 返回世界中的所有植物（只读视角）
func (w *World) AllPlants() []*Plant {
	return w.Plants
}
