package core

// EntityType 实体类型枚举
type EntityType int

const (
	EntityTypeCreature EntityType = iota // 生物（会动的）
	EntityTypePlant                      // 植物（不会动）
	EntityTypeObstacle                   // 障碍物
)

// 封装颜色
type Color struct {
	R, G, B uint8
}

// Entity 游戏世界中的实体基础结构
type Entity struct {
	ID   uint64     // 唯一标识符
	Type EntityType // 实体类型

	// 位置（圆心坐标）
	X, Y float64

	// 大小（身体半径）
	Radius float64

	// 生命属性
	Health    int // 0死亡
	MaxHealth int

	// 质量（影响碰撞和运动）
	Mass    float64 // 质量
	MaxMass float64 // 最大质量（影响生长和分裂）
	MinMass float64 // 最小质量（影响生存和死亡）

	// 颜色（RGB，0-255）
	Color Color
}

// Creature 生物 - 继承 Entity 并扩展生物特有属性
type Creature struct {
	Entity // 嵌入基础实体

	// 运动属性
	VelocityX, VelocityY float64 // 速度向量
	Speed                float64 // 最大移动速度

	// 朝向（弧度，0 表示向右，逆时针为正）
	Direction      float64 // 身体朝向
	FocusDirection float64 // 关注度朝向

	// 眼睛属性（相对于身体）
	EyeRadius   float64 // 眼睛半径
	EyeOffset   float64 // 眼睛距离身体中心的偏移
	EyeAngle    float64 // 两眼之间的张角（弧度）
	EyeColor    Color   // 眼睛颜色
	PupilRadius float64 // 瞳孔半径
}

// Plant 植物 - 静态实体
type Plant struct {
	Entity // 嵌入基础实体

	// 植物特有属性（可后续扩展）
	GrowthStage int // 生长阶段
}

// NewCreature 创建一个新生物
func NewCreature(id uint64, x, y, radius float64, father *Creature) *Creature {
	return &Creature{
		Entity: Entity{
			ID:     id,
			Type:   EntityTypeCreature,
			X:      x,
			Y:      y,
			Radius: radius,
			// 默认绿色身体
			Color: Color{R: 0, G: 255, B: 0},
		},
		Speed:       100.0, // 默认速度
		Direction:   0,     // 默认朝右
		EyeRadius:   radius * 0.3,
		EyeOffset:   radius * 0.5,
		EyeAngle:    0.6,                           // 约 34 度
		EyeColor:    Color{R: 255, G: 255, B: 255}, // 白色眼睛
		PupilRadius: radius * 0.15,
	}
}

// NewPlant 创建一个新植物
func NewPlant(id uint64, x, y, radius float64) *Plant {
	return &Plant{
		Entity: Entity{
			ID:     id,
			Type:   EntityTypePlant,
			X:      x,
			Y:      y,
			Radius: radius,
			// 默认深绿色
			Color: Color{R: 0, G: 128, B: 0},
		},
		GrowthStage: 1,
	}
}
