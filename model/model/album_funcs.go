package model

import "github.com/farerpath/model/uoid"

func NewAlbum() *Album {
	return &Album{AlbumID:uoid.New()}
}


func NewAlbumList(userID string) *AlbumList {
	return &AlbumList{UserID:userID}
}