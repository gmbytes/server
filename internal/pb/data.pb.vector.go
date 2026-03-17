package pb

import (
	"fmt"
	"math"
	"math/rand/v2"
)

// ==================== 定点数常量 ====================

const (
	// Scale 定点数放大倍率。真实坐标 0.001 对应 int64 值 1。
	Scale    int64   = 1000
	ScaleF64 float64 = 1000.0

	deg2Rad = math.Pi / 180
	rad2Deg = 180 / math.Pi
)

// X 轴为角色面朝，角色右手为Y，左手坐标系

var (
	ZeroVector    = &Vector{X: 0, Y: 0, Z: 0}
	ForwardVector = &Vector{X: Scale, Y: 0, Z: 0}         // (1, 0, 0) 的定点数表示
	OneVector     = &Vector{X: Scale, Y: Scale, Z: Scale} // (1, 1, 1) 的定点数表示
)

// ==================== 构造 / 转换 ====================

// NewVector 从真实浮点坐标创建定点数向量（自动 ×Scale）
func NewVector(x, y, z float64) *Vector {
	return &Vector{
		X: floatToFixed(x),
		Y: floatToFixed(y),
		Z: floatToFixed(z),
	}
}

// NewVectorInt 直接从定点数整数值创建向量（不做缩放）
func NewVectorInt(x, y, z int64) *Vector {
	return &Vector{X: x, Y: y, Z: z}
}

// floatToFixed 将真实浮点数转为定点数（×Scale，四舍五入）
func floatToFixed(f float64) int64 {
	return int64(math.Round(f * ScaleF64))
}

// fixedToFloat 将定点数转为真实浮点数（÷Scale）
func fixedToFloat(v int64) float64 {
	return float64(v) / ScaleF64
}

// ToFloat64 返回真实浮点坐标 (x, y, z)
func (ss *Vector) ToFloat64() (float64, float64, float64) {
	return fixedToFloat(ss.X), fixedToFloat(ss.Y), fixedToFloat(ss.Z)
}

// Xf 返回 X 的真实浮点值
func (ss *Vector) Xf() float64 { return fixedToFloat(ss.X) }

// Yf 返回 Y 的真实浮点值
func (ss *Vector) Yf() float64 { return fixedToFloat(ss.Y) }

// Zf 返回 Z 的真实浮点值
func (ss *Vector) Zf() float64 { return fixedToFloat(ss.Z) }

// ==================== 字符串 ====================

// StringF 返回真实浮点坐标的字符串表示（避免与 pb 生成的 String 冲突）
func (ss *Vector) StringF() string {
	return fmt.Sprintf("(%.3f, %.3f, %.3f)", ss.Xf(), ss.Yf(), ss.Zf())
}

// ==================== 角度 / 弧度 ====================

// ToAngle2D 返回与 ForwardVector 在 XOY 平面的角度
func (ss *Vector) ToAngle2D() float64 {
	return ss.Angle2D(ForwardVector)
}

// Angle2D 返回与目标向量 XOY 平面的角度
func (ss *Vector) Angle2D(v *Vector) float64 {
	return ss.Radian2D(v) * rad2Deg
}

// ToRadian2D 返回与 ForwardVector 在 XOY 平面的弧度
func (ss *Vector) ToRadian2D() float64 {
	return ss.Radian2D(ForwardVector)
}

// Radian2D 返回与目标向量 XOY 平面的弧度
// 注意：这里用 float64 计算三角函数，定点数乘积的缩放在 sin/cos 中可以抵消
func (ss *Vector) Radian2D(v *Vector) float64 {
	// sin = ss.X*v.Y - v.X*ss.Y （同为 Scale² 级别，比例关系不变）
	// cos = ss.X*v.X + ss.Y*v.Y
	sin := float64(ss.X)*float64(v.Y) - float64(v.X)*float64(ss.Y)
	cos := float64(ss.X)*float64(v.X) + float64(ss.Y)*float64(v.Y)
	return -math.Atan2(sin, cos)
}

// ==================== 旋转 ====================

// Rotate2D 返回绕 Z 轴旋转后的向量，单位为弧度，左手坐标系
func (ss *Vector) Rotate2D(alpha float64) *Vector {
	sinA, cosA := math.Sincos(alpha)
	fx, fy := float64(ss.X), float64(ss.Y)
	return &Vector{
		X: int64(math.Round(fx*cosA - fy*sinA)),
		Y: int64(math.Round(fx*sinA + fy*cosA)),
		Z: ss.Z,
	}
}

