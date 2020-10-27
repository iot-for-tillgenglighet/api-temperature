package models

import (
	"gorm.io/gorm"
)

type Temperature struct {
	gorm.Model
	Latitude  float64
	Longitude float64
	Device    string `gorm:"unique_index:idx_device_timestamp;index:idx_device_temp"`
	Temp      float32
	Water     bool
	Timestamp string `gorm:"unique_index:idx_device_timestamp"`
}
