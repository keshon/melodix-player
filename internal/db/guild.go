package db

import (
	"gorm.io/gorm"
)

type Guild struct {
	ID     string `gorm:"primaryKey"`
	Name   string
	Prefix string
}

func CreateGuild(guild Guild) error {
	return DB.Create(&guild).Error
}

func GetGuildByID(guildID string) (*Guild, error) {
	var guild Guild
	err := DB.Where("id = ?", guildID).First(&guild).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &guild, err
}

func GetAllGuildIDs() ([]string, error) {
	var guilds []Guild
	var guildIDs []string

	if err := DB.Find(&guilds).Error; err != nil {
		return nil, err
	}

	for _, guild := range guilds {
		guildIDs = append(guildIDs, guild.ID)
	}

	return guildIDs, nil
}

func DoesGuildExist(guildID string) (bool, error) {
	var count int64
	err := DB.Model(&Guild{}).Where("id = ?", guildID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func DeleteGuild(guildID string) error {
	return DB.Where("id = ?", guildID).Delete(&Guild{}).Error
}

func SetGuildPrefix(guildID string, prefix string) error {
	return DB.Model(&Guild{}).Where("id = ?", guildID).Update("prefix", prefix).Error
}

func ResetGuildPrefix(guildID string) error {
	return DB.Model(&Guild{}).Where("id = ?", guildID).Update("prefix", "").Error
}

func GetGuildPrefix(guildID string) (string, error) {
	var guild Guild
	err := DB.Where("id = ?", guildID).First(&guild).Error
	if err == gorm.ErrRecordNotFound {
		return "", nil
	}
	return guild.Prefix, err
}
