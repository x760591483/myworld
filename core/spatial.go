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
// SpatialHash —— 通用空间哈希（支持任意 SpatialEntity）
// ──────────────────────────────────────────────────────────────

// SpatialHash 空间哈希索引，用于加速二维空间中的邻居查询
// 避免每个实体与所有其他实体进行 O(n²) 的位置比较
type SpatialHash struct {
	cellSize    float64           // 格子大小，应 >= 最大交互半径
	bucketCount int               // 桶数量（固定大小数组）
	buckets     [][]SpatialEntity // 每个桶存放落在对应格子中的实体
}

// NewSpatialHash 创建空间哈希
// cellSize: 格子边长，建议设为最大感知/交互半径
// bucketCount: 桶数量，建议为实体数量的 2-4 倍（质数更佳）
func NewSpatialHash(cellSize float64, bucketCount int) *SpatialHash {
	buckets := make([][]SpatialEntity, bucketCount)
	for i := range buckets {
		buckets[i] = make([]SpatialEntity, 0, 8) // 预分配容量，减少 append 扩容
	}
	return &SpatialHash{
		cellSize:    cellSize,
		bucketCount: bucketCount,
		buckets:     buckets,
	}
}

// hash 将格子坐标映射到桶索引
func (sh *SpatialHash) hash(cx, cy int) int {
	// 使用大质数异或，分布更均匀
	h := cx*73856093 ^ cy*19349663
	if h < 0 {
		h = -h
	}
	return h % sh.bucketCount
}

// cellCoord 将世界坐标转为格子坐标
func (sh *SpatialHash) cellCoord(x, y float64) (int, int) {
	cx := int(math.Floor(x / sh.cellSize))
	cy := int(math.Floor(y / sh.cellSize))
	return cx, cy
}

// Clear 清空所有桶（复用底层内存，不重新分配）
func (sh *SpatialHash) Clear() {
	for i := range sh.buckets {
		sh.buckets[i] = sh.buckets[i][:0] // 长度归零，容量保留
	}
}

// Insert 将一个实体插入空间哈希
func (sh *SpatialHash) Insert(e SpatialEntity) {
	x, y := e.GetPosition()
	cx, cy := sh.cellCoord(x, y)
	idx := sh.hash(cx, cy)
	sh.buckets[idx] = append(sh.buckets[idx], e)
}

// Remove 从空间哈希中移除一个实体（按 ID 匹配）
// 用于静态实体的删除（如植物被吃掉、枯萎）
func (sh *SpatialHash) Remove(e SpatialEntity) {
	x, y := e.GetPosition()
	cx, cy := sh.cellCoord(x, y)
	idx := sh.hash(cx, cy)

	bucket := sh.buckets[idx]
	targetID := e.GetID()
	for i, item := range bucket {
		if item.GetID() == targetID {
			// 用最后一个元素覆盖当前位置，然后截断（O(1) 删除，不保序）
			bucket[i] = bucket[len(bucket)-1]
			bucket[len(bucket)-1] = nil // 避免内存泄漏
			sh.buckets[idx] = bucket[:len(bucket)-1]
			return
		}
	}
}

// Build 重建整个空间哈希索引（先清空，再全部插入）
func (sh *SpatialHash) Build(entities []SpatialEntity) {
	sh.Clear()
	for _, e := range entities {
		if e != nil {
			sh.Insert(e)
		}
	}
}

// QueryNeighbors 查询某个位置附近的所有实体候选者
// 遍历自身所在格子 + 周围 8 个格子（共 9 格），返回候选列表
// 注意：返回的候选者中可能包含自身，也可能包含距离较远的实体
// 调用方需要自行做精确的距离判断
func (sh *SpatialHash) QueryNeighbors(x, y float64) []SpatialEntity {
	cx, cy := sh.cellCoord(x, y)

	var result []SpatialEntity
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			idx := sh.hash(cx+dx, cy+dy)
			result = append(result, sh.buckets[idx]...)
		}
	}
	return result
}

// QueryNeighborsWithin 查询距离 (x, y) 在 radius 范围内的所有实体
// 在 QueryNeighbors 的基础上做精确距离过滤（使用距离平方，避免开方）
// excludeID: 要排除的实体 ID（通常是自身），传 0 表示不排除
func (sh *SpatialHash) QueryNeighborsWithin(x, y, radius float64, excludeID uint64) []SpatialEntity {
	candidates := sh.QueryNeighbors(x, y)
	radiusSq := radius * radius

	var result []SpatialEntity
	for _, e := range candidates {
		if e.GetID() == excludeID {
			continue
		}
		ex, ey := e.GetPosition()
		dx := ex - x
		dy := ey - y
		distSq := dx*dx + dy*dy
		if distSq <= radiusSq {
			result = append(result, e)
		}
	}
	return result
}

