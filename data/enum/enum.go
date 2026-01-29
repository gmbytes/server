package enum

type AttrType int32

func (a AttrType) Int32() int32 {
	return int32(a)
}

const (
	AttrType_Invalid                        AttrType = 0   // 无效属性
	AttrType_Constitution                   AttrType = 1   // 体质
	AttrType_Strength                       AttrType = 2   // 力量
	AttrType_Intelligence                   AttrType = 3   // 智力
	AttrType_Endurance                      AttrType = 4   // 耐力
	AttrType_Agility                        AttrType = 5   // 敏捷
	AttrType_MaxHp                          AttrType = 13  // 最大Hp
	AttrType_MaxMp                          AttrType = 14  // 最大Mp
	AttrType_Speed                          AttrType = 15  // 速度
	AttrType_PhyAttack                      AttrType = 16  // 物理攻击
	AttrType_PhyDefense                     AttrType = 17  // 物理防御
	AttrType_MagicAttack                    AttrType = 18  // 法术攻击
	AttrType_MagicDefense                   AttrType = 19  // 法术防御
	AttrType_PhyDamageBonus                 AttrType = 30  // 物攻增伤
	AttrType_PhyDamageReduction             AttrType = 31  // 物理减伤
	AttrType_MagicDamageBonus               AttrType = 32  // 法攻增伤
	AttrType_MagicDamageReduction           AttrType = 33  // 法术减伤
	AttrType_ControlEnhancement             AttrType = 34  // 控制增强
	AttrType_ControlResistance              AttrType = 35  // 控制抗性
	AttrType_HealingEnhancement             AttrType = 36  // 治疗增强
	AttrType_HealingReceivedBonus           AttrType = 37  // 受治疗增强
	AttrType_PhysicalDefensePenetrationRate AttrType = 38  // 物理防御穿透,忽视对方物理防御百分比
	AttrType_MagicDefensePenetrationRate    AttrType = 39  // 法术防御穿透,忽视对方法术防御百分比
	AttrType_PhyCritRate                    AttrType = 50  // 物理暴击
	AttrType_PhysicalCritDamage             AttrType = 51  // 物理暴伤
	AttrType_MagicCritRate                  AttrType = 52  // 法术暴击
	AttrType_MagicCritDamage                AttrType = 53  // 法术暴伤
	AttrType_PhysicalHitRate                AttrType = 54  // 物理命中率
	AttrType_PhysicalDodgeRate              AttrType = 55  // 物理闪避
	AttrType_MagicHitRate                   AttrType = 56  // 法术命中
	AttrType_MagicDodgeRate                 AttrType = 57  // 法术闪避
	AttrType_HealingCritRate                AttrType = 58  // 治疗暴击
	AttrType_ControlHitRate                 AttrType = 59  // 控制命中
	AttrType_ControlDodgeRate               AttrType = 60  // 控制闪避
	AttrType_Hp                             AttrType = 100 // 血量
	AttrType_Mp                             AttrType = 101 // 蓝量
)

type EntityType int8

func (a EntityType) Int() int32 {
	return int32(a)
}

const (
	EntityType_Role EntityType = 1 //
	EntityType_Npc  EntityType = 2 //
	EntityType_Max  EntityType = 3 //
)
