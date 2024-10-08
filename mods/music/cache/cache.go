package cache

import (
	"crypto/md5"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/keshon/melodix-player/internal/db"
	"github.com/keshon/melodix-player/mods/music/media"
	"github.com/keshon/melodix-player/mods/music/sources"
	"github.com/keshon/melodix-player/mods/music/utils"
)

type ICache interface {
	Curl(url string) (string, error)
	SyncCachedDir() (int, int, int, error)
	ListCachedFiles() ([]string, error)
	ListUploadedFiles() ([]string, error)
	ExtractAudioFromVideo() ([]string, error)
	syncFilesToDB(guildID string, files []os.FileInfo, cacheGuildFolder string) error
	downloadFile(filepath, url string) error
	extractAudio(path, filename string) error
	sanitizeName(filename string) string
	stripExtension(filename string) string
	humanReadableSize(size int64) string
}

// Cache struct for handling cache-related operations
type Cache struct {
	uploadsFolder string
	cacheFolder   string
	guildID       string
}

// NewCache initializes a new Cache struct
func NewCache(uploadsFolder, cacheFolder, guildID string) ICache {
	return &Cache{
		uploadsFolder: uploadsFolder,
		cacheFolder:   cacheFolder,
		guildID:       guildID,
	}
}

func (c *Cache) Curl(url string) (string, error) {
	uploadsFolder := c.uploadsFolder

	yt := sources.NewYoutube()
	song, err := yt.FetchOneByURL(url)
	if err != nil {
		return "", err
	}
	// Generate unique filename
	fileName := fmt.Sprintf("%d", time.Now().Unix())

	// Download the video
	videoFilePath := filepath.Join(uploadsFolder, fileName+".mp4")
	err = c.downloadFile(videoFilePath, song.Filepath)
	if err != nil {
		return "", fmt.Errorf("error downloading video %v", err)
	}

	// Get the file size and format information
	fileInfo, err := os.Stat(videoFilePath)
	if err != nil {
		return "", fmt.Errorf("error getting file information %v", err)
	}
	fileSize := c.humanReadableSize(fileInfo.Size())
	fileFormat := filepath.Ext(videoFilePath)

	// Check if cache folder for guild exists, create if not
	cacheGuildFolder := filepath.Join(c.cacheFolder, c.guildID)
	c.createPathIfNotExists(cacheGuildFolder)

	// Extract audio from video
	audioFilename := c.sanitizeName(song.Title) + ".mp3"
	audioFilePath := filepath.Join(cacheGuildFolder, audioFilename)
	err = c.extractAudio(videoFilePath, audioFilePath)
	if err != nil {
		return "", fmt.Errorf("error extracting audio %v", err)
	}

	// Remove the temporary video file
	err = os.Remove(videoFilePath)
	if err != nil {
		return "", fmt.Errorf("error removing temporary video file %v", err)
	}

	// Check if cached file exists in database
	existingTrack, err := db.GetTrackBySongID(song.SongID)
	if err == nil {
		existingTrack.Filepath = audioFilePath
		existingTrack.Source = media.SourceLocalFile.String()
		err := db.UpdateTrack(existingTrack)
		if err != nil {
			return "", fmt.Errorf("error updating track in database %v", err)
		}
	} else {
		newTrack := &db.Track{
			SongID:   song.SongID,
			Title:    song.Title,
			URL:      song.URL,
			Source:   media.SourceLocalFile.String(),
			Filepath: audioFilePath,
		}
		err = db.CreateTrack(newTrack)
		if err != nil {
			return "", fmt.Errorf("error creating track in database %v", err)
		}
	}

	// Get the audio file size and format
	audioFileInfo, err := os.Stat(audioFilePath)
	if err != nil {
		return "", fmt.Errorf("error getting audio file information %v", err)
	}

	audioFileSize := c.humanReadableSize(audioFileInfo.Size())

	message := fmt.Sprintf("File Size: %s\nFile Format: %s\nAudio File Size: %s", fileSize, fileFormat, audioFileSize)

	return message, nil
}

