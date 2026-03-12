package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type BotStatsRequest struct {
	Servers int `json:"servers"`
	Shards  int `json:"shards"`
}

type BotResponse struct {
	ID          string   `json:"id" bson:"botId"`
	Name        string   `json:"name" bson:"name"`
	Description string   `json:"description" bson:"description"`
	Avatar      string   `json:"avatar" bson:"avatar"`
	OwnerID     string   `json:"ownerId" bson:"ownerId"`
	Tags        []string `json:"tags" bson:"tags"`
	Prefix      string   `json:"prefix" bson:"prefix"`
	Servers     int      `json:"servers" bson:"serverCount"`
	Shards      int      `json:"shards" bson:"shards"`
	Votes       int      `json:"votes" bson:"hearts"`
	Status      string   `json:"status" bson:"status"`
	Online      bool     `json:"online" bson:"online"`
	Verified    bool     `json:"verified" bson:"verified"`
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
	Online      bool     `json:"online" bson:"online"`
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
