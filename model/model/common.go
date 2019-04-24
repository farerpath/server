package model

import (
	"time"
)

type Any interface{}

type LoginSession struct {
	LoginUserId  string        	`json:"loginUserID" bson:"loginUserID"`
	LoginDevices []LoginDevice 	`json:"loginDevices" bson:"loginDevices"`
}

type LoginDevice struct {
	LoginTime       string 		`json:"loginTime" bson:"loginTime"`
	LoginDeviceType int    		`json:"loginDeviceType" bson:"loginDeviceType"`
	Token           string 		`json:"token" bson:"token"`
}

type UserSession struct {
	UserID 			string 		`json:"userID" bson:"userID"`
	LoginTime       time.Time 	`json:"loginTime" bson:"loginTime"`
	RefreshToken    string 		`json:"refreshToken" bson:"refreshToken"`
	AuthToken		string		`json:"authToken" bson:"authToken"`
	LoginIP			string		`json:"loginIP" bson:"loginIP"`
	DeviceType		int32 		`json:"deviceType" bson:"deviceType"`
	LoginRegion		string		`json:"loginRegion" bson:"loginRegion"`
	Sign			string		`json:"sign" bson:"sign"`
}

// 로그인 성공하면 부여하는 값. 이를 ReturnValue의 Value 에 포함하여 보낸다.
type LoginReturnValue struct {
	UserName string 			`json:"userName" bson:"userName"`
	UserID   string 			`json:"userID" bson:"userID"`
	Country  string 			`json:"country" bson:"country"`
	RefreshToken    string 		`json:"refreshToken" bson:"refreshToken"`
	AuthToken		string		`json:"authToken" bson:"authToken"`
}

type ResponseValue struct {
	Checksum string 			`json:"checksum" bson:"checksum"`
	Value    Any    			`json:"value" bson:"value"`
}

type ReturnValue struct {
	Value      []byte 			`json:"value" bson:"value"`
	StatusCode int    			`json:"statusCode" bson:"statusCode"`
}
