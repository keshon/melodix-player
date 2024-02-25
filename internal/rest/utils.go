package rest

import (
	"errors"
	"math/rand"
	"os"
	"path/filepath"
)

// getImageList retrieves a list of image files from the specified folder path.
//
// folderPath string - the path of the folder to retrieve image files from.
// []string, error - a list of image file names and an error, if any.
func getImageList(folderPath string) ([]string, error) {
	var imageFiles []string
	files, err := os.ReadDir(folderPath)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		ext := filepath.Ext(file.Name())
		if ext == ".jpg" || ext == ".jpeg" || ext == ".png" {
			imageFiles = append(imageFiles, file.Name())
		}
	}
	return imageFiles, nil
}

// getRandomImage returns a random image file from the specified folder path.
//
// folderPath string - the path of the folder containing the image files.
// string, error - the randomly selected image file name, or an error if no valid images are found.
func getRandomImage(folderPath string) (string, error) {
	var imageFiles []string
	entries, err := os.ReadDir(folderPath)
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		ext := filepath.Ext(entry.Name())
		if ext == ".jpg" || ext == ".jpeg" || ext == ".png" {
			imageFiles = append(imageFiles, entry.Name())
		}
	}
	if len(imageFiles) == 0 {
		return "", errors.New("no valid images found")
	}
	randomIndex := rand.Intn(len(imageFiles))
	randomImage := imageFiles[randomIndex]
	return randomImage, nil
}
