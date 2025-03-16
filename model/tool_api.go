package model

type ToolForm struct {
	Name        string `json:"name,omitempty" minLength:"1"`
	Summary     string `json:"summary,omitempty"`
	Description string `json:"description,omitempty"`
	Disabled    bool   `json:"disabled,omitempty" validate:"optional"`
}
