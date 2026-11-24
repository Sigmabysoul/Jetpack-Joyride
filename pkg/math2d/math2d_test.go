package math2d_test

import (
	"math"
	"testing"

	"jetpack-joyride/pkg/math2d"
)

func TestVectorOperations(t *testing.T) {
	v1 := math2d.Vector{X: 3, Y: 4}
	v2 := math2d.Vector{X: 1, Y: 2}

	// Add
	add := v1.Add(v2)
	if add.X != 4 || add.Y != 6 {
		t.Errorf("expected (4, 6), got %v", add)
	}

	// Sub
	sub := v1.Sub(v2)
	if sub.X != 2 || sub.Y != 2 {
		t.Errorf("expected (2, 2), got %v", sub)
	}

	// Mul
	mul := v1.Mul(2)
	if mul.X != 6 || mul.Y != 8 {
		t.Errorf("expected (6, 8), got %v", mul)
	}

	// Dot
	dot := v1.Dot(v2)
	if dot != 11 {
		t.Errorf("expected 11, got %f", dot)
	}

	// Length
	if v1.Length() != 5 {
		t.Errorf("expected length 5, got %f", v1.Length())
	}

	// Unit
	unit := v1.Unit()
	if math.Abs(unit.Length()-1.0) > 1e-6 {
		t.Errorf("expected unit length ~1, got %f", unit.Length())
	}

	// Rotate 90 degrees
	rotated := math2d.Vector{X: 1, Y: 0}.Rotate(math.Pi / 2)
	if math.Abs(rotated.X) > 1e-6 || math.Abs(rotated.Y-1) > 1e-6 {
		t.Errorf("expected (0, 1), got %v", rotated)
	}
}

func TestPolygonAndSATCollision(t *testing.T) {
	// Two overlapping boxes
	box1 := math2d.NewRectangle(math2d.Vector{X: 10, Y: 10}, 20, 20)
	box2 := math2d.NewRectangle(math2d.Vector{X: 15, Y: 15}, 20, 20)

	if !math2d.CheckSATCollision(box1, box2) {
		t.Errorf("expected box1 and box2 to collide")
	}

	// Disjoint box
	box3 := math2d.NewRectangle(math2d.Vector{X: 100, Y: 100}, 20, 20)
	if math2d.CheckSATCollision(box1, box3) {
		t.Errorf("expected box1 and box3 to NOT collide")
	}

	// Translated box
	translated := box1.Translate(math2d.Vector{X: 90, Y: 90})
	if !math2d.CheckSATCollision(translated, box3) {
		t.Errorf("expected translated box to collide with box3")
	}

	// Rotated box
	rotated := box1.Rotate(math2d.Vector{X: 10, Y: 10}, math.Pi/4)
	if !math2d.CheckSATCollision(box1, rotated) {
		t.Errorf("expected box1 to collide with its rotated self around center")
	}

	// Rotated box corner gap (proves no AABB false positives)
	// box at (10, 10) rotated 45 deg reaches X=10+10*sqrt(2)=24.14 on horizontal axis,
	// but at corner (23, 23) it does NOT intersect the diamond
	cornerBox := math2d.NewRectangle(math2d.Vector{X: 23, Y: 23}, 2, 2)
	if math2d.CheckSATCollision(rotated, cornerBox) {
		t.Errorf("expected rotated diamond to NOT collide with box placed in outer AABB corner gap")
	}
}
