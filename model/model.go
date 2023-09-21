package model

import (
	"sort"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/kkevinchou/kitolib/modelspec"
)

type IAssetManager interface {
	GetTexture(textureName string) uint32
}

type ModelConfig struct {
	MaxAnimationJointWeights int
}

type RenderData struct {
	Name      string
	MeshID    int
	Mesh      *modelspec.MeshSpecification
	Transform mgl32.Mat4
	VAO       uint32
}

type Model struct {
	name        string
	modelGroup  *modelspec.ModelGroup
	modelConfig *ModelConfig
	renderData  []RenderData
	vertices    []modelspec.Vertex

	translation mgl32.Vec3
	rotation    mgl32.Quat
	scale       mgl32.Vec3
}

func parseRenderData(node *modelspec.Node, parentTransform mgl32.Mat4, ignoreTransform bool, vaos map[int]uint32, meshes []*modelspec.MeshSpecification) []RenderData {
	var data []RenderData

	transform := node.Transform
	if ignoreTransform {
		transform = mgl32.Ident4()
	}
	transform = parentTransform.Mul4(transform)

	for _, meshID := range node.MeshIDs {
		data = append(
			data, RenderData{
				Name:      node.Name,
				MeshID:    meshID,
				Mesh:      meshes[meshID],
				Transform: transform,
				VAO:       vaos[meshID],
			},
		)
	}

	for _, childNode := range node.Children {
		data = append(data, parseRenderData(childNode, transform, false, vaos, meshes)...)
	}

	return data
}

func NewModelsFromCollection(modelGroup *modelspec.ModelGroup, modelConfig *ModelConfig) []*Model {
	var models []*Model
	vaos := createVAOs(modelConfig, modelGroup.Meshes)

	for _, root := range modelGroup.Scenes[0].Nodes {
		m := &Model{
			name:        root.Name,
			modelGroup:  modelGroup,
			modelConfig: modelConfig,

			// ignores the transform from the root, this is applied to the model directly
			renderData: parseRenderData(root, mgl32.Ident4(), true, vaos, modelGroup.Meshes),
		}

		for _, renderData := range m.renderData {
			meshID := renderData.MeshID
			vertices := m.modelGroup.Meshes[meshID].UniqueVertices
			for _, v := range vertices {
				m.vertices = append(m.vertices, v)
			}
		}

		models = append(models, m)

		// apply transformations directly
		m.translation = root.Translation
		m.rotation = root.Rotation
		m.scale = root.Scale
	}

	return models
}

func (m *Model) Name() string {
	return m.name
}

func (m *Model) RootJoint() *modelspec.JointSpec {
	return m.modelGroup.RootJoint
}

func (m *Model) Animations() map[string]*modelspec.AnimationSpec {
	return m.modelGroup.Animations
}

func (m *Model) JointMap() map[int]*modelspec.JointSpec {
	return m.modelGroup.JointMap
}

func (m *Model) RenderData() []RenderData {
	return m.renderData
}

func (m *Model) Vertices() []modelspec.Vertex {
	return m.vertices
}

func (m *Model) Translation() mgl32.Vec3 {
	return m.translation
}

func (m *Model) Rotation() mgl32.Quat {
	return m.rotation
}

func (m *Model) Scale() mgl32.Vec3 {
	return m.scale
}

type ModelGroupContext struct {
	ModelGroup *modelspec.ModelGroup
	VAOS       map[int]uint32
}

