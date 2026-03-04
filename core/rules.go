package core

import (
	"math"
	"math/rand/v2"
)

// EntityType 实体类型枚举
type EntityType int

const (
	EntityTypeCreature EntityType = iota // 生物（会动的）
	EntityTypePlant                      // 植物（不会动）
	EntityTypeObstacle                   // 障碍物

)

// 规则和一些设定参数
// 预测世界生物数量数
const (
	// 每秒tick数
	TicksPerSecond = 60

	MaxEntitys = 10000

	// 世界长宽
	WorldWidth  = 1000
	WorldHeight = 1000

	// 世界最小生物数
	MinCreatures = 20 // 当小于该数时 将定时生成新生物
	MinPlants    = 30 // 当小于该数时 将定时生成新植物

	DefaultCreatureRadius     = 10.0   // 默认生物半径
	DefaultCreatureMaxRadius  = 15.0   // 默认生物最大半径
	DefaultCreatureMinRadius  = 8.0    // 默认生物最小半径
	DefaultCreatureMaxHealth  = 100    // 生物最大生命值
	DefaultHealth             = 80     // 生物初始生命值
	DefaultCreatureMinHealth  = 0      // 0 死亡
	DefaultCreatureMaxAge     = 216000 // 生物最大存活tick数 (假设每秒60tick，216000tick约等于1小时)
	DefaultCreatureNewTick    = 3600   // 生物满状态超过该tick数时，条件之一
	DefaultCreatureMaxSpeed   = 50.0   // 生物最大移动速度
	DefaultCreatureSpeed      = 30.0   // 生物初始移动速度
	DefaultCreatureNewspacing = 2.0    // 生物新生时和父类之间的间距
	DefaultCreatureTurnSpeed  = 0.12   // 生物眼睛转向的平滑程度
	DefaultCreatureMaxMass    = 100.0
	DefaultCreatureMinMass    = 10.0
	DefaultCreatureMass       = 50.0
	DefaultCreatureMaxEnergy  = 150.0
	DefaultCreatureEachEnergy = 2.0 // 每tick能吸收能量上限
	DefaultCreatureMinEnergy  = 0.0
	DefaultCreatureEnergy     = 100.0

	DefaultPlantRadius     = 15.0    // 默认植物半径
	DefaultPlantMaxRadius  = 20.0    // 默认植物最大半径
	DefaultPlantMinRadius  = 12.0    // 默认植物最小半径
	DefaultPlantMaxHealth  = 200     // 植物最大生命值
	DefaultPlantHealth     = 100     // 植物初始生命值
	DefaultPlantMinHealth  = 0       // 0 死亡
	DefaultPlantMaxAge     = 2160000 // 植物最大存活tick数 (假设每秒60tick，2160000tick约等于10小时)
	DefaultPlantNewTick    = 3600    // 植物满状态超过该tick数时，条件之一
	DefaultPlantNewspacing = 10.0    // 植物新生时和父类之间的间距
	DefaultPlantMaxMass    = 500.0
	DefaultPlantMinMass    = 50.0
	DefaultPlantMass       = 200.0
	DefaultPlantMaxEnergy  = 400.0 // 植物能量值上限，超过后不再生长
	DefaultPlantEachEnergy = 0.5   // 每tick增加的能量值
	DefaultPlantMinEnergy  = 0.0
	DefaultPlantEnergy     = 100.0

	// 两个物体之间相互索取距离, 即捕食者能能吃猎物  动物能吃植物的边界距离
	InteractionDistance = 8.0 // 即小于等于该距离时，生物可以吃掉植物或其他生物
)

// ── 种群维护 ────────────────────────────────────────────────

// maintainPopulation 检查当前世界生物/植物数量，
// 当数量低于最低阈值时自动补充，每帧由 Tick 调用。
func (w *World) maintainPopulation() {
	// ── 补充生物 ──────────────────────────────────────────
	for len(w.Creatures) < MinCreatures {
		if !w.spawnCreature(nil) {
			break // 找不到空位就放弃，等下一帧再试
		}
	}

	// ── 补充植物 ──────────────────────────────────────────
	for len(w.Plants) < MinPlants {
		if !w.spawnPlant(nil) {
			break
		}
	}
}

// spawnCreature 在世界中生成一个新生物。
// father != nil 时在父生物附近找空位；father == nil 时在世界范围内随机找空位。
// 返回是否成功生成。
func (w *World) spawnCreature(father *Creature) bool {
	c := NewCreature2(w.NextID(), father)
	if father != nil {
		// 在父生物附近寻找空位
		nx, ny, ok := w.SpatialIndex.GetFreePositionAround(
			father.GetX(), father.GetY(),
			father.GetRadius(), c.GetRadius(),
			DefaultCreatureNewspacing,
		)
		if !ok {
			return false
		}
		c.SetPosition(nx, ny)
	} else {
		// 一个tick只执行一次 本次不行只能下次尝试
		newX, newY, ok := w.SpatialIndex.GetFreePositionAround(c.GetX(), c.GetY(), 0, c.GetRadius(), DefaultCreatureNewspacing)
		if !ok {
			return false
		}
		c.SetPosition(newX, newY)
	}

	w.AddCreature(c)
	return true // ← 修正：成功添加后返回 true
}

