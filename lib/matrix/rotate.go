package matrix

import "math"

//	下面的代码是豆包生成的
//
// Matrix3x3 定义 3x3 矩阵
type Matrix3x3 [3][3]float64

// multiply 矩阵与向量相乘
func (m Matrix3x3) multiply(v *Vector3D) *Vector3D {
	return &Vector3D{
		X: m[0][0]*v.X + m[0][1]*v.Y + m[0][2]*v.Z,
		Y: m[1][0]*v.X + m[1][1]*v.Y + m[1][2]*v.Z,
		Z: m[2][0]*v.X + m[2][1]*v.Y + m[2][2]*v.Z,
	}
}

// multiplyMatrices 矩阵相乘
func multiplyMatrices(m1, m2 Matrix3x3) Matrix3x3 {
	var result Matrix3x3
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			for k := 0; k < 3; k++ {
				result[i][j] += m1[i][k] * m2[k][j]
			}
		}
	}
	return result
}

// rotationMatrixX 计算绕 X 轴旋转的矩阵
func rotationMatrixX(a float64) Matrix3x3 {
	cosA := math.Cos(a)
	sinA := math.Sin(a)
	return Matrix3x3{
		{1, 0, 0},
		{0, cosA, sinA},
		{0, -sinA, cosA},
	}
}

// rotationMatrixY 计算绕 Y 轴旋转的矩阵
func rotationMatrixY(b float64) Matrix3x3 {
	cosB := math.Cos(b)
	sinB := math.Sin(b)
	return Matrix3x3{
		{cosB, 0, -sinB},
		{0, 1, 0},
		{sinB, 0, cosB},
	}
}

// rotationMatrixZ 计算绕 Z 轴旋转的矩阵
func rotationMatrixZ(c float64) Matrix3x3 {
	cosC := math.Cos(c)
	sinC := math.Sin(c)
	return Matrix3x3{
		{cosC, sinC, 0},
		{-sinC, cosC, 0},
		{0, 0, 1},
	}
}

// 在 UE 中，Rotator 旋转默认采用 Z - Y - X 旋转顺序，即 Yaw - Pitch - Roll 顺序。
// 具体来说，物体首先绕 Z 轴（Yaw）进行旋转，然后绕 Y 轴（Pitch）旋转，最后绕 X 轴（Roll）旋转。
// RotateObject 计算物体绕自身坐标系 Z、Y、X 轴旋转后的坐标
func RotateObject(point *Vector3D, rotation *Vector3D) *Vector3D {
	rz := rotationMatrixZ(rotation.Z)
	ry := rotationMatrixY(rotation.Y)
	rx := rotationMatrixX(rotation.X)

	rzy := multiplyMatrices(rz, ry)
	rxyz := multiplyMatrices(rzy, rx)

	return rxyz.multiply(point)
}

// func RotateObject2(point *Vector3D, rotation *Vector3D) *Vector3D {
// 	temp1 := RotatePointAroundAxis(point, &Vector3D{X: 0, Y: 0, Z: 1}, rotation.Z, nil)
// 	temp2 := RotatePointAroundAxis(temp1, &Vector3D{X: 0, Y: 1, Z: 0}, rotation.Y, nil)
// 	return RotatePointAroundAxis(temp2, &Vector3D{X: 1, Y: 0, Z: 0}, rotation.X, nil)
// }

// RotatePointAroundAxis 使用罗德里格旋转公式计算点围绕任意轴旋转
// point: 要旋转的点的世界坐标
// axis: 旋转轴向量（会被自动归一化）
// angle: 旋转角度（弧度）
// center: 旋转中心点（可选，nil则默认为原点）
func RotatePointAroundAxis(point *Vector3D, axis *Vector3D, angle float64, center *Vector3D) *Vector3D {
	// 如果没有指定旋转中心，默认为原点
	if center == nil {
		center = &Vector3D{X: 0, Y: 0, Z: 0}
	}

	// 将点平移到以旋转中心为原点的坐标系
	p := &Vector3D{
		X: point.X - center.X,
		Y: point.Y - center.Y,
		Z: point.Z - center.Z,
	}

	// 归一化旋转轴
	k := axis.Norm()

	// 罗德里格旋转公式: v' = v*cos(θ) + (k×v)*sin(θ) + k*(k·v)*(1-cos(θ))
	cosTheta := math.Cos(angle)
	sinTheta := math.Sin(angle)

	// 计算 k·v (点积)
	kDotV := k.Dot(p)

	// 计算 k×v (叉积)
	kCrossV := k.Cross(p)

	// 应用罗德里格旋转公式
	rotated := &Vector3D{
		X: p.X*cosTheta + kCrossV.X*sinTheta + k.X*kDotV*(1-cosTheta),
		Y: p.Y*cosTheta + kCrossV.Y*sinTheta + k.Y*kDotV*(1-cosTheta),
		Z: p.Z*cosTheta + kCrossV.Z*sinTheta + k.Z*kDotV*(1-cosTheta),
	}

	// 将结果平移回原来的坐标系
	return &Vector3D{
		X: rotated.X + center.X,
		Y: rotated.Y + center.Y,
		Z: rotated.Z + center.Z,
	}
}

// RotatePointAroundAxisAtOrigin 计算点围绕通过原点的轴旋转（简化版本）
// point: 要旋转的点的世界坐标
// axis: 旋转轴向量（会被自动归一化）
// angle: 旋转角度（弧度）
func RotatePointAroundAxisAtOrigin(point *Vector3D, axis *Vector3D, angle float64) *Vector3D {
	return RotatePointAroundAxis(point, axis, angle, nil)
}
