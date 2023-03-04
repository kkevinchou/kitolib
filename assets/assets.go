package assets

import (
	"fmt"

	"github.com/kkevinchou/kitolib/assets/loaders"
	"github.com/kkevinchou/kitolib/font"
	"github.com/kkevinchou/kitolib/modelspec"
	"github.com/kkevinchou/kitolib/textures"
)

type AssetManager struct {
	textures       map[string]*textures.Texture
	animatedModels map[string]*modelspec.ModelSpecification
	collections    map[string]*modelspec.Collection
	fonts          map[string]font.Font
}

func NewAssetManager(directory string, loadVisualAssets bool) *AssetManager {
	var loadedTextures map[string]*textures.Texture
	var loadedFonts map[string]font.Font
	if loadVisualAssets {
		loadedTextures = loaders.LoadTextures(directory)
		loadedFonts = loaders.LoadFonts(directory)
	}

	assetManager := AssetManager{
		textures:    loadedTextures,
		collections: loaders.LoadModels(directory),
		fonts:       loadedFonts,
	}

	return &assetManager
}

func (a *AssetManager) GetTexture(name string) *textures.Texture {
	if _, ok := a.textures[name]; !ok {
		panic(fmt.Sprintf("could not find texture %s", name))
	}
	return a.textures[name]
}

func (a *AssetManager) GetModel(name string) *modelspec.ModelSpecification {
	if _, ok := a.animatedModels[name]; !ok {
		panic(fmt.Sprintf("could not find animated model %s", name))
	}
	return a.animatedModels[name]
}

func (a *AssetManager) GetCollection(name string) *modelspec.Collection {
	if _, ok := a.collections[name]; !ok {
		panic(fmt.Sprintf("could not find animated model %s", name))
	}
	return a.collections[name]
}

func (a *AssetManager) GetFont(name string) font.Font {
	if _, ok := a.fonts[name]; !ok {
		panic(fmt.Sprintf("could not find font %s", name))
	}
	return a.fonts[name]
}

func (a *AssetManager) Models() map[string]*modelspec.ModelSpecification {
	return a.animatedModels
}
