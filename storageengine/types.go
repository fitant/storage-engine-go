package storageengine

type ReadRequest struct {
	ID   string `json:"id" bson:"id"`
	Pass string `json:"pass" bson:"id"`
}

type ReadResponse struct {
	ID   string `json:"id" bson:"id"`
	Note string `json:"note" bson:"note"`
}

type PublishRequest struct {
	ID   string `json:"id" bson:"id"`
	Pass string `json:"pass" bson:"pass"`
	Note string `json:"note" bson:"note"`
}

type PublishResponse struct {
	ID string `json:"id" bson:"id"`
}