// ForEachPair 遍历所有可能接触的实体对，回调函数处理每对
// 通过 ID 大小关系保证每对只处理一次（避免重复）
func (sh *SpatialHash) ForEachPair(entities []SpatialEntity, callback func(a, b SpatialEntity)) {
	for _, a := range entities {
		if a == nil {
			continue
		}
		ax, ay := a.GetPosition()
		candidates := sh.QueryNeighbors(ax, ay)
		for _, b := range candidates {
			if b == nil {
				continue
			}
			// 只处理 a.ID < b.ID 的对，避免重复
			if a.GetID() >= b.GetID() {
				continue
			}
			callback(a, b)
		}
	}
}

// ──────────────────────────────────────────────────────────────
// WorldSpatialIndex —— 世界级空间索引（双层：动态 + 静态）
// ──────────────────────────────────────────────────────────────

// WorldSpatialIndex 管理世界中所有实体的空间索引
// 分为两层：
//   - dynamicHash: 动态层，存放每帧都在移动的实体（Creature），每帧重建
//   - staticHash:  静态层，存放不会移动的实体（Plant），仅在增删时更新
//
// 这样植物不需要每帧重新插入，大幅减少静态实体的开销
type WorldSpatialIndex struct {
	dynamicHash *SpatialHash // 动态实体（生物），每帧重建
	staticHash  *SpatialHash // 静态实体（植物），增删时更新
}

// NewWorldSpatialIndex 创建世界级空间索引
// cellSize: 格子边长
// dynamicBucketCount: 动态层桶数量（建议为生物数量的 2-4 倍）
// staticBucketCount:  静态层桶数量（建议为植物数量的 2-4 倍）
func NewWorldSpatialIndex(cellSize float64, dynamicBucketCount, staticBucketCount int) *WorldSpatialIndex {
	return &WorldSpatialIndex{
		dynamicHash: NewSpatialHash(cellSize, dynamicBucketCount),
		staticHash:  NewSpatialHash(cellSize, staticBucketCount),
	}
}

// RebuildDynamic 每帧调用：重建动态层（生物位置每帧都变）
func (wi *WorldSpatialIndex) RebuildDynamic(creatures []*Creature) {
	wi.dynamicHash.Clear()
	for _, c := range creatures {
		if c != nil {
			wi.dynamicHash.Insert(c)
		}
	}
}

// InsertStatic 插入一个静态实体（植物种下时调用一次即可）
func (wi *WorldSpatialIndex) InsertStatic(e SpatialEntity) {
	wi.staticHash.Insert(e)
}

// RemoveStatic 移除一个静态实体（植物被吃掉/死亡时调用）
func (wi *WorldSpatialIndex) RemoveStatic(e SpatialEntity) {
	wi.staticHash.Remove(e)
}

// RebuildStatic 完全重建静态层（批量操作后使用，一般不需要每帧调用）
func (wi *WorldSpatialIndex) RebuildStatic(plants []*Plant) {
	wi.staticHash.Clear()
	for _, p := range plants {
		if p != nil {
			wi.staticHash.Insert(p)
		}
	}
}

// QueryAllNeighbors 查询某个位置附近的所有实体（动态 + 静态）
func (wi *WorldSpatialIndex) QueryAllNeighbors(x, y float64) []SpatialEntity {
	dynamic := wi.dynamicHash.QueryNeighbors(x, y)
	static := wi.staticHash.QueryNeighbors(x, y)
	return append(dynamic, static...)
}

// QueryAllNeighborsWithin 查询距离 (x, y) 在 radius 范围内的所有实体（动态 + 静态）
func (wi *WorldSpatialIndex) QueryAllNeighborsWithin(x, y, radius float64, excludeID uint64) []SpatialEntity {
	dynamic := wi.dynamicHash.QueryNeighborsWithin(x, y, radius, excludeID)
	static := wi.staticHash.QueryNeighborsWithin(x, y, radius, excludeID)
	return append(dynamic, static...)
}

// QueryCreaturesWithin 只查询附近的生物（不含植物）
func (wi *WorldSpatialIndex) QueryCreaturesWithin(x, y, radius float64, excludeID uint64) []*Creature {
	candidates := wi.dynamicHash.QueryNeighborsWithin(x, y, radius, excludeID)
	result := make([]*Creature, 0, len(candidates))
	for _, e := range candidates {
		if c, ok := e.(*Creature); ok {
			result = append(result, c)
		}
	}
	return result
}

// QueryPlantsWithin 只查询附近的植物（不含生物）
func (wi *WorldSpatialIndex) QueryPlantsWithin(x, y, radius float64, excludeID uint64) []*Plant {
	candidates := wi.staticHash.QueryNeighborsWithin(x, y, radius, excludeID)
	result := make([]*Plant, 0, len(candidates))
	for _, e := range candidates {
		if p, ok := e.(*Plant); ok {
			result = append(result, p)
		}
	}
	return result
}
