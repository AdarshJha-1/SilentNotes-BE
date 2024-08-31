package types

import "silent-notes/internal/models"

type Response struct {
	StatusCode          int                         `json:"status_code"`
	Success             bool                        `json:"success"`
	Message             string                      `json:"message,omitempty"`
	Data                map[string]interface{}      `json:"data,omitempty"`
	Error               string                      `json:"error,omitempty"`
	Messages            map[string][]models.Message `json:"messages,omitempty"`
	IsAcceptingMessages bool                        `json:"is_accepting_messages,omitempty"`
}
