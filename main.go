package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

var (
	user32               = syscall.NewLazyDLL("user32.dll")
	systemParametersInfo = user32.NewProc("SystemParametersInfoW")
)

type Wallpapers struct {
	name []string
}

const (
	SPI_SETDESKWALLPAPER = 20
	SPIF_UPDATEINIFILE   = 0x01
	SPIF_SENDCHANGE      = 0x02
)

func main() {
	setWallpaper()
	copyRandomWallpaper()
}

func copyRandomWallpaper() {

	rand.Seed(time.Now().UnixNano())
	wallpaperDir := "./wallpaper"

	files, err := ioutil.ReadDir(wallpaperDir)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	var newFiles []os.FileInfo

	randomFiles := make([]os.FileInfo, len(files))

	for _, file := range files {
		if strings.HasSuffix(file.Name(), "jpg") {
			newFiles = append(newFiles, file)
		}
	}

	if len(newFiles) >= 50 {
		randomIndex := rand.Perm(len(newFiles))
		for i := 0; i < 50; i++ {
			randomFiles[i] = newFiles[randomIndex[i]]
		}
		if len(randomFiles) > 50 {
			randomFiles = randomFiles[:50]
		}

		randomFile := randomFiles[rand.Intn(len(randomFiles))]
		fmt.Println(randomFile)

		picturesDir := pictureFolder()

		for _, file := range randomFiles {
			srcPath := filepath.Join(wallpaperDir, file.Name())

			dstPath := filepath.Join(picturesDir, file.Name())
			if err := copyFile(srcPath, dstPath); err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}
		}
		//setDefaultWallpaper(picturesDir, randomFile)
	}

}
func pictureFolder() string {
	userDir := strings.ToLower(os.Getenv("windir"))
	userDir = strings.Replace(userDir, "windows", "", 1)

	picturesDir := filepath.Join(userDir, "Pictures")
	if _, err := os.Stat(picturesDir); os.IsNotExist(err) {
		if err := os.Mkdir(picturesDir, 0755); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	}

	wallpaperDir := filepath.Join(picturesDir, "Wallpaper")
	if _, err := os.Stat(wallpaperDir); os.IsNotExist(err) {
		if err := os.Mkdir(wallpaperDir, 0755); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	}

	return wallpaperDir
}

func copyDefaultWallpapers(files []string) {
	pictureFolder := pictureFolder()
	wallpaperPath, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	wallpaperPath = filepath.Join(wallpaperPath, "wallpaper")
	for _, file := range files {
		srcPath := filepath.Join(wallpaperPath, file)
		dstPath := filepath.Join(pictureFolder, file)
		if !checkFileS(dstPath) {
			if err := copyFile(srcPath, dstPath); err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}
		}
	}
}

func setWallpaper() {
	file := openDefaultJson()
	copyDefaultWallpapers(file)
	wall := selectWallpaper(file)
	setDefaultWallpaper(wall)
}

func checkFile(file os.FileInfo) bool {
	if _, err := os.Stat(file.Name()); err == nil {
		return true
	}
	return false
}

func checkFileS(file string) bool {
	if _, err := os.Stat(file); err == nil {
		return true
	}
	return false
}

func setDefaultWallpaper(randomFile os.FileInfo) {
	picturesDir := pictureFolder()
	if len(picturesDir) != 0 {
		dir := pictureFolder()

		wallpaperPath := syscall.StringToUTF16Ptr(filepath.Join(dir, randomFile.Name()))
		result, _, _ := systemParametersInfo.Call(uintptr(SPI_SETDESKWALLPAPER), 0, uintptr(unsafe.Pointer(wallpaperPath)), uintptr(SPIF_UPDATEINIFILE|SPIF_SENDCHANGE))
		if result != 0 {
			fmt.Println("Wallpaper set")
		} else {
			fmt.Println("Ops, code:", result)
		}
	}
}

func selectWallpaper(files []string) os.FileInfo {
	wallpapers := Wallpapers{name: files}
	randomFile := rand.Intn(len(wallpapers.name))
	wallpapersPath := pictureFolder()

	defaultWallpaperPath := filepath.Join(wallpapersPath, wallpapers.name[randomFile])

	fileInfo, err := os.Stat(defaultWallpaperPath)
	if err != nil {
		log.Fatal(err)
	}
	return fileInfo
}

func openDefaultJson() []string {
	if !checkFileS("wallpaper.json") {
		createDefaultJson()
	}
	f, err := os.Open("wallpaper.json")
	if err != nil {
		log.Fatal(err)
	}

	var wallpapers []string
	err = json.NewDecoder(f).Decode(&wallpapers)
	if err != nil {
		log.Fatal(err)
	}
	return wallpapers
}

func createDefaultJson() {
	wallpaper := Wallpapers{name: []string{"1.jpg", "2.png", "3.jpg"}}
	fmt.Println(wallpaper)
	f, err := os.OpenFile("wallpaper.json", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	jsonBytes, err := json.MarshalIndent(wallpaper.name, "", "   ")
	if err != nil {
		fmt.Println("Error Marshal Parsing!:", err)
		return
	}

	_, err = f.Write(jsonBytes)
	if err != nil {
		fmt.Println("Error Writing to File!:", err)
		return
	}
	fmt.Println("Data successfully written to wallpaper.json")
}

func copyFile(src string, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err = srcFile.Seek(0, 0); err != nil {
		return err
	}

	if _, err = io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	if err = dstFile.Sync(); err != nil {
		return err
	}

	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, srcInfo.Mode())
}
