package types

type SendMessageType struct {
	Identifier string `json:"identifier" validate:"required,min=3,max=30"`
	Content    string `json:"content" validate:"required,min=10,max=300"`
}
