package pagination

import (
	"encoding/json"
)

type (
	pagination struct {
		Page  *int `json:"page,omitempty"`
		Limit *int `json:"limit,omitempty"`
	}

	sort struct {
		Field string `json:"field,omitempty"`
		Order string `json:"order,omitempty"`
	}

	Metadata struct {
		Pagination *pagination `json:"pagination,omitempty"`
		Sort       *sort       `json:"sort,omitempty"`
	}
)

func NewMetadata(page, limit int, sortField, order *string) *Metadata {
	metadata := new(Metadata)

	metadata.Pagination = new(pagination)
	metadata.Pagination.Limit = &limit
	metadata.Pagination.Page = &page

	if sortField != nil && order != nil {
		metadata.Sort = new(sort)
		metadata.Sort.Order = *order
		metadata.Sort.Field = *sortField
	}

	return metadata
}

type PaginatedOutput struct {
	Content  any       `json:"content"`
	Metadata *Metadata `json:"metadata,omitempty"`
}

func NewPaginatedOutput(body any, pagination *Metadata) *PaginatedOutput {
	b, err := json.Marshal(body)
	if err != nil || len(b) == 0 || string(b) == "null" {
		body = make([]any, 0)
	}

	output := &PaginatedOutput{
		Content:  body,
		Metadata: pagination,
	}

	return output
}
