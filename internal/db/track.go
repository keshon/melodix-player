package db

type Track struct {
	ID        uint `gorm:"primaryKey;autoIncrement"`
	SongID    string
	Title     string
	URL       string
	Filepath  string
	Source    string
	Histories []History `gorm:"foreignKey:TrackID"`
}

func CreateTrack(track *Track) error {
	return DB.Create(track).Error
}

func GetTrackByID(id uint) (*Track, error) {
	var track Track
	if err := DB.First(&track, id).Error; err != nil {
		return nil, err
	}
	return &track, nil
}

func GetTrackBySongID(songID string) (*Track, error) {
	var track Track
	if err := DB.Where("song_id = ?", songID).First(&track).Error; err != nil {
		return nil, err
	}
	return &track, nil
}

func UpdateTrack(track *Track) error {
	return DB.Save(track).Error
}

func DeleteTrack(track *Track) error {
	return DB.Delete(track).Error
}

func GetAllTracks() ([]Track, error) {
	var tracks []Track
	if err := DB.Find(&tracks).Error; err != nil {
		return nil, err
	}
	return tracks, nil
}

func GetTrackByFilepath(filepath string) (*Track, error) {
	var track Track
	if err := DB.Where("filepath = ?", filepath).First(&track).Error; err != nil {
		return nil, err
	}
	return &track, nil
}

func GetTrackByURL(url string) (*Track, error) {
	var track Track
	if err := DB.Where("url = ?", url).First(&track).Error; err != nil {
		return nil, err
	}
	return &track, nil
}
