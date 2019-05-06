package core

import (
	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

// ProcessedNotification model
type ProcessedNotification struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	SiteID        primitive.ObjectID `bson:"siteId"`
	TimeToLive    int                `bson:"timeToLive"`
	LaunchURL     string             `bson:"launchUrl"`
	Message       Message
	Browser       []Browser
	HideRules     HideRule     `bson:"hideRules"`
	TotalSent     int          `bson:"totalSent"`
	SendTo        SendTo       `bson:"sendTo"`
	IsAtLocalTime bool         `bson:"isAtLocalTime"`
	IsFcmEnabled  bool         `bson:"isFcmEnabled"`
	FcmSenderId   string       `bson:"FcmSenderId"`
	FcmServerKey  string       `bson:"FcmServerKey"`
	VapidDetails  VapidDetails `bson:"vapidDetails"`
	Timezone      string
	NoOfCalls     int    `bson:"noOfCalls"`
	LastID        string `bson:"lastId"`
}
