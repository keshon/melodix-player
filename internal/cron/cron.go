package cron

import (
	"crypto/md5"
	"fmt"
	"strconv"
	"strings"

	"github.com/gookit/slog"

	"github.com/keshon/melodix-player/internal/db"
	"github.com/keshon/melodix-player/mods/music/player"
	cron "github.com/robfig/cron/v3"
)

type ICron interface {
	Start()
	dbInvalidTracks() error
	dbMissingTracks() error
}

type Cron struct {
}

func NewCron() ICron {
	return &Cron{}
}

func (c *Cron) Start() {

	slog.Info("Cron scheduler started")

	go func() {
		c.runAllTasks()

		cr := cron.New()
		cr.AddFunc("0 30 * * * *", func() {
			c.runAllTasks()
		})

	}()

}

func (c *Cron) runAllTasks() {
	err := c.dbInvalidTracks()
	if err != nil {
		slog.Error("Error processing invalid tracks: %v", err)
	}

	err = c.dbMissingTracks()
	if err != nil {
		slog.Error("Error processing missing tracks: %v", err)
	}
}

func (c *Cron) dbInvalidTracks() error {
	tracks, err := db.GetAllTracks()
	if err != nil {
		return err
	}

	for _, track := range tracks {
		if track.Source == player.SourceYouTube.String() {
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

		if track.Source == player.SourceLocalFile.String() {
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

func (c *Cron) dbMissingTracks() error {
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
