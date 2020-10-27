package models

import (
	"time"

	"gorm.io/gorm"
)

//Temperature defines the structure for our temperatures table
type Temperature struct {
	gorm.Model
	Latitude   float64
	Longitude  float64
	Device     string `gorm:"unique_index:idx_device_timestamp"`
	Temp       float32
	Water      bool
	Timestamp  string    `gorm:"unique_index:idx_device_timestamp"`
	Timestamp2 time.Time `gorm:"default:'1970-01-01T12:00:00Z'"`
}
