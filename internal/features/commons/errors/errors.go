package domainerrors

import (
	"embed"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"

	json "github.com/json-iterator/go"

	"github.com/fsvxavier/default-vertical-slice/internal/utils/helpers"
)

var ErrNoRows = errors.New("no rows in result set")

// A type to encapsulate use case errors.
type (
	// Errors related to saving or querying data from a database.
	RepositoryError struct {
		InternalError error
		Description   string `json:"description"`
	}

	// A type that encapsulates errors resulting from external services.
	ExternalIntegrationError struct {
		InternalError error
		Metadata      map[string]any `json:"metadata,omitempty"`
		Data          []byte         `json:"data,omitempty"`
		Code          int            `json:"code,omitempty"`
	}

	// A type to encapsulate validation errors.
	InvalidEntityError struct {
		Details    map[string][]string `json:"details"`
		EntityName string              `json:"entity"`
	}

	UnsupportedMediaTypeError struct{}

	// Errors related to business rules.
	UsecaseError struct {
		Description string `json:"description"`
	}

	NotFoundError struct {
		Description string `json:"description"`
	}

	ServerError struct {
		InternalError error
		Metadata      map[string]any `json:"metadata,omitempty"`
		Description   string         `json:"description"`
	}
)

func (err *ExternalIntegrationError) Error() string {
	if err.InternalError == nil {
		return "integration error"
	}
	return err.InternalError.Error()
}

func (err *ExternalIntegrationError) Extra() string {
	type DockApiError struct {
		Error map[string]any
	}

	var dockError DockApiError
	json.Unmarshal(err.Data, &dockError)
	return fmt.Sprintf("%v - %d", dockError.Error["description"], err.Code)
}

func (*InvalidEntityError) Error() string {
	return "invalid entity"
}

func (u *UsecaseError) Error() string {
	return u.Description
}

func (d *RepositoryError) Error() string {
	return d.Description
}

func (d *ServerError) Error() string {
	return d.Description
}

func (d *NotFoundError) Error() string {
	return d.Description
}

func (d *UnsupportedMediaTypeError) Error() string {
	return "unsupported media type"
}

func NewInvalidEntityError(details map[string][]string, entity any) *InvalidEntityError {
	return &InvalidEntityError{
		Details:    details,
		EntityName: helpers.EntityName(entity),
	}
}

type DockError struct {
	Description string `json:"description"`
	StatusCode  int    `json:"status_code"`
}

var (
	dockErrors map[string]DockError

	//go:embed errors.json
	errorsFS embed.FS
)

func parseError(err error) string {
	r := regexp.MustCompile(`^(.*)\(SQLSTATE (.*)\).*$`)
	match := r.FindStringSubmatch(err.Error())
	if len(match) == 3 {
		fmt.Printf("Message: %s / Code: %s\n", match[1], match[2])
		return match[2]
	}

	return ""
}

func findError(err error) DockError {
	dockError, ok := dockErrors[parseError(err)]
	if !ok {
		return dockErrors["500"]
	}
	return dockError
}

func HandleDatabaseError(err error) DockError {
	return findError(err)
}

func loadErrors(filepath string, embeded bool) (map[string]DockError, error) {
	var errorMap map[string]DockError
	var file []byte
	var err error

	// NOTE: This may need revisiting in the future
	if embeded {
		file, err = errorsFS.ReadFile(filepath)
	} else {
		file, err = os.ReadFile(filepath)
	}

	if err != nil {
		return errorMap, err
	}

	err = json.Unmarshal(file, &errorMap)

	if err != nil {
		return errorMap, err
	}

	return errorMap, nil
}

// MARK: - Initialization

func init() {
	var err error
	useEmbededFS := true

	dockErrors, err = loadErrors("errors.json", useEmbededFS)

	if err != nil {
		log.Panicln("Error loading errors.json")
	}
}
