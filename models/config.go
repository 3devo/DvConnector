package models

// Config model that keep tracks of application configuration
// swagger:model ConfigResponse
type Config struct {
	ID          int  `storm:"id,increment" json:"id"`
	OpenNetwork bool `json:"openNetwork"`
	UserCreated bool `json:"userCreated"`
}