// RotateAngle2D 返回绕 Z 轴旋转后的向量，单位为角度，左手坐标系
func (ss *Vector) RotateAngle2D(alphaDeg float64) *Vector {
	return ss.Rotate2D(alphaDeg * deg2Rad)
}

// ==================== 点乘 / 叉乘 ====================

// Dot2D 返回二维点积（定点数结果，值为真实点积 × Scale²）
// 如需真实值，请使用 Dot2DFloat
func (ss *Vector) Dot2D(v *Vector) int64 {
	return ss.X*v.X + ss.Y*v.Y
}

// Dot2DFloat 返回二维点积的真实浮点值
func (ss *Vector) Dot2DFloat(v *Vector) float64 {
	return float64(ss.Dot2D(v)) / (ScaleF64 * ScaleF64)
}

// Dot 返回三维点积（定点数结果，值为真实点积 × Scale²）
// 如需真实值，请使用 DotFloat
func (ss *Vector) Dot(v *Vector) int64 {
	return ss.X*v.X + ss.Y*v.Y + ss.Z*v.Z
}

// DotFloat 返回三维点积的真实浮点值
func (ss *Vector) DotFloat(v *Vector) float64 {
	return float64(ss.Dot(v)) / (ScaleF64 * ScaleF64)
}

// Cross 返回三维叉积（定点数结果，注意缩放关系：结果值 = 真实叉积 × Scale²）
// 如果只用于判断方向，可直接使用；如需真实值，需 ÷ Scale
func (ss *Vector) Cross(v *Vector) *Vector {
	return &Vector{
		X: (ss.Y*v.Z - ss.Z*v.Y) / Scale,
		Y: (ss.Z*v.X - ss.X*v.Z) / Scale,
		Z: (ss.X*v.Y - ss.Y*v.X) / Scale,
	}
}

// ==================== 长度 ====================

// LengthSq2D 返回二维长度的平方（定点数域，值 = 真实长度² × Scale²）
func (ss *Vector) LengthSq2D() int64 {
	return ss.X*ss.X + ss.Y*ss.Y
}

// LengthSq 返回三维长度的平方（定点数域，值 = 真实长度² × Scale²）
func (ss *Vector) LengthSq() int64 {
	return ss.X*ss.X + ss.Y*ss.Y + ss.Z*ss.Z
}

// LengthSq2DFloat 返回二维长度平方的真实浮点值
func (ss *Vector) LengthSq2DFloat() float64 {
	return float64(ss.LengthSq2D()) / (ScaleF64 * ScaleF64)
}

// LengthSqFloat 返回三维长度平方的真实浮点值
func (ss *Vector) LengthSqFloat() float64 {
	return float64(ss.LengthSq()) / (ScaleF64 * ScaleF64)
}

// Length2D 返回二维长度的真实浮点值
func (ss *Vector) Length2D() float64 {
	return math.Sqrt(float64(ss.LengthSq2D())) / ScaleF64
}

// Length 返回三维长度的真实浮点值
func (ss *Vector) Length() float64 {
	return math.Sqrt(float64(ss.LengthSq())) / ScaleF64
}

// ==================== 距离 ====================

// DistanceSq2D 返回二维距离的平方（定点数域）
func (ss *Vector) DistanceSq2D(v *Vector) int64 {
	dx, dy := ss.X-v.X, ss.Y-v.Y
	return dx*dx + dy*dy
}

// DistanceSq 返回三维距离的平方（定点数域）
func (ss *Vector) DistanceSq(v *Vector) int64 {
	dx, dy, dz := ss.X-v.X, ss.Y-v.Y, ss.Z-v.Z
	return dx*dx + dy*dy + dz*dz
}

// DistanceSq2DFloat 返回二维距离平方的真实浮点值
func (ss *Vector) DistanceSq2DFloat(v *Vector) float64 {
	return float64(ss.DistanceSq2D(v)) / (ScaleF64 * ScaleF64)
}

// DistanceSqFloat 返回三维距离平方的真实浮点值
func (ss *Vector) DistanceSqFloat(v *Vector) float64 {
	return float64(ss.DistanceSq(v)) / (ScaleF64 * ScaleF64)
}

