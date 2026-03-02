package core

import "math"

// ──────────────────────────────────────────────────────────────
// SpatialEntity 接口 —— 所有可被空间索引的实体都需实现
// ──────────────────────────────────────────────────────────────

// SpatialEntity 空间实体接口，任何需要参与空间查询的对象都应实现此接口
type SpatialEntity interface {
	GetID() uint64
	GetPosition() (x, y float64)
}

// 确保 Creature 和 Plant 都实现了 SpatialEntity 接口（编译期检查）
var _ SpatialEntity = (*Creature)(nil)
var _ SpatialEntity = (*Plant)(nil)

// ──────────────────────────────────────────────────────────────
// SpatialHash —— 单层空间哈希，内含两个独立桶数组
//   creatureBuckets: 动态实体（生物），每帧重建
//   plantBuckets:    静态实体（植物），增删时更新
// ──────────────────────────────────────────────────────────────

// SpatialHash 空间哈希索引
// 用两个二维桶数组分别管理生物与植物，共享同一组格子参数
type SpatialHash struct {
	cellSize        float64       // 格子边长，应 >= 最大交互半径
	bucketCount     int           // 桶数量（固定大小）
	creatureBuckets [][]*Creature // 动态实体（生物）桶
	plantBuckets    [][]*Plant    // 静态实体（植物）桶
}

// NewSpatialHash 创建空间哈希
// cellSize:    格子边长，建议设为最大感知/交互半径
// bucketCount: 桶数量，建议为实体总数的 2-4 倍（质数更佳）
func NewSpatialHash(cellSize float64, bucketCount int) *SpatialHash {
	cb := make([][]*Creature, bucketCount)
	pb := make([][]*Plant, bucketCount)
	for i := range cb {
		cb[i] = make([]*Creature, 0, 8)
		pb[i] = make([]*Plant, 0, 8)
	}
	return &SpatialHash{
		cellSize:        cellSize,
		bucketCount:     bucketCount,
		creatureBuckets: cb,
		plantBuckets:    pb,
	}
}

// hash 将格子坐标映射到桶索引（大质数异或，分布均匀）
func (sh *SpatialHash) hash(cx, cy int) int {
	h := cx*73856093 ^ cy*19349663
	if h < 0 {
		h = -h
	}
	return h % sh.bucketCount
}

// cellCoord 将世界坐标转为格子坐标
func (sh *SpatialHash) cellCoord(x, y float64) (int, int) {
	return int(math.Floor(x / sh.cellSize)), int(math.Floor(y / sh.cellSize))
}

// ── 清空 ────────────────────────────────────────────────────

// ClearCreatures 清空所有生物桶（复用底层内存）
func (sh *SpatialHash) ClearCreatures() {
	for i := range sh.creatureBuckets {
		sh.creatureBuckets[i] = sh.creatureBuckets[i][:0]
	}
}

// ClearPlants 清空所有植物桶（复用底层内存）
func (sh *SpatialHash) ClearPlants() {
	for i := range sh.plantBuckets {
		sh.plantBuckets[i] = sh.plantBuckets[i][:0]
	}
}

// Clear 同时清空生物桶和植物桶
func (sh *SpatialHash) Clear() {
	sh.ClearCreatures()
	sh.ClearPlants()
}

// ── 插入 / 移除 ─────────────────────────────────────────────

// InsertCreature 将一个生物插入生物桶
func (sh *SpatialHash) InsertCreature(c *Creature) {
	cx, cy := sh.cellCoord(c.X, c.Y)
	idx := sh.hash(cx, cy)
	sh.creatureBuckets[idx] = append(sh.creatureBuckets[idx], c)
}

// InsertPlant 将一个植物插入植物桶
func (sh *SpatialHash) InsertPlant(p *Plant) {
	cx, cy := sh.cellCoord(p.X, p.Y)
	idx := sh.hash(cx, cy)
	sh.plantBuckets[idx] = append(sh.plantBuckets[idx], p)
}

// RemovePlant 从植物桶中移除一个植物（按 ID 匹配，O(1) 不保序）
// 植物被吃掉或枯萎时调用
func (sh *SpatialHash) RemovePlant(p *Plant) {
	cx, cy := sh.cellCoord(p.X, p.Y)
	idx := sh.hash(cx, cy)

	bucket := sh.plantBuckets[idx]
	for i, item := range bucket {
		if item.GetID() == p.GetID() {
			bucket[i] = bucket[len(bucket)-1]
			bucket[len(bucket)-1] = nil // 避免内存泄漏
			sh.plantBuckets[idx] = bucket[:len(bucket)-1]
			return
		}
	}
}

