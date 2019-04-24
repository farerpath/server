package model

import (
	"time"

	"github.com/farerpath/server/model/uoid"
)

// Albums model
// MongoDB
// Album list/User
// farerpath.albumlists
type AlbumList struct {
	UserID string      `json:"_id" bson:"_id"`
	Albums []uoid.UOID `json:"albums" bson:"albums"`
}

// Album model
// MongoDB
// farerpath.albums
type Album struct {
	AlbumID      uoid.UOID   `json:"_id" bson:"_id"`
	AlbumName    string      `json:"albumName" bson:"albumName"`
	Owner        string      `json:"owner" bson:"owner"`
	BeginTime    time.Time   `json:"beginTime" bson:"beginTime"`
	EndTime      time.Time   `json:"endTime" bson:"endTime"`
	TravelPath   Path        `json:"travelPath" bson:"travelPath"`
	Members      []string    `json:"members" bson:"members"`
	Pictures     []uoid.UOID `json:"pictures" bson:"pictures"`
	PublishRange uint32      `json:"publishRange" bson:"publishRange"`
	Archived     bool        `json:"archived" bson:"archived"`
}

// DEPRECATED
// AlbumListNode is not using any more. Deprecated.
type AlbumListNode struct {
	AlbumID  string `json:"albumID" bson:"albumID"`
	IsPublic bool   `json:"isPublic" bson:"isPublic"`
}

type AlbumWithPicture struct {
	AlbumID      uoid.UOID `json:"_id" bson:"_id"`
	AlbumName    string    `json:"albumName" bson:"albumName"`
	Owner        string    `json:"owner" bson:"owner"`
	BeginTime    time.Time `json:"beginTime" bson:"beginTime"`
	EndTime      time.Time `json:"endTime" bson:"endTime"`
	TravelPath   Path      `json:"travelPath" bson:"travelPath"`
	Members      []string  `json:"members" bson:"members"`
	Pictures     []Picture `json:"pictures" bson:"pictures"`
	PublishRange uint32    `json:"publishRange" bson:"publishRange"`
	Archived     bool      `json:"archived" bson:"archived"`
}

type AlbumListWithAlbum struct {
	UserID string  `json:"_id" bson:"_id"`
	Albums []Album `json:"albums" bson:"albums"`
}

type AlbumListWithAlbumWithPicture struct {
	UserID string             `json:"_id" bson:"_id"`
	Albums []AlbumWithPicture `json:"albums" bson:"albums"`
}
