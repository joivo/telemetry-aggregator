package pusher

import "go.mongodb.org/mongo-driver/bson/primitive"

type Observation struct {
	Label       string  `json:"label" bson:"label"`
	Measurement float64 `json:"measurement" bson:"measurement"`
}

type Metric struct {
	ID           primitive.ObjectID `json:"_id, omitempty" bson:"_id, omitempty"`
	Name         string             `json:"name" bson:"name"`
	Help         string             `json:"help, omitempty" bson:"help, omitempty"`
	Timestamp    int64              `json:"timestamp" bson:"timestamp"`
	Observations []Observation      `json:"observations" bson:"observations"`
}
