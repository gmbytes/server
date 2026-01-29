package matrix

import (
	"fmt"
	"math"
)

var ZeroVector3D = &Vector3D{X: 0, Y: 0, Z: 0}
var ForwardVector3D = &Vector3D{X: 1, Y: 0, Z: 0}
var OneVector3D = &Vector3D{X: 1, Y: 1, Z: 1}

// X 轴为角色面朝，角色右手为Y，左手坐标系

type Vector3D struct {
	X float64 //前
	Y float64 //右
	Z float64 //上
}

func (ss *Vector3D) String() string {
	return fmt.Sprintf("(%.5f, %.5f, %.5f)", ss.X, ss.Y, ss.Z)
}

// ToAngle2D 返回与 ForwardVector3D 在 XOY 平面的角度
func (ss *Vector3D) ToAngle2D() float64 {
	return ss.Angle2D(ForwardVector3D)
}

// Angle2D 返回与目标向量 XOY 平面的角度
func (ss *Vector3D) Angle2D(v *Vector3D) float64 {
	return ss.Radian2D(v) * 180 / math.Pi
}

// ToRadian2D 返回与 ForwardVector3D 在 XOY 平面的弧度
func (ss *Vector3D) ToRadian2D() float64 {
	return ss.Radian2D(ForwardVector3D)
}

// Radian2D 返回与目标向量 XOY 平面的弧度
func (ss *Vector3D) Radian2D(v *Vector3D) float64 {
	sin := ss.X*v.Y - v.X*ss.Y
	cos := ss.X*v.X + ss.Y*v.Y
	return -math.Atan2(sin, cos)
}

// Rotate2D 返回绕 Z 轴旋转后的向量，单位为弧度，左手坐标系
func (ss *Vector3D) Rotate2D(alpha float64) *Vector3D {
	return &Vector3D{
		X: ss.X*math.Cos(alpha) - ss.Y*math.Sin(alpha),
		Y: ss.X*math.Sin(alpha) + ss.Y*math.Cos(alpha),
		Z: ss.Z,
	}
}

// RotateAngle2D 返回绕 Z 轴旋转后的向量，单位为角度，左手坐标系
func (ss *Vector3D) RotateAngle2D(alphaDeg float64) *Vector3D {
	return ss.Rotate2D(alphaDeg * math.Pi / 180)
}

func (ss *Vector3D) Dot2D(v *Vector3D) float64 {
	return ss.X*v.X + ss.Y*v.Y
}

func (ss *Vector3D) Dot(v *Vector3D) float64 {
	return ss.X*v.X + ss.Y*v.Y + ss.Z*v.Z
}

func (ss *Vector3D) Cross(v *Vector3D) *Vector3D {
	return &Vector3D{
		X: ss.Y*v.Z - ss.Z*v.Y,
		Y: ss.Z*v.X - ss.X*v.Z,
		Z: ss.X*v.Y - ss.Y*v.X,
	}
}

func (ss *Vector3D) LengthSq2D() float64 {
	return ss.X*ss.X + ss.Y*ss.Y
}

func (ss *Vector3D) LengthSq() float64 {
	return ss.X*ss.X + ss.Y*ss.Y + ss.Z*ss.Z
}

func (ss *Vector3D) Length2D() float64 {
	return math.Sqrt(ss.LengthSq2D())
}

func (ss *Vector3D) Length() float64 {
	return math.Sqrt(ss.LengthSq())
}

func (ss *Vector3D) DistanceSq2D(v *Vector3D) float64 {
	return (ss.X-v.X)*(ss.X-v.X) + (ss.Y-v.Y)*(ss.Y-v.Y)
}

func (ss *Vector3D) DistanceSq(v *Vector3D) float64 {
	return (ss.X-v.X)*(ss.X-v.X) + (ss.Y-v.Y)*(ss.Y-v.Y) + (ss.Z-v.Z)*(ss.Z-v.Z)
}

func (ss *Vector3D) Distance2D(v *Vector3D) float64 {
	return math.Sqrt(ss.DistanceSq2D(v))
}

func (ss *Vector3D) Distance(v *Vector3D) float64 {
	return math.Sqrt(ss.DistanceSq(v))
}

