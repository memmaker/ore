package main

type SearchResult struct {
	ID    string `json:"value" bson:"id"`
	Label string `json:"searchBy" bson:"label"`
}

type JsonObject = map[string]interface{}
type JsonList = []map[string]interface{}

type ForeignRelation struct {
	Name         string `json:"name" bson:"name"`
	RelatedModel string `json:"related_model" bson:"related_model"`
	RelatedField string `json:"related_field" bson:"related_field"`
}

type EntryContainer struct {
	Name    string       `json:"name" bson:"name"`
	Model   Model        `json:"related_model" bson:"related_model"`
	Entries []JsonObject `json:"entries" bson:"entries"`
}

type FieldData struct {
	Field JsonObject
	Entry JsonObject
}

type MongoIndex struct {
	Keys   []map[string]int `json:"keys" bson:"keys"`
	Unique bool             `json:"unique" bson:"unique"`
}

type User struct {
	Username    string `bson:"username"`
	Address     string `bson:"address"`
	SiWeMessage string `bson:"siweMessage"`
}

type Model struct {
	Name             string            `json:"name" bson:"name"`
	DisplayTemplate  string            `json:"display_template" bson:"display_template,omitempty"`
	DisplayFields    []string          `json:"display_fields" bson:"display_fields,omitempty"`
	Indexes          []MongoIndex      `json:"indexes" bson:"indexes"`
	SearchFields     []string          `json:"search_fields" bson:"search_fields,omitempty"`
	Fields           []JsonObject      `json:"fields" bson:"fields"`
	ForeignRelations []ForeignRelation `json:"foreign_relations" bson:"foreign_relations"`
}

type ModelResponse struct {
	ModelNames []string         `json:"model_names"`
	Models     map[string]Model `json:"models"`
}
