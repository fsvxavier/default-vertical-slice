package apierrors

import (
	"regexp"
	"testing"
)

func TestDockApiError_JsonMap(t *testing.T) {
	expectedCode := "test-code"
	expectedDescription := "test-description"
	dae := NewDockApiError(200, expectedCode, expectedDescription)
	expectedId := dae.Error.Id

	jm := dae.JsonMap()

	if errorNode, ok := jm["error"]; ok {
		if errMap, ok := errorNode.(map[string]any); ok {
			if id, ok := errMap["id"]; ok {
				if idS, ok := id.(string); ok {
					if idS != expectedId {
						t.Errorf("expected json node 'error->id' to be %s, got %s", expectedId, idS)
					}
				} else {
					t.Error("json node 'error->id' is not a string type")
				}
			} else {
				t.Error("json missing attribute 'error->status'")
			}

			if code, ok := errMap["code"]; ok {
				if codeS, ok := code.(string); ok {
					if codeS != expectedCode {
						t.Errorf("expected json node 'error->code' to be %s, got %s", expectedCode, codeS)
					}
				} else {
					t.Error("json node 'error->code' is not a string type")
				}
			} else {
				t.Error("json missing attribute 'error->code'")
			}

			if description, ok := errMap["code"]; ok {
				if descriptionS, ok := description.(string); ok {
					if descriptionS != expectedCode {
						t.Errorf("expected json node 'error->description' to be %s, got %s", expectedDescription, descriptionS)
					}
				} else {
					t.Error("json node 'error->description' is not a string type")
				}
			} else {
				t.Error("json missing attribute 'error->description'")
			}

		} else {
			t.Error("json node 'error' is not an object type")
		}
	} else {
		t.Error("json missing attribute 'error'")
	}
}

func TestMakeDockApiErrorCode(t *testing.T) {
	matchRegex := "^[A-Z]{3,5}-[A-Z0-9-_]{3,10}$"
	c := MakeDockApiErrorCode("TEST", "TEST")
	r, _ := regexp.Compile(matchRegex)
	if !r.MatchString(c) {
		t.Errorf("Expected MakeDockApiErrorCode to return a code matching regex %s Got %s", matchRegex, c)
	}
}