func (c *Cache) SyncCachedDir() (int, int, int, error) {
	guildID := c.guildID
	cacheFolder := c.cacheFolder

	var added, updated, removed int

	// Check if the cache folder for the guild exists
	cacheGuildFolder := filepath.Join(cacheFolder, guildID)
	_, err := os.Stat(cacheGuildFolder)
	if os.IsNotExist(err) {
		return 0, 0, 0, fmt.Errorf("cache folder for guild %s does not exist", guildID)
	}

	// Get a list of files in the cache folder
	files, err := os.ReadDir(cacheGuildFolder)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("error reading cache folder for guild %s: %v", guildID, err)
	}

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
		}

		filepath := filepath.Join(cacheGuildFolder, file.Name())
		track, err := db.GetTrackByFilepath(newPath)
		if err != nil {
			songID := md5.Sum([]byte(filepath))
			songIDStr := fmt.Sprintf("%x", songID)
			err = db.CreateTrack(&db.Track{
				SongID:   songIDStr,
				Title:    file.Name(),
				Filepath: filepath,
				Source:   media.SourceLocalFile.String(),
			})
			if err != nil {
				return 0, 0, 0, fmt.Errorf("error creating track in database %v", err)
			}
			added++
		} else {
			if track.Filepath != filepath {
				track.Filepath = filepath
				track.Source = media.SourceLocalFile.String()
				err = db.UpdateTrack(track)
				if err != nil {
					return 0, 0, 0, fmt.Errorf("error updating track in database %v", err)
				}
				updated++
			}
		}

	}

	// Iterate over database tracks and remove filepaths that no longer exist
	tracks, err := db.GetAllTracks()
	if err != nil {
		return 0, 0, 0, fmt.Errorf("error getting all tracks %v", err)
	}
	for _, track := range tracks {
		if track.Source != media.SourceLocalFile.String() {
			continue
		}

		_, err := os.Stat(track.Filepath)
		if os.IsNotExist(err) {
			slog.Info(fmt.Sprintf("%v", track))
			if track.URL == "" {
				err = db.DeleteHistory(track.SongID)
				if err != nil {
					return 0, 0, 0, fmt.Errorf("error deleting history %v", err)
				}
				db.DeleteTrack(&track)
				removed++
				continue
			} else {
				if utils.IsYouTubeURL(track.URL) {
					track.Filepath = ""
					track.Source = media.SourceYouTube.String()
					err = db.UpdateTrack(&track)
					if err != nil {
						return 0, 0, 0, fmt.Errorf("error updating track in database %v", err)
					}
					updated++
				}
			}
		}
	}

	// Iterate over database tracks and fix missing song IDs
	tracks, err = db.GetAllTracks()
	if err != nil {
		return 0, 0, 0, fmt.Errorf("error getting all tracks %v", err)
	}
	for _, track := range tracks {
		if track.SongID == "" {
			songID := md5.Sum([]byte(track.Title))
			songIDStr := fmt.Sprintf("%x", songID)
			track.SongID = songIDStr
			err = db.UpdateTrack(&track)
			if err != nil {
				return 0, 0, 0, fmt.Errorf("error updating track in database %v", err)
			}
		}
	}

	return added, updated, removed, nil
}

// ListCachedFiles lists cached files
func (c *Cache) ListCachedFiles() ([]string, error) {
	// Get the guild ID
	guildID := c.guildID
	cacheFolder := c.cacheFolder

	// Check if the cache folder for the guild exists
	cacheGuildFolder := filepath.Join(cacheFolder, guildID)
	_, err := os.Stat(cacheGuildFolder)
	if os.IsNotExist(err) {
		return []string{}, fmt.Errorf("cache folder for guild %s does not exist", guildID)
	}

	// Get a list of files in the cache folder
	files, err := os.ReadDir(cacheGuildFolder)
	if err != nil {
		return []string{}, fmt.Errorf("error reading cache folder for guild %s: %v", guildID, err)
	}

	// Initialize a buffer to store the file list
	var filelist []string

	// Iterate over the files and append their names and IDs to the buffer
	for _, file := range files {
		// Append file name and ID to the buffer
		filelist = append(filelist, file.Name())
	}

	return filelist, nil
}

func (c *Cache) ListUploadedFiles() ([]string, error) {
	// Scan uploaded folder for video files
	files, err := os.ReadDir(c.uploadsFolder)
	if err != nil {
		return []string{}, fmt.Errorf("error reading uploaded folder: %v", err)
	}

	// Send to Discord chat list of found files
	var filelist []string

	for _, file := range files {
		// Check if file is a video file
		if filepath.Ext(file.Name()) == ".mp4" || filepath.Ext(file.Name()) == ".mkv" || filepath.Ext(file.Name()) == ".webm" {
			filelist = append(filelist, file.Name())
		}
	}

	return filelist, nil
}

