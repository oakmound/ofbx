package threefbx

import (
	"math"

	"github.com/oakmound/oak/v4/alg/floatgeom"
)

type NurbsCurve struct {
	arcLengthDivisions int
	degree             int
	knots              []float64
	controlPoints      []floatgeom.Point4
	startKnot, endKnot int
}

func NewNurbsCurve(degree int, knots []float64, controlPoints []floatgeom.Point4, startKnot, endKnot int) *NurbsCurve {
	nc := &NurbsCurve{}
	nc.arcLengthDivisions = 200
	nc.degree = degree
	nc.knots = knots
	nc.controlPoints = controlPoints
	// Used by periodic NURBS to remove hidden spans
	nc.startKnot = startKnot
	nc.endKnot = endKnot
	return nc
}

func (nc *NurbsCurve) getPoint(t float64) floatgeom.Point3 {
	u := nc.knots[nc.startKnot] + t*(nc.knots[nc.endKnot]-nc.knots[nc.startKnot])

	// following results in (wx, wy, wz, w) homogeneous point
	var hpoint = calcBSplinePoint(nc, u)
	if hpoint.W() != 1.0 {
		// project to 3D space: (wx, wy, wz, w) -> (x, y, z, 1)
		hpoint = hpoint.DivConst(hpoint.W())
	}
	return floatgeom.Point3{hpoint.X(), hpoint.Y(), hpoint.Z()}
}

/*
Calculate B-Spline curve points. See The NURBS Book, page 82, algorithm A3.1.
p : degree of B-Spline
U : knot vector
P : control points (x, y, z, w)
u : parametric point
returns point for given u
*/
func calcBSplinePoint(nc *NurbsCurve, para float64) floatgeom.Point4 {
	span := findSpan(nc.degree, para, nc.knots)
	N := calcBasisFunctions(span, para, nc.degree, nc.knots)
	C := floatgeom.Point4{}

	for j := 0; j <= nc.degree; j++ {
		point := nc.controlPoints[span-nc.degree+j]
		Nj := N[j]
		wNj := point.W() * Nj
		C[0] += point.X() * wNj
		C[1] += point.Y() * wNj
		C[2] += point.Z() * wNj
		C[3] += point.W() * Nj
	}
	return C
}

/*
Finds knot vector span.
p : degree
u : parametric value
U : knot vector
returns the span
*/
func findSpan(degree int, para float64, knots []float64) int {

	var n = len(knots) - degree - 1
	if para >= knots[n] {
		return n - 1
	}
	if para <= knots[degree] {
		return degree
	}
	var low = degree
	var high = n
	var mid = int(math.Floor(float64((low + high) / 2)))
	for para < knots[mid] || para >= knots[mid+1] {
		if para < knots[mid] {
			high = mid
		} else {
			low = mid
		}
		mid = int(math.Floor(float64((low + high) / 2)))
	}
	return mid
}

/*
Calculate basis functions. See The NURBS Book, page 70, algorithm A2.2
span : span in which u lies
u    : parametric point
p    : degree
U    : knot vector
returns array[p+1] with basis functions values.
*/
func calcBasisFunctions(span int, para float64, degree int, knots []float64) []float64 {

	N := []float64{1.0}
	left := []float64{}
	right := []float64{}

	for j := 1; j <= degree; j++ {

		left[j] = para - knots[span+1-j]
		right[j] = knots[span+j] - para

		var saved = 0.0

		for r := 0; r < j; r++ {
			var rv = right[r+1]
			var lv = left[j-r]
			var temp = N[r] / (rv + lv)
			N[r] = saved + rv*temp
			saved = lv * temp
		}
		N[j] = saved
	}
	return N
}
