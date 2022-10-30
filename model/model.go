package model

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/kkevinchou/kitolib/assets"
	"github.com/kkevinchou/kitolib/modelspec"
)

type IAssetManager interface {
	GetTexture(textureName string) uint32
}

type ModelConfig struct {
	MaxAnimationJointWeights int
}

type Model struct {
	meshes      []*Mesh
	modelSpec   *modelspec.ModelSpecification
	modelConfig *ModelConfig
}

func NewModel(spec *modelspec.ModelSpecification, modelConfig *ModelConfig) *Model {
	var meshes []*Mesh
	for _, ms := range spec.Meshes {
		meshes = append(meshes, NewMesh(ms, modelConfig))
	}

	m := &Model{
		modelSpec:   spec,
		meshes:      meshes,
		modelConfig: modelConfig,
	}

	return m
}

func (m *Model) InitializeRenderingProperties(assetManager assets.AssetManager) {
	for _, ms := range m.meshes {
		ms.InitializeRenderingProperties(m.modelConfig, assetManager)
	}
}

func (m *Model) RootJoint() *modelspec.JointSpec {
	return m.modelSpec.RootJoint
}

func (m *Model) Animations() map[string]*modelspec.AnimationSpec {
	return m.modelSpec.Animations
}

func (m *Model) Meshes() []*Mesh {
	return m.meshes
}

func (m *Model) Vertices() []modelspec.Vertex {
	var vertices []modelspec.Vertex
	for _, mesh := range m.meshes {
		meshVerts := mesh.Vertices()
		vertices = append(vertices, meshVerts...)
	}
	return vertices
}

func (m *Model) MeshChunks() []*MeshChunk {
	var meshChunks []*MeshChunk
	for _, mesh := range m.Meshes() {
		meshChunks = append(meshChunks, mesh.MeshChunks()...)
	}
	return meshChunks
}

func (m *Model) RootTransforms() mgl32.Mat4 {
	return m.modelSpec.RootTransforms
}
