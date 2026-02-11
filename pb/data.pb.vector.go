package pb

import (
	"fmt"
	"math"
	"math/rand/v2"
)

const (
	deg2Rad = math.Pi / 180
	rad2Deg = 180 / math.Pi
)

// X 轴为角色面朝，角色右手为Y，左手坐标系

var ZeroVector = &Vector{X: 0, Y: 0, Z: 0}
var ForwardVector = &Vector{X: 1, Y: 0, Z: 0}
var OneVector = &Vector{X: 1, Y: 1, Z: 1}

// NewVector 从 float64 创建一个向量（纯 float64，不再使用定点缩放）
func NewVector(x, y, z float64) *Vector {
	return &Vector{X: x, Y: y, Z: z}
}

// StringF 返回浮点格式的字符串表示（避免与 pb 生成的 String 冲突）
func (ss *Vector) StringF() string {
	return fmt.Sprintf("(%.5f, %.5f, %.5f)", ss.X, ss.Y, ss.Z)
}

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
func (ss *Vector) Radian2D(v *Vector) float64 {
	sin := ss.X*v.Y - v.X*ss.Y
	cos := ss.X*v.X + ss.Y*v.Y
	return -math.Atan2(sin, cos)
}

// Rotate2D 返回绕 Z 轴旋转后的向量，单位为弧度，左手坐标系
func (ss *Vector) Rotate2D(alpha float64) *Vector {
	sinA, cosA := math.Sincos(alpha)
	return &Vector{
		X: ss.X*cosA - ss.Y*sinA,
		Y: ss.X*sinA + ss.Y*cosA,
		Z: ss.Z,
	}
}

// RotateAngle2D 返回绕 Z 轴旋转后的向量，单位为角度，左手坐标系
func (ss *Vector) RotateAngle2D(alphaDeg float64) *Vector {
	return ss.Rotate2D(alphaDeg * deg2Rad)
}

func (ss *Vector) Dot2D(v *Vector) float64 {
	return ss.X*v.X + ss.Y*v.Y
}

func (ss *Vector) Dot(v *Vector) float64 {
	return ss.X*v.X + ss.Y*v.Y + ss.Z*v.Z
}

func (ss *Vector) Cross(v *Vector) *Vector {
	return &Vector{
		X: ss.Y*v.Z - ss.Z*v.Y,
		Y: ss.Z*v.X - ss.X*v.Z,
		Z: ss.X*v.Y - ss.Y*v.X,
	}
}

func (ss *Vector) LengthSq2D() float64 {
	return ss.X*ss.X + ss.Y*ss.Y
}

func (ss *Vector) LengthSq() float64 {
	return ss.X*ss.X + ss.Y*ss.Y + ss.Z*ss.Z
}

func (ss *Vector) Length2D() float64 {
	return math.Sqrt(ss.LengthSq2D())
}

func (ss *Vector) Length() float64 {
	return math.Sqrt(ss.LengthSq())
}

func (ss *Vector) DistanceSq2D(v *Vector) float64 {
	dx, dy := ss.X-v.X, ss.Y-v.Y
	return dx*dx + dy*dy
}

func (ss *Vector) DistanceSq(v *Vector) float64 {
	dx, dy, dz := ss.X-v.X, ss.Y-v.Y, ss.Z-v.Z
	return dx*dx + dy*dy + dz*dz
}

func (ss *Vector) Distance2D(v *Vector) float64 {
	return math.Sqrt(ss.DistanceSq2D(v))
}

func (ss *Vector) Distance(v *Vector) float64 {
	return math.Sqrt(ss.DistanceSq(v))
}

func (ss *Vector) Equal2D(v *Vector) bool {
	return ss.X == v.X && ss.Y == v.Y
}

func (ss *Vector) Equal(v *Vector) bool {
	return ss.X == v.X && ss.Y == v.Y && ss.Z == v.Z
}

func (ss *Vector) ApproximatelyEqual2D(v *Vector) bool {
	return math.Abs(ss.X-v.X) < 1e-7 && math.Abs(ss.Y-v.Y) < 1e-7
}

