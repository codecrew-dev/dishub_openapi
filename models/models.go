package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BotStatsRequest struct {
	Servers int    `json:"servers"`
	Shards  int    `json:"shards"`
	Status  string `json:"status"`
}

type BotResponse struct {
	ID              string            `json:"id" bson:"botId"`
	Name            string            `json:"name" bson:"name"`
	Tag             string            `json:"tag" bson:"tag"`
	Avatar          string            `json:"avatar" bson:"avatar"`
	OwnerID         string            `json:"ownerId" bson:"ownerId"`
	Tags            []string          `json:"tags" bson:"tags"`
	Prefix          string            `json:"prefix" bson:"prefix"`
	Library         *string           `json:"library" bson:"library"`
	Website         *string           `json:"website" bson:"website"`
	SupportServer   *string           `json:"supportServer" bson:"supportServer"`
	InviteUrl       *string           `json:"inviteUrl" bson:"inviteUrl"`
	Banner          *string           `json:"banner" bson:"banner"`
	BotLangs        []string          `json:"botLangs" bson:"botLangs"`
	Servers         int               `json:"servers" bson:"serverCount"`
	Shards          int               `json:"shards" bson:"shards"`
	Votes           int               `json:"votes" bson:"hearts"`
	Description     map[string]string `json:"description" bson:"description"`
	LongDescription map[string]string `json:"longDescription" bson:"longDescription"`
	DescLang        string            `json:"descLang" bson:"descLang"`
	DiscordVerified bool              `json:"discordVerified" bson:"discordVerified"`
	Status          string            `json:"status" bson:"status"`
	LastStatsAt     *time.Time        `json:"lastStatsAt,omitempty" bson:"lastStatsAt,omitempty"`
	Badge           bool              `json:"badge" bson:"badge"`
	Rating          float64           `json:"rating" bson:"rating"`
	ReviewCount     int               `json:"reviewCount" bson:"reviewCount"`
}

type BotVotedResponse struct {
	Voted bool `json:"voted"`
}

type DeveloperApp struct {
	ID                primitive.ObjectID `bson:"_id"`
	OwnerID           string             `bson:"ownerId"`
	TargetType        string             `bson:"targetType"`
	TargetId          string             `bson:"targetId"`
	TokenPrefix       string             `bson:"tokenPrefix"`
	TokenHash         string             `bson:"tokenHash"`
	WebhookURL        string             `json:"webhookURL" bson:"webhookUrl"`
	WebhookSecret     string             `json:"webhookSecret" bson:"webhookSecret"`
	WebhookEvents     []string           `json:"webhookEvents" bson:"webhookEvents"`
	DiscordWebhookURL string             `json:"discordWebhookURL" bson:"discordWebhookUrl"`
}

type ServerActivityResponse struct {
	Date         time.Time `json:"date" bson:"date"`
	MessageCount int       `json:"messageCount" bson:"messageCount"`
	VoiceMinutes float64   `json:"voiceMinutes" bson:"voiceMinutes"`
}

type ServerResponse struct {
	ID              string                   `json:"id" bson:"serverId"`
	Name            string                   `json:"name" bson:"name"`
	Description     map[string]string        `json:"description" bson:"description"`
	LongDescription map[string]string        `json:"longDescription" bson:"longDescription"`
	Icon            string                   `json:"icon" bson:"icon"`
	OwnerID         string                   `json:"ownerId" bson:"ownerId"`
	Tags            []string                 `json:"tags" bson:"tags"`
	Votes           int                      `json:"votes" bson:"hearts"`
	Members         int                      `json:"members" bson:"memberCount"`
	InviteUrl       string                   `json:"inviteUrl" bson:"inviteUrl"`
	Verified        bool                     `json:"verified" bson:"verified"`
	BoostTier       int                      `json:"boostTier" bson:"boostTier"`
	ServerLangs     []string                 `json:"serverLangs" bson:"serverLangs"`
	DescLang        string                   `json:"descLang" bson:"descLang"`
	Rating          float64                  `json:"rating" bson:"rating"`
	ReviewCount     int                      `json:"reviewCount" bson:"reviewCount"`
	Activities      []ServerActivityResponse `json:"activities,omitempty" bson:"-"`
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

type ReviewModel struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	TargetType string             `json:"targetType" bson:"targetType"`
	TargetId   string             `json:"targetId" bson:"targetId"`
	UserID     string             `json:"userId" bson:"userId"`
	Rating     float64            `json:"rating" bson:"rating"`
	Content    string             `json:"content" bson:"content"`
	CreatedAt  time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt  time.Time          `json:"updatedAt" bson:"updatedAt"`
}

type ReviewRequest struct {
	UserID  string  `json:"userId" binding:"required"`
	Rating  float64 `json:"rating" binding:"required,min=1,max=5"`
	Content string  `json:"content" binding:"required"`
}

type PaginatedResponse[T any] struct {
	Data       []T   `json:"data"`
	Total      int64 `json:"total"`
	Page       int64 `json:"page"`
	TotalPages int64 `json:"totalPages"`
	Limit      int64 `json:"limit"`
}

type DiscordEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

type DiscordFooter struct {
	Text string `json:"text"`
}

type DiscordAuthor struct {
	Name    string `json:"name"`
	URL     string `json:"url,omitempty"`
	IconURL string `json:"icon_url,omitempty"`
}

type DiscordEmbed struct {
	Author      *DiscordAuthor      `json:"author,omitempty"`
	Title       string              `json:"title"`
	Description string              `json:"description,omitempty"`
	Color       int                 `json:"color,omitempty"`
	Fields      []DiscordEmbedField `json:"fields,omitempty"`
	Timestamp   string              `json:"timestamp,omitempty"`
	Footer      *DiscordFooter      `json:"footer,omitempty"`
}
