package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type BotStatsRequest struct {
	Servers int `json:"servers"`
	Shards  int `json:"shards"`
}

type BotResponse struct {
	Submission    struct {
		Library       string            `bson:"library"`
		Website       string            `bson:"website"`
		SupportServer string            `bson:"supportServer"`
		InviteUrl     string            `bson:"inviteUrl"`
		BotLangs      []string          `bson:"botLangs"`
		ShortDescs    map[string]string `bson:"shortDescs"`
		LongDescs     map[string]string `bson:"longDescs"`
	} `json:"-" bson:"submission"`
	ID          string   `json:"id" bson:"botId"`
	Name        string   `json:"name" bson:"name"`
	Tag           string   `json:"tag" bson:"tag"`
	Avatar        string   `json:"avatar" bson:"avatar"`
	OwnerID       string   `json:"ownerId" bson:"ownerId"`
	Tags          []string `json:"tags" bson:"tags"`
	Prefix        string   `json:"prefix" bson:"prefix"`
	Library       string   `json:"library" bson:"-"`
	Website       string   `json:"website" bson:"-"`
	SupportServer string   `json:"supportServer" bson:"-"`
	InviteUrl     string   `json:"inviteUrl" bson:"-"`
	BotLangs      []string `json:"botLangs" bson:"-"`
	Servers     int      `json:"servers" bson:"serverCount"`
	Shards      int      `json:"shards" bson:"shards"`
	Votes       int      `json:"votes" bson:"hearts"`
	ShortDescs    map[string]string `json:"shortDescs" bson:"-"`
	LongDescs     map[string]string `json:"longDescs" bson:"-"`
	DiscordVerified bool `json:"discordVerified" bson:"discordVerified"`
	Status      string   `json:"status" bson:"status"`
	Badge    bool     `json:"badge" bson:"badge"`
}

type BotVotedResponse struct {
	Voted bool `json:"voted"`
}

type DeveloperApp struct {
	ID          primitive.ObjectID `bson:"_id"`
	OwnerID     string             `bson:"ownerId"`
	TargetType  string             `bson:"targetType"`
	TargetId    string             `bson:"targetId"`
	TokenPrefix string             `bson:"tokenPrefix"`
	TokenHash   string             `bson:"tokenHash"`
}

type ServerResponse struct {
	ID          string   `json:"id" bson:"serverId"`
	Name        string   `json:"name" bson:"name"`
	Description string   `json:"description" bson:"description"`
	Icon        string   `json:"icon" bson:"icon"`
	OwnerID     string   `json:"ownerId" bson:"ownerId"`
	Tags        []string `json:"tags" bson:"tags"`
	Votes       int      `json:"votes" bson:"hearts"`
	Members     int      `json:"members" bson:"memberCount"`
	InviteUrl   string   `json:"inviteUrl" bson:"inviteUrl"`
	Verified    bool     `json:"verified" bson:"verified"`
	BoostTier   int      `json:"boostTier" bson:"boostTier"`
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
