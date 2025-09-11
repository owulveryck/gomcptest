package chatengine

// Define structs to represent the model data
type Model struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

type ListModelsResponse struct {
	Object string  `json:"object"`
	Data   []Model `json:"data"`
}

type ListToolResponse struct {
	Name        string
	Description string
	Protocol    string
	Server      string
}
