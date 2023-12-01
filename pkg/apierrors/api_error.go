package apierrors

import (
	"fmt"

	"github.com/google/uuid"
	json "github.com/json-iterator/go"
)

type DockApiErrorDetail struct {
	Attribute string   `json:"attribute"`
	Messages  []string `json:"messages"`
}

type DockApiInnerError struct {
	Id           string               `json:"id"`
	Code         string               `json:"code"`
	Description  string               `json:"description"`
	ErrorDetails []DockApiErrorDetail `json:"error_details,omitempty"`
}

func (daie *DockApiInnerError) Error() string {
	return fmt.Sprintf("id=%s,code=%s,description=%s", daie.Id, daie.Code, daie.Description)
}

type DockApiError struct {
	Error      DockApiInnerError `json:"error"`
	StatusCode int               `json:"-"`
}

func MakeDockApiErrorCode(baseCode, detailCode string) string {
	return fmt.Sprintf("%s-%s", baseCode, detailCode)
}

func NewDockApiError(statusCode int, code, description string) *DockApiError {
	return &DockApiError{
		StatusCode: statusCode,
		Error: DockApiInnerError{
			Id:          uuid.NewString(),
			Code:        code,
			Description: description,
		},
	}
}

func (dae *DockApiError) SetId(id string) error {
	_, err := uuid.Parse(id)
	if err == nil {
		dae.Error.Id = id
	}
	return err
}

func (dae *DockApiError) AddErrorDetail(attribute string, messages ...string) {
	if dae.Error.ErrorDetails == nil {
		dae.Error.ErrorDetails = make([]DockApiErrorDetail, 0)
	}

	ed := DockApiErrorDetail{
		Attribute: attribute,
		Messages:  messages,
	}

	dae.Error.ErrorDetails = append(dae.Error.ErrorDetails, ed)
}

func (dae *DockApiError) JsonMap() map[string]any {
	var m map[string]any

	b, _ := json.Marshal(dae)
	json.Unmarshal(b, &m)

	return m
}

func (dae *DockApiError) Bytes() []byte {
	b, _ := json.Marshal(dae)
	return b
}
