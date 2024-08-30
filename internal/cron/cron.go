package cron

import (
	"crypto/md5"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gookit/slog"

	"github.com/keshon/melodix-player/internal/db"
	"github.com/keshon/melodix-player/mods/music/media"
	"github.com/robfig/cron/v3"
)

type ICronTasks interface {
	Start()
}

type CronTasks struct {
}

func NewCron() ICronTasks {
	return &CronTasks{}
}

func (ct *CronTasks) Start() {

	slog.Info("Cron scheduler started")

	go func() {
		ct.runAllTasks()

		c := cron.New(cron.WithChain(cron.DelayIfStillRunning(cron.DefaultLogger)))
		c.AddFunc("@every 15m", func() {
			ct.runAllTasks()
		})

		c.Run()
	}()

}

func (ct *CronTasks) runAllTasks() {
	err := ct.dbInvalidTracks()
	if err != nil {
		slog.Error("Error processing invalid tracks: %v", err)
	}

	err = ct.dbMissingTracks()
	if err != nil {
		slog.Error("Error processing missing tracks: %v", err)
	}

	err = ct.checkAndTrimLogFile("./logs/all-levels.log")
	if err != nil {
		slog.Error("Error checking and trimming log file: %v", err)
	}
}

func (ct *CronTasks) dbInvalidTracks() error {
	tracks, err := db.GetAllTracks()
	if err != nil {
		return err
	}

	for _, track := range tracks {
		if track.Source == media.SourceYouTube.String() {
			if track.URL == "" {
				err = db.DeleteTrack(&track)
				if err != nil {
					return err
				}
			}

			if track.SongID == "" {
				ytID := strings.Split(track.URL, "v=")[1]
				track.SongID = ytID
				err = db.UpdateTrack(&track)
				if err != nil {
					return err
				}
			}
		}

		if track.Source == media.SourceLocalFile.String() {
			if track.Filepath == "" {
				err = db.DeleteTrack(&track)
				if err != nil {
					return err
				}
			}

			if track.SongID == "" {
				songID := md5.Sum([]byte(track.Title))
				songIDStr := fmt.Sprintf("%x", songID)
				track.SongID = songIDStr
				err = db.UpdateTrack(&track)
				if err != nil {
					return err
				}
			}
		}
	}

	slog.Info("Done running invalid tracks")

	return nil
}

func (ct *CronTasks) dbMissingTracks() error {
	allHistoryRecords, err := db.GetAllHistorySortedBy("")
	if err != nil {
		return err
	}

	var trackIDs []uint
	for _, record := range allHistoryRecords {
		if record.TrackID != 0 {
			trackIDs = append(trackIDs, record.TrackID)
		}
	}

	trackIDs = removeDuplicate(trackIDs)
	var deletedHistoryRecords uint
	for _, id := range trackIDs {
		doesExist, err := DoesTrackExistForHistory(id)
		if err != nil || !doesExist {
			err = db.DeleteHistory(strconv.Itoa(int(id)))
			if err != nil {
				return err
			}
			deletedHistoryRecords++
		}
	}

	slog.Info("Done running missing tracks, processed:", fmt.Sprintf("%v", deletedHistoryRecords))

	return nil
}

func removeDuplicate[T comparable](sliceList []T) []T {
	allKeys := make(map[T]bool)
	list := []T{}
	for _, item := range sliceList {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

func DoesTrackExistForHistory(trackID uint) (bool, error) {
	var count int64
	err := db.DB.Model(&db.Track{}).Where("id = ?", trackID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (c *CronTasks) checkAndTrimLogFile(filePath string) error {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	if fileInfo.Size() > 1024*1024 { // 1MB in bytes
		fileBytes, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		// Convert file contents to string
		fileContent := string(fileBytes)

		// Split file content by lines
		lines := strings.Split(fileContent, "\n")

		// Calculate the number of lines to remove (30% of total lines)
		linesToRemove := len(lines) * 30 / 100

		// Remove lines from the beginning
		trimmedLines := lines[linesToRemove:]

		// Join the remaining lines
		trimmedContent := strings.Join(trimmedLines, "\n")

		// Write back trimmed content to the file
		err = os.WriteFile(filePath, []byte(trimmedContent), 0644)
		if err != nil {
			return err
		}

		slog.Info("Trimmed log file successfully")
	}

	return nil
}
