package matrix

import (
	"math"
)

func GetClosestPointOnLineSegment(segP1, segP2, point *Vector3D) *Vector3D {
	vp := point.Sub2D(segP1)
	vl := segP2.Sub2D(segP1)
	ratio := vp.Dot2D(vl) / vl.LengthSq2D()
	if ratio < 0 {
		ratio = 0
	} else if ratio > 1 {
		ratio = 1
	}
	return segP1.Add2D(vl.Mul2D(ratio))
}

func HasCollisionBetweenCapsuleAndPoint(capsuleP1, capsuleP2 *Vector3D, capsuleRadius float64, point *Vector3D) bool {
	return GetClosestPointOnLineSegment(capsuleP1, capsuleP2, point).DistanceSq2D(point) <= capsuleRadius*capsuleRadius
}

func HasCollisionBetweenCapsuleAndCircle(capsuleP1, capsuleP2 *Vector3D, capsuleRadius float64, circleCenter *Vector3D, circleRadius float64) bool {
	return GetClosestPointOnLineSegment(capsuleP1, capsuleP2, circleCenter).DistanceSq2D(circleCenter) <= (capsuleRadius+circleRadius)*(capsuleRadius+circleRadius)
}

func HasCollisionBetweenCapsuleAndCapsule(capsule1P1, capsule1P2 *Vector3D, capsule1Radius float64, capsule2P1, capsule2P2 *Vector3D, capsule2Radius float64) bool {
	return HasCollisionBetweenCapsuleAndCircle(capsule1P1, capsule1P2, capsule1Radius, capsule2P1, capsule2Radius) ||
		HasCollisionBetweenCapsuleAndCircle(capsule1P1, capsule1P2, capsule1Radius, capsule2P2, capsule2Radius)
}

// Lerp 实现线性插值
func Lerp(a, b *Vector3D, t float64) *Vector3D {
	return &Vector3D{
		X: a.X + (b.X-a.X)*t,
		Y: a.Y + (b.Y-a.Y)*t,
		Z: a.Z + (b.Z-a.Z)*t,
	}
}

// SuggestProjectileVelocity_CustomArc 计算发射速度并返回运动时间
//
//	Engine\Source\Runtime\Engine\Private\GameplayStatics.cpp 中的 UGameplayStatics::SuggestProjectileVelocity_CustomArc
func SuggestProjectileVelocity_CustomArc(startPos, endPos *Vector3D, overrideGravityZ float64, arcParam float64) (ok bool, velocity *Vector3D, moveTime float64) {
	// 确保起点和终点位置不同
	startToEnd := endPos.Sub(startPos)
	startToEndDist := startToEnd.Length()
	// TODO 尚未考虑垂直下落
	if startToEndDist > 1e-7 {
		gravityZ := overrideGravityZ
		if math.Abs(overrideGravityZ) < 1e-7 {
			// 这里假设重力为 -980 cm/s^2
			gravityZ = 980
		}
		// 根据弧参数选择弧
		startToEndDir := startToEnd.Div(startToEndDist)
		launchDir := Lerp(&Vector3D{X: 0, Y: 0, Z: 1}, startToEndDir, arcParam).Norm()
		/*
			y方向和z方向的速度比为:
			a = launchDir.Z / sqrt(launchDir.X^2+launcherDir.Y^2)
			a是有符号的

			1: 水平方向 X = vx * moveTime
			2: 垂直方向 Y = vy  * moveTime - 0.5 * g * moveTime^2
			3: vy/vx = a

			由3得出:
			4: vy = a * vx

			vy * moveTime = a * vx * moveTime = a * X

			2 -> Y = a * X - 0.5 * g * moveTime^2

			求 moveTime
			moveTime = sqrt( 2 * ( a * X - Y ) / g )
		*/
		///////////////////////////////////////////////////////////////////////////////////////////////

		a := launchDir.Z / math.Sqrt(launchDir.X*launchDir.X+launchDir.Y*launchDir.Y)
		X := math.Sqrt(startToEnd.X*startToEnd.X + startToEnd.Y*startToEnd.Y)
		Y := startToEnd.Z
		if X*a-Y > 0 {
			moveTime = math.Sqrt(2 * (X*a - Y) / gravityZ)
			vx := X / moveTime
			vy := a * vx
			V := math.Sqrt(vx*vx + vy*vy)
			dirV := launchDir.Mul(V)
			ok = true
			velocity = dirV
			return
		} else {
			return
		}
	}
	return
}

