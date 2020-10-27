package models

import (
	"gorm.io/gorm"
)

type Temperature struct {
	gorm.Model
	Latitude  float64
	Longitude float64
	Device    string `gorm:"unique_index:idx_device_timestamp"`
	Temp      float32
	Water     bool
	Timestamp string `gorm:"unique_index:idx_device_timestamp"`
}
