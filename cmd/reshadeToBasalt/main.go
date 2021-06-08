package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/otiai10/copy"
	"gopkg.in/ini.v1"
)

type basaltShader struct {
	name string
	path string
}

func main() {
	args := os.Args[1:]
	if len(args) > 0 {
		var err error
		inputFile := args[0]
		// path to the reshade preset
		reshadePresetAbsolute, err := filepath.Abs(inputFile)
		if err != nil {
			log.Fatal(err)
		}
		reshadePresetPath := filepath.Dir(reshadePresetAbsolute)
		// name of the reshade preset directory
		var reshadePresetName string
		if reshadePresetPath == "." {
			reshadePresetPath, err = os.Getwd()
			reshadePresetName = strings.TrimSuffix(filepath.Base(inputFile), filepath.Ext(inputFile))
			if err != nil {
				log.Fatal(err)
			}
		} else {
			reshadePresetName = filepath.Base(reshadePresetPath)
		}
		// name of new basalt preset, reshade name + _vkBasalt
		basaltPresetPath := filepath.Dir(reshadePresetPath) + "/" + reshadePresetName + "_vkBasalt"
		basaltPresetFile := basaltPresetPath + "/" + reshadePresetName + "_vkBasalt.conf"

		// load the reshade prefix ini
		reshadePresetConfig, err := ini.Load(inputFile)
		if err != nil {
			log.Fatal(err)
		}
		basaltPresetConfig := ini.Empty()

		// make the new basalt preset directory
		os.RemoveAll(basaltPresetPath)
		os.Mkdir(basaltPresetPath, os.ModePerm)

		// clone the reshade shaders repo into rmp
		if _, existsErr := os.Stat("/tmp/reshade-shaders"); existsErr != nil {
			if os.IsNotExist(existsErr) {
				err = exec.Command("git", "clone", "--single-branch", "--branch", "master", "--depth=1", "https://github.com/crosire/reshade-shaders", "/tmp/reshade-shaders").Run()
				if err != nil {
					log.Fatal(err)
				}
			}
		}

		// copy the reshade shaders and textures to the vkbasalt preset directory
		err = copy.Copy("/tmp/reshade-shaders/Shaders", basaltPresetPath+"/Shaders")
		if err != nil {
			log.Fatal(err)
		}
		err = copy.Copy("/tmp/reshade-shaders/Textures", basaltPresetPath+"/Textures")
		if err != nil {
			log.Fatal(err)
		}

		// remove the reshade shaders repo in tmp
		os.RemoveAll("/tmp/reshade-shaders")
		// add the include and texture paths for vkbasalt
		basaltPresetConfig.Section("").Key("reshadeIncludePath").SetValue(basaltPresetPath + "/Shaders")
		basaltPresetConfig.Section("").Key("reshadeTexturePath").SetValue(basaltPresetPath + "/Textures")

		// copy the preset's shaders and textures to the vkbasalt dir
		err = filepath.WalkDir(reshadePresetPath, func(path string, di fs.DirEntry, err error) error {
			ext := strings.ToLower(filepath.Ext(path))
			name := filepath.Base(path)
			switch ext {
			case ".fx", ".fxh":
				err := copy.Copy(path, basaltPresetPath+"/Shaders/"+name)
				if err != nil {
					return err
				}
			case ".png", ".bmp", ".jpg":
				err := copy.Copy(path, basaltPresetPath+"/Textures/"+name)
				if err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			log.Fatal(err)
		}

		// get the list of used techniques from the reshade preset
		reshadePresetTechniques := strings.Split(strings.ToLower(reshadePresetConfig.Section("").Key("Techniques").String()), ",")
		var basaltPresetEffects []basaltShader

		// for each technique, if the corresponding shader exists, add it's name and path to basaltPresetEffects
		err = filepath.WalkDir(basaltPresetPath+"/Shaders/", func(path string, di fs.DirEntry, err error) error {
			if filepath.Ext(path) == ".fx" {
				shader := strings.ToLower(strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)))
				for _, technique := range reshadePresetTechniques {
					if technique == "contrastadaptivesharpen" {
						technique = "cas"
					}
					switch technique {
					case "contrastadaptivesharpen":
						technique = "cas"
					case "hdr":
						technique = "fakehdr"
					}
					if shader == technique {
						newShader := basaltShader{name: shader, path: path}
						basaltPresetEffects = append(basaltPresetEffects, newShader)
						return nil
					}
				}
			}
			return nil
		})
		if err != nil {
			log.Fatal(err)
		}

		// generate a string listing all the used effects
		var basaltEffectsString string

		for _, shader := range basaltPresetEffects {
			if len(basaltEffectsString) != 0 {
				basaltEffectsString += ":"
			}
			basaltEffectsString += shader.name
			// add a key with effect name = shader path for each used effect
			basaltPresetConfig.Section("").Key(shader.name).SetValue(shader.path)
		}

		basaltPresetConfig.Section("").Key("effects").SetValue(basaltEffectsString)

		// add the options for each reshade technique to the vkbasalt preset
		for _, section := range reshadePresetConfig.Sections() {
			sectionName := section.Name()
			if sectionName != "DEFAULT" {
				techniqueName := strings.ToLower(strings.Split(sectionName, ".")[0])
				for _, shader := range basaltPresetEffects {
					if strings.HasPrefix(techniqueName, shader.name) {
						for key, value := range section.KeysHash() {
							basaltPresetConfig.Section("").Key(shader.name + key).SetValue(value)
						}
						if shader.name == "lut" {
							basaltPresetConfig.Section("").Key(shader.name + "File").SetValue(basaltPresetPath + "/Textures/lut.png")
						}
					}
				}
			}
		}

		// finally save the output file
		err = basaltPresetConfig.SaveTo(basaltPresetFile)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Export the following environment variables before running your game to use the generated vkBasalt preset:")
		fmt.Println("export ENABLE_VKBASALT=1")
		fmt.Println("export VKBASALT_CONFIG_FILE=\"" + basaltPresetFile + "\"")

	} else {
		fmt.Println("reshadeToBasalt requires one argument: path to Reshade preset ini")
	}
}
