package assets

import (
	"fmt"

	"github.com/kkevinchou/kitolib/assets/loaders"
	"github.com/kkevinchou/kitolib/font"
	"github.com/kkevinchou/kitolib/modelspec"
	"github.com/kkevinchou/kitolib/textures"
)

type AssetManager struct {
	textures    map[string]*textures.Texture
	modelGroups map[string]*modelspec.ModelGroup
	fonts       map[string]font.Font
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
		modelGroups: loaders.LoadModelGroups(directory),
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

func (a *AssetManager) GetModelGroup(name string) *modelspec.ModelGroup {
	if _, ok := a.modelGroups[name]; !ok {
		panic(fmt.Sprintf("could not find animated model %s", name))
	}
	return a.modelGroups[name]
}

func (a *AssetManager) GetFont(name string) font.Font {
	if _, ok := a.fonts[name]; !ok {
		panic(fmt.Sprintf("could not find font %s", name))
	}
	return a.fonts[name]
}