// spawnPlant 在世界中生成一个新植物。
// father != nil 时在父植物附近找空位；father == nil 时在世界范围内随机找空位。
// 返回是否成功生成。
func (w *World) spawnPlant(father *Plant) bool {
	p := NewPlant2(w.NextID(), father)

	if father != nil {
		// 在父植物附近寻找空位
		nx, ny, ok := w.SpatialIndex.GetFreePositionAround(
			father.X, father.Y,
			father.Radius, p.Radius,
			DefaultPlantNewspacing,
		)
		if !ok {
			return false
		}
		p.SetPosition(nx, ny)
	} else {
		newX, newY, ok := w.SpatialIndex.GetFreePositionAround(p.GetX(), p.GetY(), 0, p.GetRadius(), DefaultPlantNewspacing)
		if !ok {
			return false
		}
		p.SetPosition(newX, newY)

	}

	w.AddPlant(p)
	return true
}

// updateIntention 决策阶段：根据周边感知更新生物的速度意图。
//
// 只允许修改 c.VelocityX / c.VelocityY / c.Direction，
// 禁止在此函数内直接修改 c.X / c.Y（位置由 Move 阶段统一积分）。
// 禁止在此函数内处理碰撞/边界（由 Resolve 阶段处理）。
// 禁止在此函数内处理吃掉/受伤等规则（在独立的规则函数中处理）。
func updateCreatureIntention(c *Creature, w *World, dt float64) {
	// TODO: 查询周边信息 → 决策 → 写入速度
	// 示例：暂时维持当前速度不变，等 AI 逻辑填入
	_ = w
	_ = dt
}
func updatePlantIntention(p *Plant, w *World, dt float64) {
	// 植被也有生长阶段
	// ----- 规则 ------
	// 1. 每个tick增加年龄，直到达到最大年龄时死亡
	// 2. 每个秒增加能量值，直到达到最大能量值
	// 3. 当能量低于10%时，生命值每tick减少1点；当能量高于90%时，生命值每tick增加1点
	// 4. 当生命值为0时死亡
	// 5. 当年龄达到最大年龄时死亡
	// 6. 当生命和能量都满时慢状态超过1分钟时，且年龄在最大年龄20%~50%之间时，有50%概率繁殖新植物

	p.Age++
	if p.Age >= p.MaxAge {
		p.Health = 0
		return
	}
	if p.Health <= 0 {
		return
	}

	if p.Energy < p.MaxEnergy {
		// 当能量值低于最大能量值时，每tick增加能量值，直到达到最大能量值
		p.Energy += DefaultPlantEachEnergy * p.RandomValue
	}

	p.TickCounter++
	if p.TickCounter == TicksPerSecond {
		p.TickCounter = 0
		// 每秒处理

		if p.Health < p.MaxHealth && p.Energy > 0.9*p.MaxEnergy {
			// 当能量值高于最大能量值的90%时，每tick增加生命值，直到达到最大生命值
			p.Health += 1
		}

		if p.Energy < 0.1*p.MaxEnergy {
			// 当能量值低于最大能量值的10%时，每tick减少生命值，直到达到0
			p.Health -= 1
		}

	}

	if p.Health >= p.MaxHealth && p.Energy >= p.MaxEnergy {
		// 当生命值和能量值都达到最大值时，且年龄在最大年龄的20%~50%之间时，慢状态超过1分钟时，有50%概率繁殖新植物
		ageFloat := float64(p.Age)
		maxAgeFloat := float64(p.MaxAge)
		if ageFloat >= 0.2*maxAgeFloat && ageFloat <= 0.5*maxAgeFloat && p.TickCounter == 0 {
			// 随机值
			dd := rand.Float64()
			if dd < 0.5 {
				w.spawnPlant(p)
			}
		}
	}

}

// lerpAngle 将当前角度 from 平滑地向目标角度 to 插值，
// t 为插值因子（0~1），始终走最短弧。
func lerpAngle(from, to, t float64) float64 {
	// from 和 to 插值小于0.001时，直接返回 to，避免抖动
	if math.Abs(from-to) < 0.001 {
		return to
	}
	diff := math.Mod(to-from+3*math.Pi, 2*math.Pi) - math.Pi // 归一化到 (-π, π]
	return from + diff*t
}

// 生物眼睛朝向更新规则：眼睛平滑转向速度方向
func updateEyeDirection(c *Creature) {
	// 注意：世界/屏幕坐标 Y 轴向下，而渲染用数学坐标系（Y 轴向上），
	// 所以对 VelocityY 取反，使角度在数学坐标系下正确。
	if c.VelocityX != 0 || c.VelocityY != 0 {
		targetDir := math.Atan2(-c.VelocityY, c.VelocityX)
		c.Direction = lerpAngle(c.Direction, targetDir, DefaultCreatureTurnSpeed)
		c.FocusDirection = c.Direction
	}
	// 如果静止，保持原朝向不变
}
