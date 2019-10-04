package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Metric struct {
	ID           primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name         string             `json:"name" bson:"name"`
	Help         string             `json:"help" bson:"help"`
	Timestamp    int64              `json:"timestamp" bson:"timestamp"`
	Value 		 float64			`json:"value" bson:"value"`
	Metadata     map[string]string  `json:"metadata" bson:"metadata"`
}

type Version struct {
	Tag string `json:"tag"`
}

