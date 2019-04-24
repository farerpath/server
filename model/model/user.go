package model

import "time"

// Account model
// MongoDB
// User account model
// farerpath.accounts
type Account struct {
	Email            string        	`json:"email" bson:"email"`
	Password         string        	`json:"password" bson:"password"`
	UserID           string        	`json:"_id" bson:"_id"`
	UserName         string        	`json:"userName" bson:"userName"`
	FirstName        string        	`json:"firstName" bson:"firstName"`
	LastName         string        	`json:"lastName" bson:"lastName"`
	Country          string        	`json:"country" bson:"country"`
	ProfilePhotoPath string        	`json:"profilePhotoPath" bson:"profilePhotoPath"`
	Birthday         time.Time		`json:"age" bson:"age"`
	SessionDuration  uint32        	`json:"sessionDuration" bson:"sessionDuration"`
	IsBirthdayPublic bool          	`json:"isBirthdayPublic" bson:"isBirthdayPublic"`
	IsCountryPublic  bool          	`json:"isCountryPublic" bson:"isCountryPublic"`
	IsProfilePublic  bool          	`json:"isProfilePublic" bson:"isProfilePublic"`
}
