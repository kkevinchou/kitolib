package modelspec

import "github.com/go-gl/mathgl/mgl32"

type JointSpec struct {
	ID   int
	Name string

	// BindTransform is a transform that converts from the parent's space to the local space.
	// this is relative to the parent
	BindTransform mgl32.Mat4

	// InverseBindTransform is a transform that converts from the bind space of a joint to
	// world space. This is not relative to the parent. it is the inverse of FullBindTransform
	InverseBindTransform mgl32.Mat4

	// FullBindTransform is a transform that converts from world space to the joint's bind space.
	// it is the inverse of InverseBindTransform
	FullBindTransform mgl32.Mat4

	Children []*JointSpec
	Parent   *JointSpec
}
