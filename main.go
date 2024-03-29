package main

import (
	"fmt"
	"io"
	"io/ioutil"
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

const (
	SPI_SETDESKWALLPAPER = 20
	SPIF_UPDATEINIFILE   = 0x01
	SPIF_SENDCHANGE      = 0x02
)

func main() {

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
	randomIndex := rand.Perm(len(newFiles))
	for i := 0; i < 50; i++ {
		randomFiles[i] = newFiles[randomIndex[i]]
	}
	if len(randomFiles) > 50 {
		randomFiles = randomFiles[:50]
	}

	randomFile := randomFiles[rand.Intn(len(randomFiles))]
	fmt.Println(randomFile)

	userDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	picturesDir := filepath.Join(userDir, "Pictures/Wallpaper")
	if _, err := os.Stat(picturesDir); os.IsNotExist(err) {
		if err := os.Mkdir(picturesDir, 0755); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	}

	for _, file := range randomFiles {
		srcPath := filepath.Join(wallpaperDir, file.Name())

		dstPath := filepath.Join(picturesDir, file.Name())
		if err := copyFile(srcPath, dstPath); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)

		}
	}
	wallpaperPath := syscall.StringToUTF16Ptr(filepath.Join(picturesDir, randomFile.Name()))
	result, _, _ := systemParametersInfo.Call(uintptr(SPI_SETDESKWALLPAPER), 0, uintptr(unsafe.Pointer(wallpaperPath)), uintptr(SPIF_UPDATEINIFILE|SPIF_SENDCHANGE))
	if result != 0 {
		fmt.Println("Wallpaper set")
	} else {
		fmt.Println("Ops, code:", result)
	}
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
