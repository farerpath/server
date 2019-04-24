package model

import (
	"github.com/farerpath/server/model/uoid"
	"time"
)

// Picture model
// Picture data for rest-api
// farerpath.pictures
type Picture struct {
	PictureName   	string		`json:"pictureName" bson:"pictureName"`
	Owner         	string		`json:"owner" bson:"owner"`
	PictureID     	uoid.UOID	`json:"_id" bson:"_id"`
	TimeMetadata	time.Time	`json:"timeMetadata" bson:"timeMetadata"`
	PublishRange  	uint32		`json:"publishRange" bson:"publishRange"`
	CountNiceShot 	uint32		`json:"countNiceShot" bson:"countNiceSot"`
	Archived      	bool 		`json:"archived" bson:"archived"`
	Albums        	[]string	`json:"albums" bson:"albums"`
	Comments      	[]Comment	`json:"comments" bson:"comments"`
	Path          	Path		`json:"path" bson:"path"`
}

type Path struct {
	Country  string  		`json:"country" bson:"country"`
	City     string  		`json:"city" bson:"city"`
	Location GPSData 		`json:"location" bson:"location"`
}

type Comment struct {
	Owner     string 		`json:"owner" bson:"owner"`
	UserName  string 		`json:"userName" bson:"userName"`
	Value     string 		`json:"value" bson:"value"`
	Time      string 		`json:"time" bson:"time"`
	CommentID uoid.UOID 	`json:"_id" bson:"_id"`
}

type GPSData struct {
	Longitude string 		`json:"longitude" bson:"longitude"`
	Latitude  string 		`json:"latitude" bson:"latitude"`
	Altitude  string 		`json:"altitude" bson:"altitude"`
}
