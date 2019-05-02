package core

import (
	"time"

	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

// Message - Notification message
type Message struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	Title    string
	Message  string
	Language string
}

// Action - Browser push action configs
type Action struct {
	ID     primitive.ObjectID `bson:"_id,omitempty"`
	Action string
	Title  string
	URL    string
}

// Browser push configs
type Browser struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty"`
	BrowserName        string             `bson:"browserName"`
	IconURL            string             `bson:"iconURL"`
	ImageURL           string             `bson:"imageURL"`
	Badge              string
	Vibration          bool
	IsActive           bool `bson:"isActive"`
	IsEnabledCTAButton bool `bson:"isEnabledCTAButton"`
	Actions            []Action
}

// HideRule when notification popup will be closed.
type HideRule struct {
	ID    primitive.ObjectID `bson:"_id,omitempty"`
	Type  string
	Value int
}

type SendTo struct {
	AllSubscriber bool                 `bson:"allSubscriber"`
	Segments      []primitive.ObjectID `bson:"segments"`
}

// Notification model
type Notification struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty"`
	SendTo             SendTo             `bson:"sendTo"`
	Priority           string             `bson:"priority"`
	TimeToLive         int                `bson:"timeToLive"`
	TotalSent          int                `bson:"totalSent"`
	TotalDeliver       int                `bson:"totalDeliver"`
	TotalShow          int                `bson:"totalShow"`
	TotalError         int                `bson:"totalError"`
	TotalClick         int                `bson:"totalClick"`
	TotalClose         int                `bson:"totalClose"`
	IsAtLocalTime      bool               `bson:"isAtLocalTime"`
	IsProcessed        string             `bson:"isProcessed"`
	IsSchedule         bool               `bson:"isSchedule"`
	TimezonesCompleted []string           `bson:"timezonesCompleted"`
	IsDeleted          bool               `bson:"isDeleted"`
	FromRSSFeed        bool               `bson:"fromRSSFeed"`
	SiteID             primitive.ObjectID `bson:"siteId"`
	Messages           []Message
	Browsers           []Browser
	HideRules          HideRule           `bson:"hideRules"`
	LaunchURL          string             `bson:"launchUrl"`
	UserID             primitive.ObjectID `bson:"userId"`
	SentAt             time.Time          `bson:"sentAt"`
	CreatedAt          time.Time          `bson:"createdAt"`
	UpdatedAt          time.Time          `bson:"updatedAt"`
}
