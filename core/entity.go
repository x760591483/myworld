package core

import (
	"math"
	"math/rand/v2"
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
	Age       uint32 // 存活tick数
	MaxAge    uint32 // 最大存活tick数

	// 质量（影响碰撞和运动）
	Mass    float64 // 质量
	MaxMass float64 // 最大质量（影响生长和分裂）
	MinMass float64 // 最小质量（影响生存和死亡）

	// 颜色（RGB，0-255）
	Color Color

	// 随机值（可用于行为决策，范围 [0,1)）
	RandomValue float64
	// 计数器 tick太快
	TickCounter uint32
	// 死亡消失计数器 死亡后持续存在一段时间（如30秒）以供其他实体感知，之后从世界中移除
	DeathCounter uint32
}

// GetID 实现 SpatialEntity 接口 —— 返回实体唯一 ID
func (e *Entity) GetID() uint64 {
	return e.ID
}

// GetPosition 实现 SpatialEntity 接口 —— 返回实体位置
func (e *Entity) GetPosition() (float64, float64) {
	return e.X, e.Y
}

func (e *Entity) GetX() float64 {
	return e.X
}

func (e *Entity) GetY() float64 {
	return e.Y
}

func (e *Entity) GetRadius() float64 {
	return e.Radius
}

func (e *Entity) SetPosition(x, y float64) {
	e.X = x
	e.Y = y
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

	// 能量
	Energy    float64 // 当前能量
	MaxEnergy float64 // 最大能量
}

// Plant 植物 - 静态实体
type Plant struct {
	Entity // 嵌入基础实体
	Gene

	// 植物特有属性（可后续扩展）
	GrowthStage int // 生长阶段

	// 能量
	Energy    float64 // 当前能量,表示植物的营养价值，生物吃掉植物后获得的能量等于植物当前能量
	MaxEnergy float64 // 最大能量
	// 满状态时间
	FullTime uint32 // 满状态持续时间

}

