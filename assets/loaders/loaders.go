package loaders

import (
	"fmt"
	"strings"
	"sync"

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

	fileMetaData := utils.GetFileMetaData(directory, subDirectories, extensions)

	filesChan := make(chan string, len(fileMetaData))
	textureInfoChan := make(chan gltextures.TextureInfo, len(fileMetaData))

	workerCount := 10
	doneCount := 0
	var doneCountLock sync.Mutex

	for i := 0; i < workerCount; i++ {
		go func(workerIndex int) {
			for fileName := range filesChan {
				textureInfo := gltextures.ReadTextureInfo(fileName)
				textureInfoChan <- textureInfo
			}

			doneCountLock.Lock()
			doneCount += 1
			if doneCount == workerCount {
				close(textureInfoChan)
			}
			doneCountLock.Unlock()
		}(i)
	}

	for _, metaData := range fileMetaData {
		filesChan <- metaData.Path
	}
	close(filesChan)

	textureMap := map[string]*textures.Texture{}
	for textureInfo := range textureInfoChan {
		textureID := gltextures.CreateOpenGLTexture(textureInfo)
		if _, ok := textureMap[textureInfo.Name]; ok {
			panic(fmt.Sprintf("texture with duplicate name %s found", textureInfo.Name))
		}
		textureMap[textureInfo.Name] = &textures.Texture{ID: textureID}
	}

	return textureMap
}

func LoadScenes(directory string) map[string]*modelspec.Scene {
	var subDirectories []string = []string{"gltf"}

	extensions := map[string]any{
		".gltf": nil,
	}

	scenes := map[string]*modelspec.Scene{}
	fileMetaData := utils.GetFileMetaData(directory, subDirectories, extensions)

	for _, metaData := range fileMetaData {
		if strings.HasPrefix(metaData.Name, "_") {
			continue
		}

		if metaData.Extension == ".gltf" {
			// if metaData.Name != "vehicle" && metaData.Name != "demo_scene_west" {
			// 	continue
			// }
			modelGroup, err := gltf.ParseGLTF(metaData.Name, metaData.Path, &gltf.ParseConfig{TextureCoordStyle: gltf.TextureCoordStyleOpenGL})
			if err != nil {
				fmt.Println("failed to parse gltf for", metaData.Path, ", error:", err)
				continue
			}
			scenes[metaData.Name] = modelGroup
		} else {
			panic(fmt.Sprintf("wtf unexpected extension %s", metaData.Extension))
		}

	}

	return scenes
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
