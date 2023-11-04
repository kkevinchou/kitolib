package assets

import (
	"fmt"
	"time"

	"github.com/kkevinchou/kitolib/assets/assetslog"
	"github.com/kkevinchou/kitolib/assets/loaders"
	"github.com/kkevinchou/kitolib/font"
	"github.com/kkevinchou/kitolib/modelspec"
	"github.com/kkevinchou/kitolib/textures"
)

type AssetManager struct {
	textures  map[string]*textures.Texture
	documents map[string]*modelspec.Document
	fonts     map[string]font.Font
}

func NewAssetManager(directory string, loadVisualAssets bool) *AssetManager {
	var loadedTextures map[string]*textures.Texture
	var loadedFonts map[string]font.Font
	var textureLoadTime time.Duration

	if loadVisualAssets {
		start := time.Now()
		loadedTextures = loaders.LoadTextures(directory)
		textureLoadTime = time.Since(start)
		loadedFonts = loaders.LoadFonts(directory)
	}

	start := time.Now()
	documents := loaders.LoadDocuments(directory)
	assetslog.Logger.Println(textureLoadTime, "to load textures")
	assetslog.Logger.Println(time.Since(start), "to load models")

	assetManager := AssetManager{
		textures:  loadedTextures,
		documents: documents,
		fonts:     loadedFonts,
	}

	return &assetManager
}

func (a *AssetManager) GetTexture(name string) *textures.Texture {
	if _, ok := a.textures[name]; !ok {
		panic(fmt.Sprintf("could not find texture %s", name))
	}
	return a.textures[name]
}

func (a *AssetManager) GetDocument(name string) *modelspec.Document {
	if _, ok := a.documents[name]; !ok {
		panic(fmt.Sprintf("could not find animated model %s", name))
	}
	return a.documents[name]
}

func (a *AssetManager) GetFont(name string) font.Font {
	if _, ok := a.fonts[name]; !ok {
		panic(fmt.Sprintf("could not find font %s", name))
	}
	return a.fonts[name]
}

func (a *AssetManager) LoadDocument(name string, filepath string) bool {
	scene := loaders.LoadDocument(name, filepath)
	if _, ok := a.documents[name]; ok {
		fmt.Printf("document with name %s already previously loaded", name)
		return false
	}

	a.documents[name] = scene
	return true
}
