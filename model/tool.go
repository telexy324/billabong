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
	FilesRaw    string   `gorm:"default:'[]'" json:"-"`
	Files       []Upload `gorm:"-" json:"files"`
}

func (m *Tool) BeforeSave(tx *gorm.DB) error {
	if data, err := json.Marshal(m.Files); err != nil {
		return err
	} else {
		m.FilesRaw = string(data)
	}
	return nil
}

func (m *Tool) AfterFind(tx *gorm.DB) error {
	if err := json.Unmarshal([]byte(m.FilesRaw), &m.Files); err != nil {
		return err
	}

	return nil
}