// Distance2D 返回二维距离的真实浮点值
func (ss *Vector) Distance2D(v *Vector) float64 {
	return math.Sqrt(float64(ss.DistanceSq2D(v))) / ScaleF64
}

// Distance 返回三维距离的真实浮点值
func (ss *Vector) Distance(v *Vector) float64 {
	return math.Sqrt(float64(ss.DistanceSq(v))) / ScaleF64
}

// ==================== 相等比较 ====================

// Equal2D 判断 XY 是否完全相等
func (ss *Vector) Equal2D(v *Vector) bool {
	return ss.X == v.X && ss.Y == v.Y
}

// Equal 判断 XYZ 是否完全相等
func (ss *Vector) Equal(v *Vector) bool {
	return ss.X == v.X && ss.Y == v.Y && ss.Z == v.Z
}

// ApproximatelyEqual2D 判断 XY 是否近似相等（容差 1 个定点数单位，即真实 0.001）
func (ss *Vector) ApproximatelyEqual2D(v *Vector) bool {
	return abs64(ss.X-v.X) <= 1 && abs64(ss.Y-v.Y) <= 1
}

// ApproximatelyEqual 判断 XYZ 是否近似相等（容差 1 个定点数单位，即真实 0.001）
func (ss *Vector) ApproximatelyEqual(v *Vector) bool {
	return abs64(ss.X-v.X) <= 1 && abs64(ss.Y-v.Y) <= 1 && abs64(ss.Z-v.Z) <= 1
}

func abs64(v int64) int64 {
	if v < 0 {
		return -v
	}
	return v
}

// ==================== 正交 / 拷贝 / 反转 ====================

// Orthogonal2D 返回 XOY 平面上的正交向量（逆时针旋转90度）
func (ss *Vector) Orthogonal2D() *Vector {
	return &Vector{
		X: -ss.Y,
		Y: ss.X,
		Z: ss.Z,
	}
}

// Copy 深拷贝
func (ss *Vector) Copy() *Vector {
	return &Vector{
		X: ss.X,
		Y: ss.Y,
		Z: ss.Z,
	}
}

// CopyNewZ 拷贝但替换 Z（参数为真实浮点值）
func (ss *Vector) CopyNewZ(z float64) *Vector {
	return &Vector{
		X: ss.X,
		Y: ss.Y,
		Z: floatToFixed(z),
	}
}

// CopyNewZInt 拷贝但替换 Z（参数为定点数值）
func (ss *Vector) CopyNewZInt(z int64) *Vector {
	return &Vector{
		X: ss.X,
		Y: ss.Y,
		Z: z,
	}
}

// CopyTo 将自身值拷贝到 dst
func (ss *Vector) CopyTo(dst *Vector) {
	dst.X = ss.X
	dst.Y = ss.Y
	dst.Z = ss.Z
}

// Reverse2D 返回 XY 取反的向量（Z 不变）
func (ss *Vector) Reverse2D() *Vector {
	return &Vector{
		X: -ss.X,
		Y: -ss.Y,
		Z: ss.Z,
	}
}

// Reverse 返回 XYZ 均取反的向量
func (ss *Vector) Reverse() *Vector {
	return &Vector{
		X: -ss.X,
		Y: -ss.Y,
		Z: -ss.Z,
	}
}

// ==================== 加减乘除 ====================

// Add2D 二维加法（Z 保持自身值）
func (ss *Vector) Add2D(v *Vector) *Vector {
	return &Vector{
		X: ss.X + v.X,
		Y: ss.Y + v.Y,
		Z: ss.Z,
	}
}

// Add 三维加法
func (ss *Vector) Add(v *Vector) *Vector {
	return &Vector{
		X: ss.X + v.X,
		Y: ss.Y + v.Y,
		Z: ss.Z + v.Z,
	}
}

// Sub2D 二维减法（Z 保持自身值）
func (ss *Vector) Sub2D(v *Vector) *Vector {
	return &Vector{
		X: ss.X - v.X,
		Y: ss.Y - v.Y,
		Z: ss.Z,
	}
}

// Sub 三维减法
func (ss *Vector) Sub(v *Vector) *Vector {
	return &Vector{
		X: ss.X - v.X,
		Y: ss.Y - v.Y,
		Z: ss.Z - v.Z,
	}
}

