package math2d

// Polygon represents a 2D convex polygon defined by its vertices in counter-clockwise order.
type Polygon struct {
	Vertices []Vector
}

// NewRectangle creates an axis-aligned rectangle centered at center with given width and height.
func NewRectangle(center Vector, width, height float64) Polygon {
	halfW := width / 2.0
	halfH := height / 2.0
	return Polygon{
		Vertices: []Vector{
			{X: center.X - halfW, Y: center.Y - halfH},
			{X: center.X + halfW, Y: center.Y - halfH},
			{X: center.X + halfW, Y: center.Y + halfH},
			{X: center.X - halfW, Y: center.Y + halfH},
		},
	}
}

// Translate returns a new polygon shifted by offset vector v.
func (p Polygon) Translate(v Vector) Polygon {
	newVertices := make([]Vector, len(p.Vertices))
	for i, vertex := range p.Vertices {
		newVertices[i] = vertex.Add(v)
	}
	return Polygon{Vertices: newVertices}
}

// Rotate returns a new polygon rotated around a given pivot point by angleRad.
func (p Polygon) Rotate(pivot Vector, angleRad float64) Polygon {
	newVertices := make([]Vector, len(p.Vertices))
	for i, vertex := range p.Vertices {
		diff := vertex.Sub(pivot)
		rotated := diff.Rotate(angleRad)
		newVertices[i] = pivot.Add(rotated)
	}
	return Polygon{Vertices: newVertices}
}

// Centroid computes the arithmetic mean of the polygon vertices.
func (p Polygon) Centroid() Vector {
	if len(p.Vertices) == 0 {
		return Vector{}
	}
	var sum Vector
	for _, v := range p.Vertices {
		sum = sum.Add(v)
	}
	return sum.Mul(1.0 / float64(len(p.Vertices)))
}