func (c *Cache) ExtractAudioFromVideo() ([]string, error) {
	uploadsFolder := c.uploadsFolder
	cacheFolder := c.cacheFolder
	guildID := c.guildID
	var filesStats []string

	files, err := os.ReadDir(uploadsFolder)
	if err != nil {
		return []string{}, fmt.Errorf("error reading uploaded folder: %v", err)
	}

	// Iterate each file
	for _, file := range files {

		// Check if file is a video file
		if filepath.Ext(file.Name()) == ".mp4" || filepath.Ext(file.Name()) == ".mkv" || filepath.Ext(file.Name()) == ".webm" || filepath.Ext(file.Name()) == ".flv" {

			// Check if cache folder for guild exists, create if not
			cacheGuildFolder := filepath.Join(cacheFolder, guildID)
			c.createPathIfNotExists(cacheGuildFolder)

			// Extract audio from video
			videoFilePath := filepath.Join(uploadsFolder, file.Name())
			filenameNoExt := c.stripExtension(file.Name())
			audioFilename := c.sanitizeName(filenameNoExt) + ".mp3"
			audioFilePath := filepath.Join(cacheGuildFolder, audioFilename)
			err = c.extractAudio(videoFilePath, audioFilePath)
			if err != nil {
				continue
			}

			// Remove the temporary video file
			err = os.Remove(videoFilePath)
			if err != nil {
				return []string{}, fmt.Errorf("error removing temporary video file: %v", err)
			}

			// Check if cached file exists in database
			song, err := db.GetTrackByFilepath(audioFilename)
			if err == nil {
				song.Filepath = audioFilePath
				err := db.UpdateTrack(song)
				if err != nil {
					continue
				}
			} else {
				newTrack := &db.Track{
					Title:    audioFilename,
					Source:   media.SourceLocalFile.String(),
					Filepath: audioFilePath,
				}
				err = db.CreateTrack(newTrack)
				if err != nil {
					continue
				}
			}

			// Get the audio file size and format
			audioFileInfo, err := os.Stat(audioFilePath)
			if err != nil {
				return []string{}, fmt.Errorf("error getting audio file information: %v", err)
			}

			fileSize := c.humanReadableSize(audioFileInfo.Size())

			// Send message with audio extraction information
			fileSizeAndFormat := fmt.Sprintf("%s ` %s `", audioFileInfo.Name(), fileSize)
			filesStats = append(filesStats, fileSizeAndFormat)

		}
	}

	return filesStats, nil
}

func (c *Cache) downloadFile(filepath, url string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file %v", err)
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download file %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status %s", resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("error copying file %v", err)
	}

	return nil
}

func (c *Cache) extractAudio(videoFilePath, audioFilePath string) error {
	cmd := exec.Command("ffmpeg", "-i", videoFilePath, "-vn", "-acodec", "libmp3lame", "-b:a", "256k", audioFilePath)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error extracting audio %v", err)
	}

	fmt.Printf("Audio extracted and saved to: %s\n", audioFilePath)
	return nil
}

func (c *Cache) createPathIfNotExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, 0755)
		return fmt.Errorf("error creating path %v", err)
	}
	return nil
}

func (c *Cache) sanitizeName(filename string) string {
	// Replace spaces with dots using strings.ReplaceAll
	filename = strings.ReplaceAll(filename, ":", "-")
	filename = strings.ReplaceAll(filename, " ", "-")
	filename = strings.ReplaceAll(filename, "_", "-")
	filename = strings.ReplaceAll(filename, "--", "-")
	filename = strings.ReplaceAll(filename, "---", "-")
	filename = strings.ReplaceAll(filename, ",-", "-")
	filename = strings.ReplaceAll(filename, ".-", "-")
	return filename
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

func (c *Cache) syncFilesToDB(guildID string, files []os.FileInfo, cacheGuildFolder string) error {
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
				Source:   media.SourceLocalFile.String(),
			})
		} else {
			db.UpdateTrack(&db.Track{
				Filepath: filepath,
				Source:   media.SourceLocalFile.String(),
			})
		}
	}
	return nil
}
