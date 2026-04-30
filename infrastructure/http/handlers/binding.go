package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// errUnknownField is returned by bindStrictJSON when the request body
// carries a field absent from the target struct.
var errUnknownField = errors.New("unknown field in request body")

// bindStrictJSON decodes the request body into target with two strict
// guarantees:
//
//  1. unknown JSON fields produce an error (instead of being silently
//     ignored as Gin's default ShouldBindJSON does); this surfaces
//     migrations explicitly to clients that still send retired fields.
//  2. struct-tag validation (`binding:"required,min=...,max=..."`) runs
//     after the decode succeeds, exactly like ShouldBindJSON.
//
// Returns errUnknownField for case 1, or the validator's error for
// case 2. Callers map both to HTTP 400.
func bindStrictJSON(c *gin.Context, target any) error {
	if c.Request.Body == nil {
		return errors.New("empty request body")
	}
	dec := json.NewDecoder(c.Request.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(target); err != nil {
		if strings.HasPrefix(err.Error(), "json: unknown field") {
			return fmt.Errorf("%w (%s)", errUnknownField, strings.TrimPrefix(err.Error(), "json: "))
		}
		return fmt.Errorf("decode json body: %w", err)
	}
	if err := binding.Validator.ValidateStruct(target); err != nil {
		return fmt.Errorf("validate body: %w", err)
	}
	return nil
}
