package math2d

// CheckSATCollision uses the Separating Axis Theorem (SAT) to test whether two convex polygons overlap.
func CheckSATCollision(p1, p2 Polygon) bool {
	if len(p1.Vertices) < 3 || len(p2.Vertices) < 3 {
		return false
	}

	// Check axes perpendicular to edges of p1
	if !testPolygonAxes(p1, p2) {
		return false
	}

	// Check axes perpendicular to edges of p2
	if !testPolygonAxes(p2, p1) {
		return false
	}

	return true
}

func testPolygonAxes(p1, p2 Polygon) bool {
	n := len(p1.Vertices)
	for i := 0; i < n; i++ {
		pCurrent := p1.Vertices[i]
		pNext := p1.Vertices[(i+1)%n]

		// Edge vector
		edge := pNext.Sub(pCurrent)

		// Perpendicular normal axis (-dy, dx)
		axis := Vector{X: -edge.Y, Y: edge.X}.Unit()
		if axis.Length() == 0 {
			continue
		}

		// Project p1 and p2 onto axis
		minA, maxA := projectPolygon(p1, axis)
		minB, maxB := projectPolygon(p2, axis)

		// Check for separating gap
		if maxA < minB || maxB < minA {
			return false
		}
	}
	return true
}

func projectPolygon(p Polygon, axis Vector) (float64, float64) {
	min := p.Vertices[0].Dot(axis)
	max := min
	for i := 1; i < len(p.Vertices); i++ {
		proj := p.Vertices[i].Dot(axis)
		if proj < min {
			min = proj
		}
		if proj > max {
			max = proj
		}
	}
	return min, max
}
