package model

type IAssetManager interface {
	GetTexture(textureName string) uint32
}
