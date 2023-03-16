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
	MeshID    int
	Transform mgl32.Mat4
}

type Model struct {
	name              string
	collection        *modelspec.Collection
	collectionContext *CollectionContext
	modelConfig       *ModelConfig
	renderData        []RenderData
	vertices          []modelspec.Vertex

	translation mgl32.Vec3
	rotation    mgl32.Quat
	scale       mgl32.Vec3
}

func NewModelsFromCollection(ctx *CollectionContext, modelConfig *ModelConfig) []*Model {
	var models []*Model

	for _, root := range ctx.Collection.Scenes[0].Nodes {
		nodes := []*modelspec.Node{root}
		m := &Model{
			name:              root.Name,
			collectionContext: ctx,
			collection:        ctx.Collection,
			modelConfig:       modelConfig,
		}

		for len(nodes) > 0 {
			var children []*modelspec.Node

			for _, node := range nodes {
				for _, meshID := range node.MeshIDs {
					m.renderData = append(
						m.renderData,
						// don't apply the node transform here since we set it at the model node
						// TODO - when we implement children, they will need to have the transform set i think?
						RenderData{MeshID: meshID, Transform: node.Transform},
					)
					vertices := m.collection.Meshes[meshID].UniqueVertices
					for _, v := range vertices {
						m.vertices = append(m.vertices, v)
					}
				}
				models = append(models, m)
				children = append(children, node.Children...)
			}
			nodes = children
		}

		// the root transforms are set on the model itself so that it will be
		// transformed when loaded into the scene. child nodes are considered part
		// of the same entity as the parent so the transforms are applied at render time

		// scaleMat4 := mgl32.Scale3D(root.Scale[0], root.Scale[1], root.Scale[2])
		// m.renderData[0].Transform = root.Rotation.Mat4().Mul4(scaleMat4)

		// m.renderData[0].Transform = mgl32.Ident4()
		m.renderData[0].Transform = root.Transform

		// m.translation = root.Translation
		// m.rotation = root.Rotation
		// m.scale = root.Scale

		m.rotation = mgl32.QuatIdent()
		m.scale = mgl32.Vec3{1, 1, 1}
	}

	return models
}

func (m *Model) Name() string {
	return m.name
}

func (m *Model) RootJoint() *modelspec.JointSpec {
	return m.collection.RootJoint
}

func (m *Model) Animations() map[string]*modelspec.AnimationSpec {
	return m.collection.Animations
}

func (m *Model) JointMap() map[int]*modelspec.JointSpec {
	return m.collection.JointMap
}

// func (m *Model) Vertices() []modelspec.Vertex {
// 	var vertices []modelspec.Vertex
// 	for _, mesh := range m.collection.Meshes {
// 		vertices = append(vertices, mesh.Vertices...)
// 	}
// 	return vertices
// }

func (m *Model) CollectionContext() *CollectionContext {
	return m.collectionContext
}

func (m *Model) Collection() *modelspec.Collection {
	return m.collection
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

type CollectionContext struct {
	Collection *modelspec.Collection
	VAOS       map[int]uint32
}

func CreateContext(collection *modelspec.Collection) *CollectionContext {
	c := &CollectionContext{
		Collection: collection,
		VAOS:       map[int]uint32{},
	}
	c.initialize(ModelConfig{MaxAnimationJointWeights: 4})
	return c
}

func (c *CollectionContext) initialize(modelConfig ModelConfig) {
	for i, m := range c.Collection.Meshes {
		// initialize the VAO
		var vao uint32
		gl.GenVertexArrays(1, &vao)
		gl.BindVertexArray(vao)
		c.VAOS[i] = vao

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
