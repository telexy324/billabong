package model

import (
	"encoding/json"
	"gorm.io/gorm"
)

type Tool struct {
	Common

	Name        string   `json:"name"`
	Summary     string   `json:"summary"`
	Description string   `json:"description"`
	Downloads   int      `json:"downloads"`
	Enabled     bool     `json:"enabled"`
	FileIdsRaw  string   `gorm:"default:'[]'" json:"-"`
	Files       []Upload `gorm:"-" json:"files"`
}

func (m *Tool) BeforeSave(tx *gorm.DB) error {
	if m.Files != nil && len(m.Files) > 0 {
		fileIds := make([]uint64, 0, len(m.Files))
		for _, file := range m.Files {
			fileIds = append(fileIds, file.ID)
		}
	}
	if data, err := json.Marshal(m.Files); err != nil {
		return err
	} else {
		m.FileIdsRaw = string(data)
	}
	return nil
}

func (m *Tool) AfterFind(tx *gorm.DB) error {
	fileIds := make([]uint64, 0)
	if err := json.Unmarshal([]byte(m.FileIdsRaw), &fileIds); err != nil {
		return err
	}
	if err := tx.Model(&Upload{}).Where("id in (?)", fileIds).Find(&m.Files).Error; err != nil {
		return err
	}
	return nil
}