// ── 批量重建 ─────────────────────────────────────────────────

// RebuildCreatures 每帧调用：清空并重新插入所有生物
func (sh *SpatialHash) RebuildCreatures(creatures []*Creature) {
	sh.ClearCreatures()
	for _, c := range creatures {
		if c != nil {
			sh.InsertCreature(c)
		}
	}
}

// RebuildPlants 批量重建植物索引（初始化或大批量操作后使用）
func (sh *SpatialHash) RebuildPlants(plants []*Plant) {
	sh.ClearPlants()
	for _, p := range plants {
		if p != nil {
			sh.InsertPlant(p)
		}
	}
}

// ── 邻居查询 ─────────────────────────────────────────────────

// neighborsRaw 收集 (cx,cy) 周围 9 格中的生物和植物候选者（未做距离过滤）
func (sh *SpatialHash) neighborsRaw(cx, cy int) ([]*Creature, []*Plant) {
	var cs []*Creature
	var ps []*Plant
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			idx := sh.hash(cx+dx, cy+dy)
			cs = append(cs, sh.creatureBuckets[idx]...)
			ps = append(ps, sh.plantBuckets[idx]...)
		}
	}
	return cs, ps
}

// QueryNeighbors 返回 (x, y) 附近 9 格内的所有生物和植物候选者
// 候选者中可能包含距离较远的实体，调用方需自行精确过滤
func (sh *SpatialHash) QueryNeighbors(x, y float64) ([]*Creature, []*Plant) {
	cx, cy := sh.cellCoord(x, y)
	return sh.neighborsRaw(cx, cy)
}

// QueryCreaturesWithin 查询距离 (x, y) 在 radius 内的生物
// excludeID: 要排除的实体 ID，传 0 表示不排除
func (sh *SpatialHash) QueryCreaturesWithin(x, y, radius float64, excludeID uint64) []*Creature {
	cx, cy := sh.cellCoord(x, y)
	candidates, _ := sh.neighborsRaw(cx, cy)
	radiusSq := radius * radius

	result := make([]*Creature, 0, len(candidates))
	for _, c := range candidates {
		if c.GetID() == excludeID {
			continue
		}
		dx := c.X - x
		dy := c.Y - y
		if dx*dx+dy*dy <= radiusSq {
			result = append(result, c)
		}
	}
	return result
}

// QueryPlantsWithin 查询距离 (x, y) 在 radius 内的植物
// excludeID: 要排除的实体 ID，传 0 表示不排除
func (sh *SpatialHash) QueryPlantsWithin(x, y, radius float64, excludeID uint64) []*Plant {
	cx, cy := sh.cellCoord(x, y)
	_, candidates := sh.neighborsRaw(cx, cy)
	radiusSq := radius * radius

	result := make([]*Plant, 0, len(candidates))
	for _, p := range candidates {
		if p.GetID() == excludeID {
			continue
		}
		dx := p.X - x
		dy := p.Y - y
		if dx*dx+dy*dy <= radiusSq {
			result = append(result, p)
		}
	}
	return result
}

// QueryAllWithin 查询距离 (x, y) 在 radius 内的所有实体（生物 + 植物）
// excludeID: 要排除的实体 ID，传 0 表示不排除
func (sh *SpatialHash) QueryAllWithin(x, y, radius float64, excludeID uint64) []SpatialEntity {
	cx, cy := sh.cellCoord(x, y)
	cs, ps := sh.neighborsRaw(cx, cy)
	radiusSq := radius * radius

	result := make([]SpatialEntity, 0, len(cs)+len(ps))
	for _, c := range cs {
		if c.GetID() == excludeID {
			continue
		}
		dx := c.X - x
		dy := c.Y - y
		if dx*dx+dy*dy <= radiusSq {
			result = append(result, c)
		}
	}
	for _, p := range ps {
		if p.GetID() == excludeID {
			continue
		}
		dx := p.X - x
		dy := p.Y - y
		if dx*dx+dy*dy <= radiusSq {
			result = append(result, p)
		}
	}
	return result
}

// ForEachCreaturePair 遍历所有可能接触的生物对，每对只回调一次（a.ID < b.ID）
func (sh *SpatialHash) ForEachCreaturePair(creatures []*Creature, callback func(a, b *Creature)) {
	for _, a := range creatures {
		if a == nil {
			continue
		}
		cx, cy := sh.cellCoord(a.X, a.Y)
		candidates, _ := sh.neighborsRaw(cx, cy)
		for _, b := range candidates {
			if b == nil || a.GetID() >= b.GetID() {
				continue
			}
			callback(a, b)
		}
	}
}
