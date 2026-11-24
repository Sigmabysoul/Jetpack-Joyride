package math2d

import "math"

// Vector represents a 2D vector.
type Vector struct {
	X float64
	Y float64
}

// Add returns the vector sum of v and other.
func (v Vector) Add(other Vector) Vector {
	return Vector{X: v.X + other.X, Y: v.Y + other.Y}
}

// Sub returns the vector difference of v and other.
func (v Vector) Sub(other Vector) Vector {
	return Vector{X: v.X - other.X, Y: v.Y - other.Y}
}

// Mul returns the vector scaled by scalar s.
func (v Vector) Mul(s float64) Vector {
	return Vector{X: v.X * s, Y: v.Y * s}
}

// Dot returns the dot product of v and other.
func (v Vector) Dot(other Vector) float64 {
	return v.X*other.X + v.Y*other.Y
}

// Cross returns the 2D cross product (scalar) of v and other.
func (v Vector) Cross(other Vector) float64 {
	return v.X*other.Y - v.Y*other.X
}

// Length returns the Euclidean length of the vector.
func (v Vector) Length() float64 {
	return math.Hypot(v.X, v.Y)
}

// Unit returns the normalized unit vector. If length is 0, returns zero vector.
func (v Vector) Unit() Vector {
	l := v.Length()
	if l == 0 {
		return Vector{0, 0}
	}
	return Vector{X: v.X / l, Y: v.Y / l}
}

// Rotate returns the vector rotated by angleRad (in radians) counter-clockwise.
func (v Vector) Rotate(angleRad float64) Vector {
	cos := math.Cos(angleRad)
	sin := math.Sin(angleRad)
	return Vector{
		X: v.X*cos - v.Y*sin,
		Y: v.X*sin + v.Y*cos,
	}
}
