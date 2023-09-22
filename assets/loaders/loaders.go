package loaders

import (
	"fmt"
	"strings"

	"github.com/kkevinchou/kitolib/assets/loaders/glfonts"
	"github.com/kkevinchou/kitolib/assets/loaders/gltextures"
	"github.com/kkevinchou/kitolib/assets/loaders/gltf"
	"github.com/kkevinchou/kitolib/font"
	"github.com/kkevinchou/kitolib/modelspec"
	"github.com/kkevinchou/kitolib/textures"
	"github.com/kkevinchou/kitolib/utils"
)

func LoadTextures(directory string) map[string]*textures.Texture {
	var subDirectories []string = []string{"images", "icons", "gltf"}

	extensions := map[string]any{
		".png":  nil,
		".jpeg": nil,
		".jpg":  nil,
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

func LoadModelGroups(directory string) map[string]*modelspec.ModelGroup {
	var subDirectories []string = []string{"gltf"}

	extensions := map[string]any{
		".gltf": nil,
	}

	modelGroups := map[string]*modelspec.ModelGroup{}
	fileMetaData := utils.GetFileMetaData(directory, subDirectories, extensions)

	for _, metaData := range fileMetaData {
		if strings.HasPrefix(metaData.Name, "_") {
			continue
		}

		if metaData.Extension == ".gltf" {
			// if metaData.Name != "vehicle" && metaData.Name != "demo_scene_west" {
			// 	continue
			// }
			modelGroup, err := gltf.ParseGLTF(metaData.Path, &gltf.ParseConfig{TextureCoordStyle: gltf.TextureCoordStyleOpenGL})
			if err != nil {
				fmt.Println("failed to parse gltf for", metaData.Path, ", error:", err)
				continue
			}
			modelGroups[metaData.Name] = modelGroup
		} else {
			panic(fmt.Sprintf("wtf unexpected extension %s", metaData.Extension))
		}

	}

	return modelGroups
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