func (ss *Vector) ApproximatelyEqual(v *Vector) bool {
	return math.Abs(ss.X-v.X) < 1e-7 && math.Abs(ss.Y-v.Y) < 1e-7 && math.Abs(ss.Z-v.Z) < 1e-7
}

func (ss *Vector) Orthogonal2D() *Vector {
	return &Vector{
		X: -ss.Y,
		Y: ss.X,
		Z: ss.Z,
	}
}

func (ss *Vector) Copy() *Vector {
	return &Vector{
		X: ss.X,
		Y: ss.Y,
		Z: ss.Z,
	}
}

func (ss *Vector) CopyNewZ(z float64) *Vector {
	return &Vector{
		X: ss.X,
		Y: ss.Y,
		Z: z,
	}
}

func (ss *Vector) CopyTo(dst *Vector) {
	dst.X = ss.X
	dst.Y = ss.Y
	dst.Z = ss.Z
}

func (ss *Vector) Reverse2D() *Vector {
	return &Vector{
		X: -ss.X,
		Y: -ss.Y,
		Z: ss.Z,
	}
}

func (ss *Vector) Reverse() *Vector {
	return &Vector{
		X: -ss.X,
		Y: -ss.Y,
		Z: -ss.Z,
	}
}

func (ss *Vector) Add2D(v *Vector) *Vector {
	return &Vector{
		X: ss.X + v.X,
		Y: ss.Y + v.Y,
		Z: ss.Z,
	}
}

func (ss *Vector) Add(v *Vector) *Vector {
	return &Vector{
		X: ss.X + v.X,
		Y: ss.Y + v.Y,
		Z: ss.Z + v.Z,
	}
}

func (ss *Vector) Sub2D(v *Vector) *Vector {
	return &Vector{
		X: ss.X - v.X,
		Y: ss.Y - v.Y,
		Z: ss.Z,
	}
}

func (ss *Vector) Sub(v *Vector) *Vector {
	return &Vector{
		X: ss.X - v.X,
		Y: ss.Y - v.Y,
		Z: ss.Z - v.Z,
	}
}

func (ss *Vector) Mul2D(v float64) *Vector {
	return &Vector{
		X: ss.X * v,
		Y: ss.Y * v,
		Z: ss.Z,
	}
}

func (ss *Vector) Mul(v float64) *Vector {
	return &Vector{
		X: ss.X * v,
		Y: ss.Y * v,
		Z: ss.Z * v,
	}
}

func (ss *Vector) Div2D(v float64) *Vector {
	if v == 0 {
		return ss.Copy()
	}
	inv := 1 / v
	return &Vector{
		X: ss.X * inv,
		Y: ss.Y * inv,
		Z: ss.Z,
	}
}

func (ss *Vector) Div(v float64) *Vector {
	if v == 0 {
		return ss.Copy()
	}
	inv := 1 / v
	return &Vector{
		X: ss.X * inv,
		Y: ss.Y * inv,
		Z: ss.Z * inv,
	}
}

func (ss *Vector) Norm2D() *Vector {
	lenSq := ss.LengthSq2D()
	if lenSq == 0 {
		return ForwardVector.Copy()
	}
	l := 1 / math.Sqrt(lenSq)
	return &Vector{
		X: ss.X * l,
		Y: ss.Y * l,
		Z: ss.Z,
	}
}

func (ss *Vector) Norm() *Vector {
	lenSq := ss.LengthSq()
	if lenSq == 0 {
		return ss.Copy()
	}
	if lenSq == 1 {
		return ss.Copy()
	}
	l := 1 / math.Sqrt(lenSq)
	return &Vector{
		X: ss.X * l,
		Y: ss.Y * l,
		Z: ss.Z * l,
	}
}

func GenerateRandomVector(min, max *Vector) *Vector {
	return &Vector{
		X: rand.Float64()*(max.X-min.X) + min.X,
		Y: rand.Float64()*(max.Y-min.Y) + min.Y,
		Z: rand.Float64()*(max.Z-min.Z) + min.Z,
	}
}
