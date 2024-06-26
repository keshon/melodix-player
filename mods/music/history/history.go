package history

import (
	"time"

	"github.com/keshon/melodix-player/internal/db"
)

type Song struct {
	Title     string        // Title of the song
	URL       string        // URL provided by the user
	Filepath  string        // Path/URL for downloading the song
	Thumbnail Thumbnail     // Thumbnail image for the song
	Duration  time.Duration // Duration of the song
	SongID    string        // Unique ID for the song
	Source    string        // Source type of the song
}

type Thumbnail struct {
	URL    string
	Width  uint
	Height uint
}

type History struct{}

type HistoryTrackInfo struct {
	History db.History
	Track   db.Track
}

type IHistory interface {
	AddTrackToHistory(guildID string, song *Song) error
	AddPlaybackAllStats(guildID, songID string, duration float64) error
	AddPlaybackCountStats(guildID, songID string) error
	AddPlaybackDurationStats(guildID, songID string, duration float64) error
	GetHistory(guildID string, sortBy string) ([]HistoryTrackInfo, error)
	GetTrackFromHistory(guildID string, trackID uint) (db.Track, error)
}

func NewHistory() IHistory {
	return &History{}
}

func (h *History) AddTrackToHistory(guildID string, song *Song) error {
	var track *db.Track

	existingTrack, err := db.GetTrackBySongID(song.SongID)
	if err != nil {
		newTrack := &db.Track{
			SongID:   song.SongID,
			Title:    song.Title,
			URL:      song.URL,
			Source:   song.Source,
			Filepath: song.Filepath,
		}

		if err := db.CreateTrack(newTrack); err != nil {
			return err
		}

		existingTrack, _ = db.GetTrackBySongID(song.SongID)
	}

	if existingTrack == nil {
		newTrack := &db.Track{
			SongID:   song.SongID,
			Title:    song.Title,
			URL:      song.URL,
			Source:   song.Source,
			Filepath: song.Filepath,
		}
		track = newTrack
	} else {
		track = existingTrack
	}

	exists, err := db.DoesHistoryExistForGuild(track.ID, guildID)
	if err != nil {
		return err
	}

	if !exists {
		history := db.History{
			GuildID: guildID,
			TrackID: track.ID,
		}
		return db.CreateHistory(&history)
	}

	return nil
}

// AddPlaybackStats updates all playback statistics (duration and count) for a track.
func (h *History) AddPlaybackAllStats(guildID, songID string, duration float64) error {

	existingTrackRecord, err := db.GetTrackBySongID(songID)
	if err != nil {
		return err
	}

	existingHistoryRecord, err := db.GetHistoryByTrackIDAndGuildID(existingTrackRecord.ID, guildID)
	if err != nil {
		return err
	}

	newPlayCount := existingHistoryRecord.PlayCount + 1
	newDuration := existingHistoryRecord.Duration + duration

	return db.UpdateTrackStatsForGuild(existingTrackRecord.ID, guildID, newPlayCount, newDuration)
}

// AddPlaybackCountStats updates playback count statistics for a track.
func (h *History) AddPlaybackCountStats(guildID, songID string) error {

	existingTrackRecord, err := db.GetTrackBySongID(songID)
	if err != nil {
		return err
	}

	existingHistoryRecord, err := db.GetHistoryByTrackIDAndGuildID(existingTrackRecord.ID, guildID)
	if err != nil {
		return err
	}

	newPlayCount := existingHistoryRecord.PlayCount + 1
	newDuration := existingHistoryRecord.Duration

	return db.UpdateTrackStatsForGuild(existingTrackRecord.ID, guildID, newPlayCount, newDuration)
}

// AddPlaybackDurationStats updates playback duration statistics for a track.
func (h *History) AddPlaybackDurationStats(guildID, songID string, duration float64) error {

	existingTrackRecord, err := db.GetTrackBySongID(songID)
	if err != nil {
		return err
	}

	existingHistoryRecord, err := db.GetHistoryByTrackIDAndGuildID(existingTrackRecord.ID, guildID)
	if err != nil {
		return err
	}

	newPlayCount := existingHistoryRecord.PlayCount
	newDuration := existingHistoryRecord.Duration + duration

	return db.UpdateTrackStatsForGuild(existingTrackRecord.ID, guildID, newPlayCount, newDuration)
}

// GetHistory retrieves the play history for a guild, sorted by the specified criteria.
func (h *History) GetHistory(guildID string, sortBy string) ([]HistoryTrackInfo, error) {
	var historyEntries []db.History
	var err error

	if guildID == "" {
		historyEntries, err = db.GetAllHistorySortedBy(sortBy)
		if err != nil {
			return nil, err
		}
	} else {
		historyEntries, err = db.GetGuildHistorySortedBy(guildID, sortBy)
		if err != nil {
			return nil, err
		}
	}

	var historyWithTracks []HistoryTrackInfo

	for _, historyEntry := range historyEntries {

		track, err := db.GetTrackByID(historyEntry.TrackID)
		if err != nil {
			return nil, err
		}

		combinedInfo := HistoryTrackInfo{
			History: historyEntry,
			Track:   *track,
		}

		historyWithTracks = append(historyWithTracks, combinedInfo)
	}

	return historyWithTracks, nil
}

// GetTrackFromHistory retrieves a track from the play history based on its ID and guild.
func (h *History) GetTrackFromHistory(guildID string, trackID uint) (db.Track, error) {
	exists, err := db.DoesHistoryExistForGuild(trackID, guildID)
	if err != nil {
		return db.Track{}, err
	}

	if exists {
		track, err := db.GetTrackByID(trackID)
		if err != nil {
			return db.Track{}, err
		}

		return *track, nil
	}

	return db.Track{}, err
}