// 获取与给定向量正交的两个单位向量
func GetOrthogonalVectors(normal *Vector3D) (u, v *Vector3D) {
	// 寻找一个不平行于法向量的任意向量
	aux := &Vector3D{X: 1, Y: 0, Z: 0}
	if math.Abs(normal.Dot(aux)) > 0.9 { // 如果法向量接近X轴，改用Y轴
		aux = &Vector3D{X: 0, Y: 1, Z: 0}
	}

	// 计算第一个正交向量 u = aux × normal
	u = aux.Cross(normal).Norm()

	// 计算第二个正交向量 v = normal × u
	v = normal.Cross(u).Norm()

	return
}

// 获取垂直于向量AB的圆上点（圆心在B）
func GetPerpendicularCirclePoints(a, b *Vector3D, r float64, numPoints int) []*Vector3D {
	// 1. 计算向量AB (从A指向B)
	ab := Vector3D{
		X: b.X - a.X,
		Y: b.Y - a.Y,
		Z: b.Z - a.Z,
	}

	// 2. 归一化AB向量（单位向量）
	abNorm := ab.Norm()

	// 3. 构建垂直平面上的两个正交基向量
	u, v := GetOrthogonalVectors(abNorm)

	// 4. 生成圆上点
	points := make([]*Vector3D, 0)
	angle := 360 / float64(numPoints)
	theta := float64(0)
	for i := 0; i < numPoints; i++ {
		// 随机角度（可选：均匀分布时用 math.Pi*2*float64(i)/float64(numPoints))
		// 参数方程：圆心 + r*(cosθ*u + sinθ*v)
		theta += angle
		cos := math.Cos(theta)
		sin := math.Sin(theta)

		points = append(points, &Vector3D{
			X: b.X + r*(cos*u.X+sin*v.X),
			Y: b.Y + r*(cos*u.Y+sin*v.Y),
			Z: b.Z + r*(cos*u.Z+sin*v.Z),
		})

	}

	return points
}

// 2D胶囊体线框图的轮廓点
// center是胶囊体框的中心点坐标
// forward 朝向。垂直于胶囊体框的平面。
// up 胶囊体的上方向。
func GetCapsulePoints(center *Vector3D, forward *Vector3D, up *Vector3D, r float64, h float64, hasArch bool, numPoints int) []*Vector3D {
	if up == nil {
		upOne := &Vector3D{0, 0, 1}
		right := upOne.Cross(forward)
		up = forward.Cross(right).Norm()
	}
	// 4. 计算胶囊体两端的圆心位置
	center1 := center.Add(up.Mul(h / 2))
	center2 := center.Add(up.Mul(-h / 2))
	// 6. 生成胶囊体轮廓点
	points := make([]*Vector3D, 0)

	// 计算每个部分的点数分配

	sidePoints := numPoints / 2 // 每条边的点数
	if sidePoints < 3 {
		sidePoints = 3
	}
	// 计算胶囊体的右向量
	rightNorm := up.Cross(forward).Norm()
	arcStartPoint1 := center1.Add(rightNorm.Mul(r))  // 第一个半圆的起始点
	arcEndPoint1 := center1.Add(rightNorm.Mul(-r))   // 第一个半圆的结束点
	artStartPoint2 := center2.Add(rightNorm.Mul(-r)) // 第二个半圆的起点
	arcEndPoint2 := center2.Add(rightNorm.Mul(r))    // 第二个半圆的结束点
	if hasArch {
		points = append(points, arcStartPoint1)
		// 生成第一个半圆 (center1为圆心，从0到π)
		for i := 0; i < sidePoints; i++ {
			theta := math.Pi * float64(i) / float64(sidePoints-1)
			pp := RotatePointAroundAxis(arcStartPoint1, forward, theta, center1)
			points = append(points, pp)
		}
		points = append(points, artStartPoint2)
		// 生成第二个半圆 (center2为圆心，从π到2π)
		for i := 0; i < sidePoints; i++ {
			theta := math.Pi * float64(i) / float64(sidePoints-1)
			pp := RotatePointAroundAxis(artStartPoint2, forward, theta, center2)
			points = append(points, pp)
		}
		points = append(points, arcEndPoint2)
		points = append(points, arcStartPoint1) // 最后一个点连接到起点，封闭成胶囊体
	} else {
		points = append(points, arcStartPoint1) // 最后一个点连接到起点，封闭成胶囊体
		points = append(points, arcEndPoint1)
		points = append(points, artStartPoint2)
		points = append(points, arcEndPoint2)
		points = append(points, arcStartPoint1)
	}
	return points
}