func createVAOs(modelConfig *ModelConfig, meshes []*modelspec.MeshSpecification) map[int]uint32 {
	vaos := map[int]uint32{}
	for i, m := range meshes {
		// initialize the VAO
		var vao uint32
		gl.GenVertexArrays(1, &vao)
		gl.BindVertexArray(vao)
		vaos[i] = vao

		var vertexAttributes []float32
		var jointIDsAttribute []int32
		var jointWeightsAttribute []float32

		// set up the source data for the VBOs
		for _, vertex := range m.UniqueVertices {
			position := vertex.Position
			normal := vertex.Normal
			texture0Coords := vertex.Texture0Coords
			texture1Coords := vertex.Texture1Coords
			jointIDs := vertex.JointIDs
			jointWeights := vertex.JointWeights

			vertexAttributes = append(vertexAttributes,
				position.X(), position.Y(), position.Z(),
				normal.X(), normal.Y(), normal.Z(),
				texture0Coords.X(), texture0Coords.Y(),
				texture1Coords.X(), texture1Coords.Y(),
			)

			ids, weights := fillWeights(jointIDs, jointWeights, modelConfig.MaxAnimationJointWeights)
			for _, id := range ids {
				jointIDsAttribute = append(jointIDsAttribute, int32(id))
			}
			jointWeightsAttribute = append(jointWeightsAttribute, weights...)
		}

		totalAttributeSize := len(vertexAttributes) / len(m.UniqueVertices)

		// lay out the position, normal, texture (index 0 and 1) coords in a VBO
		var vbo uint32
		gl.GenBuffers(1, &vbo)
		gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
		gl.BufferData(gl.ARRAY_BUFFER, len(vertexAttributes)*4, gl.Ptr(vertexAttributes), gl.STATIC_DRAW)

		ptrOffset := 0
		floatSize := 4

		// position
		gl.VertexAttribPointer(0, 3, gl.FLOAT, false, int32(totalAttributeSize)*4, nil)
		gl.EnableVertexAttribArray(0)

		ptrOffset += 3

		// normal
		gl.VertexAttribPointer(1, 3, gl.FLOAT, false, int32(totalAttributeSize)*4, gl.PtrOffset(ptrOffset*floatSize))
		gl.EnableVertexAttribArray(1)

		ptrOffset += 3

		// texture coords 0
		gl.VertexAttribPointer(2, 2, gl.FLOAT, false, int32(totalAttributeSize)*4, gl.PtrOffset(ptrOffset*floatSize))
		gl.EnableVertexAttribArray(2)

		ptrOffset += 2

		// texture coords 1
		gl.VertexAttribPointer(3, 2, gl.FLOAT, false, int32(totalAttributeSize)*4, gl.PtrOffset(ptrOffset*floatSize))
		gl.EnableVertexAttribArray(3)

		// lay out the joint IDs in a VBO
		var vboJointIDs uint32
		gl.GenBuffers(1, &vboJointIDs)
		gl.BindBuffer(gl.ARRAY_BUFFER, vboJointIDs)
		gl.BufferData(gl.ARRAY_BUFFER, len(jointIDsAttribute)*4, gl.Ptr(jointIDsAttribute), gl.STATIC_DRAW)
		gl.VertexAttribIPointer(4, int32(modelConfig.MaxAnimationJointWeights), gl.INT, int32(modelConfig.MaxAnimationJointWeights)*4, nil)
		gl.EnableVertexAttribArray(4)

		// lay out the joint weights in a VBO
		var vboJointWeights uint32
		gl.GenBuffers(1, &vboJointWeights)
		gl.BindBuffer(gl.ARRAY_BUFFER, vboJointWeights)
		gl.BufferData(gl.ARRAY_BUFFER, len(jointWeightsAttribute)*4, gl.Ptr(jointWeightsAttribute), gl.STATIC_DRAW)
		gl.VertexAttribPointer(5, int32(modelConfig.MaxAnimationJointWeights), gl.FLOAT, false, int32(modelConfig.MaxAnimationJointWeights)*4, nil)
		gl.EnableVertexAttribArray(5)

		// set up the EBO, each triplet of indices point to three vertices
		// that form a triangle.
		var ebo uint32
		gl.GenBuffers(1, &ebo)
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
		gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(m.VertexIndices)*4, gl.Ptr(m.VertexIndices), gl.STATIC_DRAW)
	}

	return vaos
}

func fillWeights(jointIDs []int, weights []float32, maxAnimationJointWeights int) ([]int, []float32) {
	j := []int{}
	w := []float32{}

	if len(jointIDs) <= maxAnimationJointWeights {
		j = append(j, jointIDs...)
		w = append(w, weights...)
		// fill in empty jointIDs and weights
		for i := 0; i < maxAnimationJointWeights-len(jointIDs); i++ {
			j = append(j, 0)
			w = append(w, 0)
		}
	} else if len(jointIDs) > maxAnimationJointWeights {
		jointWeights := []JointWeight{}
		for i := range jointIDs {
			jointWeights = append(jointWeights, JointWeight{JointID: jointIDs[i], Weight: weights[i]})
		}
		sort.Sort(sort.Reverse(byWeights(jointWeights)))

		// take top 3 weights
		jointWeights = jointWeights[:maxAnimationJointWeights]
		NormalizeWeights(jointWeights)
		for _, jw := range jointWeights {
			j = append(j, jw.JointID)
			w = append(w, jw.Weight)
		}
	}

	return j, w
}
