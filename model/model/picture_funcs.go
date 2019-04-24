package model

import (
	"github.com/farerpath/model/uoid"
)

func NewPicture() *Picture {
	return &Picture{PictureID:uoid.New()}
}