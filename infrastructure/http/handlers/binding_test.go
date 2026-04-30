package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type bindStrictTarget struct {
	Name string `json:"name" binding:"required,min=1,max=10"`
}

func newBindContext(t *testing.T, body string, nilBody bool) *gin.Context {
	t.Helper()
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	if nilBody {
		c.Request, _ = http.NewRequest(http.MethodPost, "/", nil)
		c.Request.Body = nil
		return c
	}
	req, err := http.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	if err != nil {
		t.Fatalf("test setup: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	return c
}

func TestBindStrictJSON_DecodesValidPayload(t *testing.T) {
	t.Parallel()

	c := newBindContext(t, `{"name":"OK"}`, false)
	var target bindStrictTarget
	if err := bindStrictJSON(c, &target); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if target.Name != "OK" {
		t.Errorf("expected name 'OK', got %q", target.Name)
	}
}

func TestBindStrictJSON_ReturnsError_WhenRequestBodyIsNil(t *testing.T) {
	t.Parallel()

	c := newBindContext(t, "", true)
	var target bindStrictTarget
	err := bindStrictJSON(c, &target)
	if err == nil {
		t.Fatal("expected error for nil request body, got nil")
	}
	if !strings.Contains(err.Error(), "empty request body") {
		t.Errorf("expected 'empty request body' error, got %q", err.Error())
	}
}

func TestBindStrictJSON_ReturnsErrUnknownField_WhenBodyHasUnknownKey(t *testing.T) {
	t.Parallel()

	c := newBindContext(t, `{"name":"OK","sneaky":42}`, false)
	var target bindStrictTarget
	err := bindStrictJSON(c, &target)
	if err == nil {
		t.Fatal("expected error for unknown field, got nil")
	}
	if !errors.Is(err, errUnknownField) {
		t.Errorf("expected errUnknownField, got %v", err)
	}
	if !strings.Contains(err.Error(), "sneaky") {
		t.Errorf("expected error to mention the field name 'sneaky', got %q", err.Error())
	}
}

func TestBindStrictJSON_ReturnsValidationError_WhenStructTagFails(t *testing.T) {
	t.Parallel()

	c := newBindContext(t, `{"name":"this name exceeds ten chars"}`, false)
	var target bindStrictTarget
	err := bindStrictJSON(c, &target)
	if err == nil {
		t.Fatal("expected validator error, got nil")
	}
	if errors.Is(err, errUnknownField) {
		t.Errorf("expected validator error, got errUnknownField: %v", err)
	}
}

func TestBindStrictJSON_ReturnsDecodeError_WhenBodyIsMalformedJSON(t *testing.T) {
	t.Parallel()

	c := newBindContext(t, `{not json`, false)
	var target bindStrictTarget
	err := bindStrictJSON(c, &target)
	if err == nil {
		t.Fatal("expected decode error, got nil")
	}
	if errors.Is(err, errUnknownField) {
		t.Errorf("malformed JSON should not classify as unknown field, got %v", err)
	}
}
