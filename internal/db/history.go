package db

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type History struct {
	ID         uint `gorm:"primaryKey;autoIncrement"`
	GuildID    string
	TrackID    uint
	PlayCount  uint
	Duration   float64
	LastPlayed time.Time
}

func CreateHistory(history *History) error {
	history.LastPlayed = time.Now()
	return DB.Create(history).Error
}

func GetAllHistorySortedBy(sortBy string) ([]History, error) {
	var history []History
	var query *gorm.DB

	switch sortBy {
	case "duration":
		query = DB.Order("duration DESC")
	case "play_count":
		query = DB.Order("play_count DESC")
	case "last_played":
		query = DB.Order("last_played DESC")
	default:
		query = DB
	}

	if err := query.Find(&history).Error; err != nil {
		return nil, err
	}

	return history, nil
}

func GetGuildHistorySortedBy(guildID, sortBy string) ([]History, error) {
	var history []History
	var query *gorm.DB

	switch sortBy {
	case "duration":
		query = DB.Where("guild_id = ?", guildID).Order("duration DESC")
	case "play_count":
		query = DB.Where("guild_id = ?", guildID).Order("play_count DESC")
	case "last_played":
		query = DB.Where("guild_id = ?", guildID).Order("last_played DESC")
	default:
		return nil, fmt.Errorf("unsupported sort criteria: %s", sortBy)
	}

	if err := query.Find(&history).Error; err != nil {
		return nil, err
	}

	return history, nil
}

func DoesHistoryExistForGuild(trackID uint, guildID string) (bool, error) {
	var count int64
	err := DB.Model(&History{}).Where("track_id = ? AND guild_id = ?", trackID, guildID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func GetHistoryByTrackIDAndGuildID(trackID uint, guildID string) (*History, error) {
	var track History
	if err := DB.Where("track_id = ? AND guild_id = ?", trackID, guildID).First(&track).Error; err != nil {
		return nil, err
	}
	return &track, nil
}

func UpdateTrackStatsForGuild(trackID uint, guildID string, playCount uint, duration float64) error {
	return DB.Model(&History{}).
		Where("track_id = ? AND guild_id = ?", trackID, guildID).
		UpdateColumns(map[string]interface{}{
			"play_count":  playCount,
			"duration":    duration,
			"last_played": time.Now(),
		}).Error
}

func DeleteHistory(trackSongID string) error {
	return DB.Where("track_id = ?", trackSongID).Delete(&History{}).Error
}

func DoesTrackExistForHistory(trackID uint) (bool, error) {
	var count int64
	err := DB.Model(&Track{}).Where("id = ?", trackID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