// Mul2D 二维标量乘法（参数为整数倍率，Z 不变）
// 例如 Mul2D(2) 表示坐标×2
func (ss *Vector) Mul2D(v int64) *Vector {
	return &Vector{
		X: ss.X * v,
		Y: ss.Y * v,
		Z: ss.Z,
	}
}

// Mul 三维标量乘法（参数为整数倍率）
func (ss *Vector) Mul(v int64) *Vector {
	return &Vector{
		X: ss.X * v,
		Y: ss.Y * v,
		Z: ss.Z * v,
	}
}

// MulFloat2D 二维标量乘以浮点数（Z 不变）
func (ss *Vector) MulFloat2D(v float64) *Vector {
	return &Vector{
		X: int64(math.Round(float64(ss.X) * v)),
		Y: int64(math.Round(float64(ss.Y) * v)),
		Z: ss.Z,
	}
}

// MulFloat 三维标量乘以浮点数
func (ss *Vector) MulFloat(v float64) *Vector {
	return &Vector{
		X: int64(math.Round(float64(ss.X) * v)),
		Y: int64(math.Round(float64(ss.Y) * v)),
		Z: int64(math.Round(float64(ss.Z) * v)),
	}
}

// Div2D 二维标量除法（参数为整数倍率，Z 不变，除零返回自身拷贝）
func (ss *Vector) Div2D(v int64) *Vector {
	if v == 0 {
		return ss.Copy()
	}
	return &Vector{
		X: ss.X / v,
		Y: ss.Y / v,
		Z: ss.Z,
	}
}

// Div 三维标量除法（参数为整数倍率，除零返回自身拷贝）
func (ss *Vector) Div(v int64) *Vector {
	if v == 0 {
		return ss.Copy()
	}
	return &Vector{
		X: ss.X / v,
		Y: ss.Y / v,
		Z: ss.Z / v,
	}
}

// DivFloat2D 二维标量除以浮点数（Z 不变）
func (ss *Vector) DivFloat2D(v float64) *Vector {
	if v == 0 {
		return ss.Copy()
	}
	inv := 1.0 / v
	return &Vector{
		X: int64(math.Round(float64(ss.X) * inv)),
		Y: int64(math.Round(float64(ss.Y) * inv)),
		Z: ss.Z,
	}
}

// DivFloat 三维标量除以浮点数
func (ss *Vector) DivFloat(v float64) *Vector {
	if v == 0 {
		return ss.Copy()
	}
	inv := 1.0 / v
	return &Vector{
		X: int64(math.Round(float64(ss.X) * inv)),
		Y: int64(math.Round(float64(ss.Y) * inv)),
		Z: int64(math.Round(float64(ss.Z) * inv)),
	}
}

// ==================== 归一化 ====================

// Norm2D 返回 XOY 平面的单位向量（定点数表示，长度 = Scale），保留原始 Z 值
func (ss *Vector) Norm2D() *Vector {
	lenSq := ss.LengthSq2D()
	if lenSq == 0 {
		return &Vector{X: ForwardVector.X, Y: ForwardVector.Y, Z: ss.Z}
	}
	l := ScaleF64 / math.Sqrt(float64(lenSq))
	return &Vector{
		X: int64(math.Round(float64(ss.X) * l)),
		Y: int64(math.Round(float64(ss.Y) * l)),
		Z: ss.Z,
	}
}

// Norm 返回三维单位向量（定点数表示，长度 = Scale）
func (ss *Vector) Norm() *Vector {
	lenSq := ss.LengthSq()
	if lenSq == 0 {
		return ss.Copy()
	}
	scaleSq := Scale * Scale
	if lenSq == scaleSq {
		return ss.Copy()
	}
	l := ScaleF64 / math.Sqrt(float64(lenSq))
	return &Vector{
		X: int64(math.Round(float64(ss.X) * l)),
		Y: int64(math.Round(float64(ss.Y) * l)),
		Z: int64(math.Round(float64(ss.Z) * l)),
	}
}

// ==================== 随机 ====================

// GenerateRandomVector 在 [min, max] 范围内生成随机向量（参数为定点数向量）
func GenerateRandomVector(min, max *Vector) *Vector {
	return &Vector{
		X: int64(rand.Float64()*float64(max.X-min.X)) + min.X,
		Y: int64(rand.Float64()*float64(max.Y-min.Y)) + min.Y,
		Z: int64(rand.Float64()*float64(max.Z-min.Z)) + min.Z,
	}
}