// 生成随机值 （每次创建实体时调用，保证每个实体都有一个独特的随机值，范围 [0,1)）
func GenerateRandomValue() float64 {
	// 循环3次取最大的
	v1 := rand.Float64() // 生成一个随机值，范围 [0,1)
	v2 := rand.Float64()
	v3 := rand.Float64()
	return math.Max(v1, math.Max(v2, v3))
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

// NewCreature2 创建一个新生物，当父生物存在时，子生物继承父生物的颜色和速度的一部分 且位置选择附近随机点，但当周边没有空位时，生物创建失败，当父生物不存在时，随机生成颜色和速度，位置随机生成
func NewCreature2(id uint64, father *Creature) *Creature {
	var x, y float64
	var radius float64
	var color Color
	var eyeRadius float64
	var eyeOffset float64
	var pupilRadius float64
	var speed float64
	var velocityX, velocityY float64
	if father != nil {
		// 父生物存在，位置在父生物附近随机点 需判断重叠关系
		x = father.X
		y = father.Y                                    // 加入时重新设定位置，先放在父生物位置，后续通过空间索引寻找附近空位
		radius = father.Radius + (rand.Float64()*2-1)*2 // 父生物半径基础上随机变化 ±2
		if radius < DefaultCreatureMinRadius {
			radius = DefaultCreatureMinRadius
		} else if radius > DefaultCreatureMaxRadius {
			radius = DefaultCreatureMaxRadius
		}
		color = father.Color
		color.R = uint8(float64(color.R) + (rand.Float64()*2-1)*10)
		color.G = uint8(float64(color.G) + (rand.Float64()*2-1)*10)
		color.B = uint8(float64(color.B) + (rand.Float64()*2-1)*10)
		speed = father.Speed + (rand.Float64()*2-1)*5 // 父生物速度基础上随机变化 ±5

	} else {
		// 父生物不存在，位置随机生成 也需要判断重叠关系
		x = -1.0
		y = -1.0                                                // 加入时随机生成位置，先设为负数，后续通过空间索引寻找空位
		radius = DefaultCreatureRadius + (rand.Float64()*2-1)*2 // 默认半径基础上随机变化 ±2
		if radius < DefaultCreatureMinRadius {
			radius = DefaultCreatureMinRadius
		} else if radius > DefaultCreatureMaxRadius {
			radius = DefaultCreatureMaxRadius
		}
		color = Color{
			R: uint8(rand.Float64() * 256),
			G: uint8(rand.Float64() * 256),
			B: uint8(rand.Float64() * 256),
		}
		speed = DefaultCreatureSpeed + (rand.Float64()*2-1)*5 // 默认速度基础上随机变化 ±5
	}
	eyeRadius = radius * 0.3
	eyeOffset = radius * 0.5
	pupilRadius = radius * 0.15
	// 随机生成初始速度方向和大小
	angle := rand.Float64() * 2 * 3.1415926535
	velocityX = speed * math.Cos(angle)
	velocityY = speed * math.Sin(angle)

	return &Creature{
		Entity: Entity{
			ID:     id,
			Type:   EntityTypeCreature,
			X:      x,
			Y:      y,
			Radius: radius,
			Color:  color,
		},
		Speed:       speed,
		Direction:   0, // 默认朝右
		EyeRadius:   eyeRadius,
		EyeOffset:   eyeOffset,
		EyeAngle:    0.6,                           // 约 34 度
		EyeColor:    Color{R: 255, G: 255, B: 255}, // 白色眼睛
		PupilRadius: pupilRadius,
		VelocityX:   velocityX,
		VelocityY:   velocityY,
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

// NewPlant2 创建一个新植物，位置随机生成，半径随机生成，颜色随机生成
func NewPlant2(id uint64, father *Plant) *Plant {
	var x, y float64
	var radius float64
	var color Color
	var healthMax int
	var health int

	var maxMass float64
	var mass float64

	var energyMax float64
	var energy float64

	// 生成随机值
	randomValue := GenerateRandomValue()

	if father != nil {
		// 父植物存在，位置在父植物附近随机点
		x = father.X
		y = father.Y // 加入时重新设定位置，先放在父植物位置，后续通过空间索引寻找附近空位

		radius = father.Radius + (rand.Float64()*2-1)*2 // 父植物半径基础上随机变化 ±2
		if radius < DefaultPlantMinRadius {
			radius = DefaultPlantMinRadius
		} else if radius > DefaultPlantMaxRadius {
			radius = DefaultPlantMaxRadius
		}
		color = father.Color
		color.R = uint8(float64(color.R) + (rand.Float64()*2-1)*10)
		color.G = uint8(float64(color.G) + (rand.Float64()*2-1)*10)
		color.B = uint8(float64(color.B) + (rand.Float64()*2-1)*10)

		healthMax = father.MaxHealth + rand.IntN(11) - 4     // 父植物生命值基础上随机变化
		maxMass = father.MaxMass + (rand.Float64()*2-0.8)*20 // 父植物质量基础上随机变化
		energyMax = father.MaxEnergy + (rand.Float64()*2-0.8)*20

		{
			// 父类将生命1/3传给子类 总量保持不变
			health = int(float64(father.Health) * 0.33)
			father.Health -= health

			// 父类将质量1/4传给子类 总量保持不变
			mass = father.Mass * 0.25
			father.Mass -= mass

			// 父类将因为后代先能效消耗1/3 然后再将剩余能量的1/4传给子类 总量保持不变
			father.Energy -= father.Energy * 0.33 // 代价
			energy = father.Energy * 0.25
			father.Energy -= energy
		}

	} else {
		// 父植物不存在，位置随机生成
		x = -1.0
		y = -1.0                                             // 加入时随机生成位置，先设为负数，后续通过空间索引寻找空位
		radius = DefaultPlantRadius + (rand.Float64()*2-1)*2 // 默认半径基础上随机变化 ±2
		if radius < DefaultPlantMinRadius {
			radius = DefaultPlantMinRadius
		} else if radius > DefaultPlantMaxRadius {
			radius = DefaultPlantMaxRadius
		}
		color = Color{
			R: uint8(rand.Float64() * 256),
			G: uint8(rand.Float64() * 256),
			B: uint8(rand.Float64() * 256),
		}

		health = int(DefaultPlantHealth * (0.5 + rand.Float64()*0.5))               // 默认生命值基础上随机变化 50% - 100%
		healthMax = health + int(float64(DefaultPlantMaxHealth-health)*randomValue) // 确保最大生命值不小于当前生命值

		mass = DefaultPlantMass * (0.5 + rand.Float64()*0.5)    // 默认质量基础上随机变化 50% - 100%
		maxMass = mass + (DefaultPlantMaxMass-mass)*randomValue // 确保最大质量不小于当前质量

		energy = DefaultPlantEnergy * (0.5 + rand.Float64()*0.5)        // 默认能量基础上随机变化 50% - 100%
		energyMax = energy + (DefaultPlantMaxEnergy-energy)*randomValue // 确保最大能量不小于当前能量
	}

	health2 := int(DefaultPlantHealth * randomValue)
	if health2 > healthMax {
		healthMax = health2
	}

	return &Plant{
		Entity: Entity{
			ID:           id,
			Type:         EntityTypePlant,
			X:            x,
			Y:            y,
			Radius:       radius,
			Color:        color,
			RandomValue:  randomValue,
			MaxHealth:    healthMax,
			Health:       health,
			MaxMass:      maxMass,
			Mass:         mass,
			MinMass:      DefaultPlantMinMass,
			Age:          0,
			MaxAge:       uint32(float64(DefaultPlantMaxAge) * randomValue),
			DeathCounter: DefaultDeathDuration,
		},
		GrowthStage: 1,
		Energy:      energy,
		MaxEnergy:   energyMax,
		FullTime:    0,
	}

}
