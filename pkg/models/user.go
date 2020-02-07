package models

import "gopkg.in/mgo.v2/bson"

type User struct {
	Id                bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	Email             string        `json:"email" bson:"email"`
	Username          string        `json:"username" bson:"username"`
	PasswordHash      []byte        `json:"passwordHash" bson:"passwordHash"`
	CreationTimestamp int64         `json:"creationTimestamp" bson:"creationTimestamp"`
}
