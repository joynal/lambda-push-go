package main

import (
	"time"

	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

type VapidDetails struct {
	VapidPublicKeys  string `bson:"vapidPublicKeys"`
	VapidPrivateKeys string `bson:"vapidPrivateKeys"`
}

type DisplayCondition struct {
	Type  string `bson:"type"`
	Value int    `bson:"value"`
}

// Notification model
type NotificationAccount struct {
	ID                primitive.ObjectID `bson:"_id,omitempty"`
	DisplayCondition  DisplayCondition   `bson:"displayCondition"`
	VapidDetails      VapidDetails       `bson:"vapidDetails"`
	TotalSubscriber   int                `bson:"totalSubscriber"`
	TotalUnSubscriber int                `bson:"totalUnSubscriber"`
	Status            bool               `bson:"status"`
	HTTPSEnabled      bool               `bson:"httpsEnabled"`
	IsFcmEnabled      bool               `bson:"isFcmEnabled"`
	IsDeleted         bool               `bson:"isDeleted"`
	SiteID            primitive.ObjectID `bson:"siteId"`
	UserID            primitive.ObjectID `bson:"userId"`
	Domain            string             `bson:"domain"`
	SubDomain         string             `bson:"subDomain"`
	CreatedAt         time.Time          `bson:"createdAt"`
	UpdatedAt         time.Time          `bson:"updatedAt"`
}
