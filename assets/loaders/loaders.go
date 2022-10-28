package loaders

import (
	"fmt"
	"strings"

	"github.com/kkevinchou/kitolib/assets/loaders/glfonts"
	"github.com/kkevinchou/kitolib/assets/loaders/gltextures"
	"github.com/kkevinchou/kitolib/assets/loaders/gltf"
	"github.com/kkevinchou/kitolib/font"
	utils "github.com/kkevinchou/kitolib/libutils"
	"github.com/kkevinchou/kitolib/modelspec"
	"github.com/kkevinchou/kitolib/textures"
)

func LoadTextures(directory string) map[string]*textures.Texture {
	var subDirectories []string = []string{"images", "icons", "gltf"}

	extensions := map[string]any{
		".png": nil,
	}

	textureMap := map[string]*textures.Texture{}
	fileMetaData := utils.GetFileMetaData(directory, subDirectories, extensions)

	for _, metaData := range fileMetaData {
		textureID := gltextures.NewTexture(metaData.Path)
		if _, ok := textureMap[metaData.Name]; ok {
			panic(fmt.Sprintf("texture with duplicate name %s found", metaData.Name))
		}
		textureMap[metaData.Name] = &textures.Texture{ID: textureID}
	}

	return textureMap
}

func LoadModels(directory string) map[string]*modelspec.ModelSpecification {
	var subDirectories []string = []string{"gltf"}

	extensions := map[string]any{
		".gltf": nil,
	}

	animationMap := map[string]*modelspec.ModelSpecification{}
	fileMetaData := utils.GetFileMetaData(directory, subDirectories, extensions)

	var err error
	var modelSpec *modelspec.ModelSpecification

	for _, metaData := range fileMetaData {
		if strings.HasPrefix(metaData.Name, "_") {
			continue
		}

		if metaData.Extension == ".gltf" {
			modelSpec, err = gltf.ParseGLTF(metaData.Path, &gltf.ParseConfig{TextureCoordStyle: gltf.TextureCoordStyleOpenGL})
			if err != nil {
				fmt.Println("failed to parse gltf for", metaData.Path, ", error:", err)
				continue
			}
		} else {
			panic(fmt.Sprintf("wtf unexpected extension %s", metaData.Extension))
		}

		animationMap[metaData.Name] = modelSpec

	}

	return animationMap
}

func LoadFonts(directory string) map[string]font.Font {
	var subDirectories []string = []string{"fonts"}

	extensions := map[string]any{
		".ttf": nil,
	}

	fonts := map[string]font.Font{}
	fileMetaData := utils.GetFileMetaData(directory, subDirectories, extensions)

	for _, metaData := range fileMetaData {
		if strings.HasPrefix(metaData.Name, "_") {
			continue
		}
		fonts[metaData.Name] = glfonts.NewFont(metaData.Path, 12)
	}

	return fonts
}
