package model

type Tool struct {
	Common

	Name        string `json:"name"`
	Summary     string `json:"summary"`
	Description string `json:"description"`
	Downloads   int    `json:"downloads"`
	Disabled    bool   `json:"disabled"`
}
