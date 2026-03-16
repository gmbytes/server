package enum

import (
	"testing"
)

// -----------------------------------------------------------------------
// AttrType.Int32
// -----------------------------------------------------------------------

func TestAttrType_Int32_WhenConstitution_ExpectOne(t *testing.T) {
	if got := AttrType_Constitution.Int32(); got != 1 {
		t.Errorf("AttrType_Constitution.Int32() 期望 1，得到 %d", got)
	}
}

func TestAttrType_Int32_WhenInvalid_ExpectZero(t *testing.T) {
	if got := AttrType_Invalid.Int32(); got != 0 {
		t.Errorf("AttrType_Invalid.Int32() 期望 0，得到 %d", got)
	}
}

func TestAttrType_Int32_WhenHp_ExpectOneHundred(t *testing.T) {
	if got := AttrType_Hp.Int32(); got != 100 {
		t.Errorf("AttrType_Hp.Int32() 期望 100，得到 %d", got)
	}
}

func TestAttrType_Int32_WhenMp_ExpectOneHundredOne(t *testing.T) {
	if got := AttrType_Mp.Int32(); got != 101 {
		t.Errorf("AttrType_Mp.Int32() 期望 101，得到 %d", got)
	}
}

func TestAttrType_Int32_WhenMaxHp_ExpectThirteen(t *testing.T) {
	if got := AttrType_MaxHp.Int32(); got != 13 {
		t.Errorf("AttrType_MaxHp.Int32() 期望 13，得到 %d", got)
	}
}

// -----------------------------------------------------------------------
// AttrType 常量值验证
// -----------------------------------------------------------------------

func TestAttrType_Constants_ExpectCorrectValues(t *testing.T) {
	cases := []struct {
		name string
		attr AttrType
		want int32
	}{
		{"Invalid", AttrType_Invalid, 0},
		{"Constitution", AttrType_Constitution, 1},
		{"Strength", AttrType_Strength, 2},
		{"Intelligence", AttrType_Intelligence, 3},
		{"Endurance", AttrType_Endurance, 4},
		{"Agility", AttrType_Agility, 5},
		{"MaxHp", AttrType_MaxHp, 13},
		{"MaxMp", AttrType_MaxMp, 14},
		{"Speed", AttrType_Speed, 15},
		{"PhyAttack", AttrType_PhyAttack, 16},
		{"PhyDefense", AttrType_PhyDefense, 17},
		{"MagicAttack", AttrType_MagicAttack, 18},
		{"MagicDefense", AttrType_MagicDefense, 19},
		{"PhyDamageBonus", AttrType_PhyDamageBonus, 30},
		{"PhyDamageReduction", AttrType_PhyDamageReduction, 31},
		{"MagicDamageBonus", AttrType_MagicDamageBonus, 32},
		{"MagicDamageReduction", AttrType_MagicDamageReduction, 33},
		{"ControlEnhancement", AttrType_ControlEnhancement, 34},
		{"ControlResistance", AttrType_ControlResistance, 35},
		{"HealingEnhancement", AttrType_HealingEnhancement, 36},
		{"HealingReceivedBonus", AttrType_HealingReceivedBonus, 37},
		{"PhyCritRate", AttrType_PhyCritRate, 50},
		{"PhysicalCritDamage", AttrType_PhysicalCritDamage, 51},
		{"MagicCritRate", AttrType_MagicCritRate, 52},
		{"MagicCritDamage", AttrType_MagicCritDamage, 53},
		{"Hp", AttrType_Hp, 100},
		{"Mp", AttrType_Mp, 101},
	}

	for _, c := range cases {
		got := c.attr.Int32()
		if got != c.want {
			t.Errorf("AttrType_%s.Int32() 期望 %d，得到 %d", c.name, c.want, got)
		}
	}
}

// -----------------------------------------------------------------------
// AttrType 互不相同
// -----------------------------------------------------------------------

func TestAttrType_Constants_ExpectAllUnique(t *testing.T) {
	attrs := []AttrType{
		AttrType_Invalid,
		AttrType_Constitution,
		AttrType_Strength,
		AttrType_Intelligence,
		AttrType_Endurance,
		AttrType_Agility,
		AttrType_MaxHp,
		AttrType_MaxMp,
		AttrType_Speed,
		AttrType_PhyAttack,
		AttrType_PhyDefense,
		AttrType_MagicAttack,
		AttrType_MagicDefense,
		AttrType_PhyDamageBonus,
		AttrType_PhyDamageReduction,
		AttrType_MagicDamageBonus,
		AttrType_MagicDamageReduction,
		AttrType_ControlEnhancement,
		AttrType_ControlResistance,
		AttrType_HealingEnhancement,
		AttrType_HealingReceivedBonus,
		AttrType_PhysicalDefensePenetrationRate,
		AttrType_MagicDefensePenetrationRate,
		AttrType_PhyCritRate,
		AttrType_PhysicalCritDamage,
		AttrType_MagicCritRate,
		AttrType_MagicCritDamage,
		AttrType_PhysicalHitRate,
		AttrType_PhysicalDodgeRate,
		AttrType_MagicHitRate,
		AttrType_MagicDodgeRate,
		AttrType_HealingCritRate,
		AttrType_ControlHitRate,
		AttrType_ControlDodgeRate,
		AttrType_Hp,
		AttrType_Mp,
	}

	seen := make(map[int32]bool)
	for _, a := range attrs {
		v := a.Int32()
		if seen[v] {
			t.Errorf("AttrType 值 %d 重复", v)
		}
		seen[v] = true
	}
}

// -----------------------------------------------------------------------
// EntityType.Int
// -----------------------------------------------------------------------

func TestEntityType_Int_WhenRole_ExpectOne(t *testing.T) {
	if got := EntityType_Role.Int(); got != 1 {
		t.Errorf("EntityType_Role.Int() 期望 1，得到 %d", got)
	}
}

func TestEntityType_Int_WhenNpc_ExpectTwo(t *testing.T) {
	if got := EntityType_Npc.Int(); got != 2 {
		t.Errorf("EntityType_Npc.Int() 期望 2，得到 %d", got)
	}
}

func TestEntityType_Int_WhenMax_ExpectThree(t *testing.T) {
	if got := EntityType_Max.Int(); got != 3 {
		t.Errorf("EntityType_Max.Int() 期望 3，得到 %d", got)
	}
}

// -----------------------------------------------------------------------
// EntityType 常量顺序：Role < Npc < Max
// -----------------------------------------------------------------------

func TestEntityType_Constants_ExpectAscendingOrder(t *testing.T) {
	if !(EntityType_Role < EntityType_Npc) {
		t.Error("期望 EntityType_Role < EntityType_Npc")
	}
	if !(EntityType_Npc < EntityType_Max) {
		t.Error("期望 EntityType_Npc < EntityType_Max")
	}
}
