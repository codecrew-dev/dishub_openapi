package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BotStatsRequest struct {
	Servers int `json:"servers"`
	Shards  int `json:"shards"`
}

type BotResponse struct {
	ID          string   `json:"id" bson:"botId"`
	Name        string   `json:"name" bson:"name"`
	Tag           string   `json:"tag" bson:"tag"`
	Avatar        string   `json:"avatar" bson:"avatar"`
	OwnerID       string   `json:"ownerId" bson:"ownerId"`
	Tags          []string `json:"tags" bson:"tags"`
	Prefix        string   `json:"prefix" bson:"prefix"`
	Library       *string  `json:"library" bson:"library"`
	Website       *string  `json:"website" bson:"website"`
	SupportServer *string  `json:"supportServer" bson:"supportServer"`
	InviteUrl     *string  `json:"inviteUrl" bson:"inviteUrl"`
	Banner        *string  `json:"banner" bson:"banner"`
	BotLangs      []string `json:"botLangs" bson:"botLangs"`
	Servers     int      `json:"servers" bson:"serverCount"`
	Shards      int      `json:"shards" bson:"shards"`
	Votes       int      `json:"votes" bson:"hearts"`
	Description   map[string]string `json:"description" bson:"description"`
	LongDescription map[string]string `json:"longDescription" bson:"longDescription"`
	DescLang      string   `json:"descLang" bson:"descLang"`
	DiscordVerified bool `json:"discordVerified" bson:"discordVerified"`
	Status      string   `json:"status" bson:"status"`
	Badge    bool     `json:"badge" bson:"badge"`
}

type BotVotedResponse struct {
	Voted bool `json:"voted"`
}

type DeveloperApp struct {
	ID            primitive.ObjectID `bson:"_id"`
	OwnerID       string             `bson:"ownerId"`
	TargetType    string             `bson:"targetType"`
	TargetId      string             `bson:"targetId"`
	TokenPrefix   string             `bson:"tokenPrefix"`
	TokenHash     string             `bson:"tokenHash"`
	WebhookURL    string             `json:"webhookURL" bson:"webhookURL"`
	WebhookSecret string             `json:"webhookSecret" bson:"webhookSecret"`
	WebhookEvents []string           `json:"webhookEvents" bson:"webhookEvents"`
}

type ServerResponse struct {
	ID          string   `json:"id" bson:"serverId"`
	Name        string   `json:"name" bson:"name"`
	Description map[string]string   `json:"description" bson:"description"`
	LongDescription map[string]string `json:"longDescription" bson:"longDescription"`
	Icon        string   `json:"icon" bson:"icon"`
	OwnerID     string   `json:"ownerId" bson:"ownerId"`
	Tags        []string `json:"tags" bson:"tags"`
	Votes       int      `json:"votes" bson:"hearts"`
	Members     int      `json:"members" bson:"memberCount"`
	InviteUrl   string   `json:"inviteUrl" bson:"inviteUrl"`
	Verified    bool     `json:"verified" bson:"verified"`
	BoostTier   int      `json:"boostTier" bson:"boostTier"`
	ServerLangs []string `json:"serverLangs" bson:"serverLangs"`
	DescLang    string   `json:"descLang" bson:"descLang"`
}

type ServerVotedResponse struct {
	Voted bool `json:"voted"`
}

type UserResponse struct {
	ID         string           `json:"id" bson:"discordId"`
	Username   string           `json:"username" bson:"username"`
	GlobalName string           `json:"globalName" bson:"globalName"`
	Avatar     string           `json:"avatar" bson:"avatar"`
	Badges     []string         `json:"badges" bson:"badges"`
	Bots       []BotResponse    `json:"bots" bson:"bots"`
	Servers    []ServerResponse `json:"servers" bson:"servers"`
}

type TeamMember struct {
	UserID   string    `json:"userId" bson:"userId"`
	JoinedAt time.Time `json:"joinedAt" bson:"joinedAt"`
}

type TeamResponse struct {
	ID              primitive.ObjectID `json:"id" bson:"_id"`
	Name            string             `json:"name" bson:"name"`
	Description     string             `json:"description" bson:"description"`
	Avatar          string             `json:"avatar" bson:"avatar"`
	Website         string             `json:"website" bson:"website"`
	GitUrl          string             `json:"gitUrl" bson:"gitUrl"`
	OwnerID         string             `json:"ownerId" bson:"ownerId"`
	DiscordUrl      string             `json:"discordUrl" bson:"discordUrl"`
	DescriptionLang string             `json:"descriptionLang" bson:"descriptionLang"`
	Descriptions    map[string]string  `json:"descriptions" bson:"descriptions"`
	Members         []TeamMember       `json:"members" bson:"members"`
	CreatedAt       time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt       time.Time          `json:"updatedAt" bson:"updatedAt"`
}
