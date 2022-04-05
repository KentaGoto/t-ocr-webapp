package main

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
)

func dirwalk(dir string) []string {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	var paths []string
	for _, file := range files {
		if file.IsDir() {
			paths = append(paths, dirwalk(filepath.Join(dir, file.Name()))...)
			continue
		}
		paths = append(paths, filepath.Join(dir, file.Name()))
	}

	return paths
}

func ocr(format string, img string, path string, lang string) {
	if strings.Contains(format, img) {
		fmt.Println(path)
		cmd := exec.Command("tesseract", path, path, "-l", lang)
		err := cmd.Run()
		if err != nil {
			panic(err)
		}
	} else {

	}
}

func runCommand(dir string, lang string) {
	paths := dirwalk(dir)
	fmt.Println("Processing...")
	imgs := [...]string{"jpeg", "jpg", "bmp", "png", "gif"}

	wg := &sync.WaitGroup{}

	for _, path := range paths {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()

			f, _ := os.Open(path)
			defer f.Close()

			_, format, err := image.DecodeConfig(f) // Get the image file format.
			if err != nil {
				fmt.Println(err)
			}

			for _, img := range imgs {
				ocr(format, img, path, lang)
			}
		}(path)
	}
	wg.Wait()
}

// Supported image types: jpeg, bmp, png, gif
func main() {
	var dir string
	var lang string

	if len(os.Args) == 2 {
		dir = os.Args[1]
		if dir == "-h" || dir == "--help" {
			fmt.Println(`USAGE
  $ go run main.go <Dir> <Lang Code>`)
			os.Exit(1)
		}
	}

	if len(os.Args) != 3 {
		fmt.Println("The number of arguments specified is incorrect.")
		os.Exit(1)
	} else {
		dir = os.Args[1]
		lang = os.Args[2] // Tesseract language specification options.
	}

	runCommand(dir, lang)
	fmt.Println("\nDone!")
}
