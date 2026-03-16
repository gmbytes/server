package data

import (
	"testing"

	"server/internal/data/enum"
)

// -----------------------------------------------------------------------
// Attrs.GetValue
// -----------------------------------------------------------------------

func TestAttrs_GetValue_WhenTypeExists_ExpectCorrectValue(t *testing.T) {
	attrs := Attrs{
		{Type: enum.AttrType_MaxHp, Val: 1000, Rate: 0},
		{Type: enum.AttrType_MaxMp, Val: 500, Rate: 0},
		{Type: enum.AttrType_Speed, Val: 150, Rate: 0},
	}

	if got := attrs.GetValue(enum.AttrType_MaxHp); got != 1000 {
		t.Errorf("GetValue(MaxHp) 期望 1000，得到 %d", got)
	}
	if got := attrs.GetValue(enum.AttrType_MaxMp); got != 500 {
		t.Errorf("GetValue(MaxMp) 期望 500，得到 %d", got)
	}
	if got := attrs.GetValue(enum.AttrType_Speed); got != 150 {
		t.Errorf("GetValue(Speed) 期望 150，得到 %d", got)
	}
}

func TestAttrs_GetValue_WhenTypeNotExists_ExpectZero(t *testing.T) {
	attrs := Attrs{
		{Type: enum.AttrType_MaxHp, Val: 1000, Rate: 0},
	}

	if got := attrs.GetValue(enum.AttrType_MaxMp); got != 0 {
		t.Errorf("GetValue(不存在的类型) 期望 0，得到 %d", got)
	}
}

func TestAttrs_GetValue_WhenEmpty_ExpectZero(t *testing.T) {
	attrs := Attrs{}

	if got := attrs.GetValue(enum.AttrType_Hp); got != 0 {
		t.Errorf("空 Attrs.GetValue 期望 0，得到 %d", got)
	}
}

func TestAttrs_GetValue_WhenInvalidType_ExpectZero(t *testing.T) {
	attrs := Attrs{
		{Type: enum.AttrType_Constitution, Val: 50, Rate: 0},
	}

	if got := attrs.GetValue(enum.AttrType_Invalid); got != 0 {
		t.Errorf("GetValue(Invalid) 期望 0，得到 %d", got)
	}
}

func TestAttrs_GetValue_WhenMultipleAttrsWithDifferentTypes_ExpectCorrectLookup(t *testing.T) {
	attrs := Attrs{
		{Type: enum.AttrType_PhyAttack, Val: 200, Rate: 10},
		{Type: enum.AttrType_MagicAttack, Val: 300, Rate: 20},
		{Type: enum.AttrType_PhyDefense, Val: 100, Rate: 5},
		{Type: enum.AttrType_MagicDefense, Val: 150, Rate: 8},
	}

	cases := []struct {
		ty   enum.AttrType
		want int64
	}{
		{enum.AttrType_PhyAttack, 200},
		{enum.AttrType_MagicAttack, 300},
		{enum.AttrType_PhyDefense, 100},
		{enum.AttrType_MagicDefense, 150},
	}

	for _, c := range cases {
		if got := attrs.GetValue(c.ty); got != c.want {
			t.Errorf("GetValue(%d) 期望 %d，得到 %d", c.ty, c.want, got)
		}
	}
}

func TestAttrs_GetValue_WhenNegativeValue_ExpectNegative(t *testing.T) {
	attrs := Attrs{
		{Type: enum.AttrType_Speed, Val: -10, Rate: 0},
	}

	if got := attrs.GetValue(enum.AttrType_Speed); got != -10 {
		t.Errorf("GetValue 期望 -10，得到 %d", got)
	}
}

func TestAttrs_GetValue_WhenZeroValue_ExpectZero(t *testing.T) {
	attrs := Attrs{
		{Type: enum.AttrType_Agility, Val: 0, Rate: 0},
	}

	if got := attrs.GetValue(enum.AttrType_Agility); got != 0 {
		t.Errorf("GetValue(零值属性) 期望 0，得到 %d", got)
	}
}

// -----------------------------------------------------------------------
// Attr 结构体字段
// -----------------------------------------------------------------------

func TestAttr_Fields_WhenConstructed_ExpectCorrectValues(t *testing.T) {
	a := &Attr{
		Type: enum.AttrType_Strength,
		Val:  99,
		Rate: 50,
	}

	if a.Type != enum.AttrType_Strength {
		t.Errorf("Attr.Type 期望 Strength，得到 %d", a.Type)
	}
	if a.Val != 99 {
		t.Errorf("Attr.Val 期望 99，得到 %d", a.Val)
	}
	if a.Rate != 50 {
		t.Errorf("Attr.Rate 期望 50，得到 %d", a.Rate)
	}
}

// -----------------------------------------------------------------------
// EntityInitData
// -----------------------------------------------------------------------

func TestEntityInitData_WhenAttrsSet_ExpectCorrectAccess(t *testing.T) {
	attrs := &Attrs{
		{Type: enum.AttrType_Hp, Val: 5000, Rate: 0},
	}
	data := &EntityInitData{Attrs: attrs}

	if data.Attrs == nil {
		t.Error("EntityInitData.Attrs 不应为 nil")
	}

	if got := data.Attrs.GetValue(enum.AttrType_Hp); got != 5000 {
		t.Errorf("EntityInitData 的属性查找期望 5000，得到 %d", got)
	}
}

func TestEntityInitData_WhenNilAttrs_ExpectNilAttrs(t *testing.T) {
	data := &EntityInitData{Attrs: nil}
	if data.Attrs != nil {
		t.Error("EntityInitData.Attrs 应为 nil")
	}
}
