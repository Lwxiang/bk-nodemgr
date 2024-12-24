package storage

import (
	"time"
)

// BasicInfo represents a basic info for every table.
type BasicInfo struct {
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
	IsDeleted bool      `json:"is_deleted" bson:"is_deleted"`
}

func NewBasicInfo() BasicInfo {
	return BasicInfo{
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsDeleted: false,
	}
}