// ==================== Lerp 插值 ====================

// Lerp 线性插值，t 为 [0, 1] 的浮点值
func (ss *Vector) Lerp(v *Vector, t float64) *Vector {
	return &Vector{
		X: ss.X + int64(math.Round(float64(v.X-ss.X)*t)),
		Y: ss.Y + int64(math.Round(float64(v.Y-ss.Y)*t)),
		Z: ss.Z + int64(math.Round(float64(v.Z-ss.Z)*t)),
	}
}

// Lerp2D 二维线性插值（Z 保持自身值）
func (ss *Vector) Lerp2D(v *Vector, t float64) *Vector {
	return &Vector{
		X: ss.X + int64(math.Round(float64(v.X-ss.X)*t)),
		Y: ss.Y + int64(math.Round(float64(v.Y-ss.Y)*t)),
		Z: ss.Z,
	}
}

// ==================== MoveTowards ====================

// MoveTowards 向目标移动，最大距离为 maxDist（真实浮点值）
func (ss *Vector) MoveTowards(target *Vector, maxDist float64) *Vector {
	dx, dy, dz := float64(target.X-ss.X), float64(target.Y-ss.Y), float64(target.Z-ss.Z)
	distSq := dx*dx + dy*dy + dz*dz
	maxFixed := maxDist * ScaleF64
	if distSq == 0 || distSq <= maxFixed*maxFixed {
		return target.Copy()
	}
	dist := math.Sqrt(distSq)
	ratio := maxFixed / dist
	return &Vector{
		X: ss.X + int64(math.Round(dx*ratio)),
		Y: ss.Y + int64(math.Round(dy*ratio)),
		Z: ss.Z + int64(math.Round(dz*ratio)),
	}
}

// MoveTowards2D 二维向目标移动（Z 保持自身值）
func (ss *Vector) MoveTowards2D(target *Vector, maxDist float64) *Vector {
	dx, dy := float64(target.X-ss.X), float64(target.Y-ss.Y)
	distSq := dx*dx + dy*dy
	maxFixed := maxDist * ScaleF64
	if distSq == 0 || distSq <= maxFixed*maxFixed {
		return &Vector{X: target.X, Y: target.Y, Z: ss.Z}
	}
	dist := math.Sqrt(distSq)
	ratio := maxFixed / dist
	return &Vector{
		X: ss.X + int64(math.Round(dx*ratio)),
		Y: ss.Y + int64(math.Round(dy*ratio)),
		Z: ss.Z,
	}
}

// ==================== 便捷方法 ====================

// IsZero 判断是否为零向量
func (ss *Vector) IsZero() bool {
	return ss.X == 0 && ss.Y == 0 && ss.Z == 0
}

// IsZero2D 判断 XY 是否为零
func (ss *Vector) IsZero2D() bool {
	return ss.X == 0 && ss.Y == 0
}

// SetFromFloat64 从真实浮点坐标设置（就地修改）
func (ss *Vector) SetFromFloat64(x, y, z float64) {
	ss.X = floatToFixed(x)
	ss.Y = floatToFixed(y)
	ss.Z = floatToFixed(z)
}

// Min 分量取最小值
func (ss *Vector) Min(v *Vector) *Vector {
	return &Vector{
		X: min64(ss.X, v.X),
		Y: min64(ss.Y, v.Y),
		Z: min64(ss.Z, v.Z),
	}
}

// Max 分量取最大值
func (ss *Vector) Max(v *Vector) *Vector {
	return &Vector{
		X: max64(ss.X, v.X),
		Y: max64(ss.Y, v.Y),
		Z: max64(ss.Z, v.Z),
	}
}

// Clamp 将向量分量限制在 [min, max] 范围内
func (ss *Vector) Clamp(minV, maxV *Vector) *Vector {
	return &Vector{
		X: clamp64(ss.X, minV.X, maxV.X),
		Y: clamp64(ss.Y, minV.Y, maxV.Y),
		Z: clamp64(ss.Z, minV.Z, maxV.Z),
	}
}

func min64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func max64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func clamp64(v, lo, hi int64) int64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
