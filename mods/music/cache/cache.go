package cache

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/keshon/melodix-player/internal/db"
	"github.com/keshon/melodix-player/mods/music/player"
)

// WIP

type ICache interface {
	Curl(url string) (string, error)
	syncFilesToDB(guildID string) error
	listFiles(guildID, folder string) (string, error)
	downloadFile(filepath, url string) (string, error)
	extractAudio(path, filename string) (string, error)
	sanitizeName(filename string)
	stripExtension(filename string) string
	generateTempFilename() string
	humanReadableSize(size int64) string
	syncFilesToDatabase(guildID string, files []os.FileInfo, cacheGuildFolder string) error
}

// Cache struct for handling cache-related operations
type Cache struct{}

// NewCache initializes a new Cache struct
func NewCache() *Cache {
	return &Cache{}
}

func (c *Cache) Curl(url string) (string, error) {
	// Your implementation here
	return "", nil
}

// SyncFiles syncs cached files to the database
func (c *Cache) syncFilesToDB(guildID string) error {
	// Your implementation here

	return nil
}

// ListFiles lists cached files
func (c *Cache) listFiles(guildID, folder string) (string, error) {
	// Your implementation here

	return "", nil
}

func (c *Cache) downloadFile(filepath, url string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	return err
}

func (c *Cache) extractAudio(videoFilePath, audioFilePath string) error {
	cmd := exec.Command("ffmpeg", "-i", videoFilePath, "-vn", "-acodec", "libmp3lame", "-b:a", "256k", audioFilePath)
	err := cmd.Run()
	if err != nil {
		return err
	}

	fmt.Printf("Audio extracted and saved to: %s\n", audioFilePath)
	return nil
}

func (c *Cache) createPathIfNotExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, 0755)
		return err
	}
	return nil
}

func (c *Cache) sanitizeName(filename string) string {
	// Replace spaces with dots using strings.ReplaceAll
	newFilename := strings.ReplaceAll(filename, " ", "_")
	return newFilename
}

func (c *Cache) stripExtension(filename string) string {
	basename := filepath.Base(filename)
	if strings.HasPrefix(basename, ".") {
		return basename
	}
	extension := filepath.Ext(basename)
	if extension == "" {
		return basename
	}
	nameWithoutExtension := strings.TrimSuffix(basename, extension)
	return nameWithoutExtension
}

func (c *Cache) generateTempFilename() string {
	return fmt.Sprintf("%d", time.Now().Unix())
}

func (c *Cache) humanReadableSize(size int64) string {
	const (
		b = 1 << (10 * iota)
		kb
		mb
		gb
		tb
		pb
	)
	if size < kb {
		return fmt.Sprintf("%d B", size)
	}
	if size < mb {
		return fmt.Sprintf("%.2f KB", float64(size)/kb)
	}
	if size < gb {
		return fmt.Sprintf("%.2f MB", float64(size)/mb)
	}
	if size < tb {
		return fmt.Sprintf("%.2f GB", float64(size)/gb)
	}
	if size < pb {
		return fmt.Sprintf("%.2f TB", float64(size)/tb)
	}
	return fmt.Sprintf("%.2f PB", float64(size)/pb)
}

func (c *Cache) SyncFilesToDatabase(guildID string, files []os.FileInfo, cacheGuildFolder string) error {
	// Iterate over the files and append their names and IDs to the buffer
	for _, file := range files {
		filenameNoExt := c.stripExtension(file.Name())
		audioFilename := c.sanitizeName(filenameNoExt) + filepath.Ext(file.Name())

		// Rename the file to formatted name
		oldPath := filepath.Join(cacheGuildFolder, file.Name())
		newPath := filepath.Join(cacheGuildFolder, audioFilename)

		// Rename the file to formatted name
		err := os.Rename(oldPath, newPath)
		if err != nil {
			// Handle error if renaming fails
			fmt.Printf("Error renaming file %s to %s: %v\n", oldPath, newPath, err)
			continue
		}

		filepath := filepath.Join(cacheGuildFolder, file.Name())
		_, err = db.GetTrackByFilepath(newPath)
		if err != nil {
			db.CreateTrack(&db.Track{
				Title:    file.Name(),
				Filepath: filepath,
				Source:   player.SourceLocalFile.String(),
			})
		} else {
			db.UpdateTrack(&db.Track{
				Filepath: filepath,
				Source:   player.SourceLocalFile.String(),
			})
		}
	}
	return nil
}
