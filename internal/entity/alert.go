package entity

type Alert struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}