func (ss *Vector3D) Equal2D(v *Vector3D) bool {
	return ss.X == v.X && ss.Y == v.Y
}

func (ss *Vector3D) Equal(v *Vector3D) bool {
	return ss.X == v.X && ss.Y == v.Y && ss.Z == v.Z
}

func (ss *Vector3D) ApproximatelyEqual2D(v *Vector3D) bool {
	return math.Abs(ss.X-v.X) < 1e-7 && math.Abs(ss.Y-v.Y) < 1e-7
}

func (ss *Vector3D) ApproximatelyEqual(v *Vector3D) bool {
	return math.Abs(ss.X-v.X) < 1e-7 && math.Abs(ss.Y-v.Y) < 1e-7 && math.Abs(ss.Z-v.Z) < 1e-7
}

func (ss *Vector3D) Orthogonal2D() *Vector3D {
	return &Vector3D{
		X: -ss.Y,
		Y: ss.X,
		Z: ss.Z,
	}
}

func (ss *Vector3D) Copy() *Vector3D {
	return &Vector3D{
		X: ss.X,
		Y: ss.Y,
		Z: ss.Z,
	}
}

func (ss *Vector3D) CopyNewZ(z float64) *Vector3D {
	return &Vector3D{
		X: ss.X,
		Y: ss.Y,
		Z: z,
	}
}

func (ss *Vector3D) CopyTo(dst *Vector3D) {
	dst.X = ss.X
	dst.Y = ss.Y
	dst.Z = ss.Z
}

func (ss *Vector3D) Reverse2D() *Vector3D {
	return &Vector3D{
		X: -ss.X,
		Y: -ss.Y,
		Z: ss.Z,
	}
}

func (ss *Vector3D) Reverse() *Vector3D {
	return &Vector3D{
		X: -ss.X,
		Y: -ss.Y,
		Z: -ss.Z,
	}
}

func (ss *Vector3D) Add2D(v *Vector3D) *Vector3D {
	return &Vector3D{
		X: ss.X + v.X,
		Y: ss.Y + v.Y,
		Z: ss.Z,
	}
}

func (ss *Vector3D) Add(v *Vector3D) *Vector3D {
	return &Vector3D{
		X: ss.X + v.X,
		Y: ss.Y + v.Y,
		Z: ss.Z + v.Z,
	}
}

func (ss *Vector3D) Sub2D(v *Vector3D) *Vector3D {
	return &Vector3D{
		X: ss.X - v.X,
		Y: ss.Y - v.Y,
		Z: ss.Z,
	}
}

func (ss *Vector3D) Sub(v *Vector3D) *Vector3D {
	return &Vector3D{
		X: ss.X - v.X,
		Y: ss.Y - v.Y,
		Z: ss.Z - v.Z,
	}
}

func (ss *Vector3D) Mul2D(v float64) *Vector3D {
	return &Vector3D{
		X: ss.X * v,
		Y: ss.Y * v,
		Z: ss.Z,
	}
}

func (ss *Vector3D) Mul(v float64) *Vector3D {
	return &Vector3D{
		X: ss.X * v,
		Y: ss.Y * v,
		Z: ss.Z * v,
	}
}

func (ss *Vector3D) Div2D(v float64) *Vector3D {
	return &Vector3D{
		X: ss.X / v,
		Y: ss.Y / v,
		Z: ss.Z,
	}
}

func (ss *Vector3D) Div(v float64) *Vector3D {
	return &Vector3D{
		X: ss.X / v,
		Y: ss.Y / v,
		Z: ss.Z / v,
	}
}

func (ss *Vector3D) Norm2D() *Vector3D {
	lenSq := ss.LengthSq2D()
	if lenSq == 0 {
		return ForwardVector3D.Copy()
	}

	l := 1 / math.Sqrt(lenSq)
	return &Vector3D{
		X: ss.X * l,
		Y: ss.Y * l,
		Z: ss.Z,
	}
}

func (ss *Vector3D) Norm() *Vector3D {
	length := ss.Length()
	if length == 0 || length == 1 {
		return ss
	}

	l := 1 / length
	return &Vector3D{
		X: ss.X * l,
		Y: ss.Y * l,
		Z: ss.Z * l,
	}
}
