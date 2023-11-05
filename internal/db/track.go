package db

type Track struct {
	ID        uint `gorm:"primaryKey;autoIncrement"`
	YTID      string
	Name      string
	URL       string
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

func GetTrackByYTID(ytid string) (*Track, error) {
	var track Track
	if err := DB.Where("yt_id = ?", ytid).First(&track).Error; err != nil {
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
