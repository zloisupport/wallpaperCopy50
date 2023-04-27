package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
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
