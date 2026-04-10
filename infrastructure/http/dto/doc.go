// Package dto contains the request and response data transfer objects
// used by the HTTP handlers.
//
// DTOs are intentionally separate from domain entities. They have JSON
// tags, validation tags, and may have a different shape than the
// underlying entity (e.g., omitting internal fields, flattening
// nested structures, accepting strings where the domain uses typed
// IDs). Mapping between DTOs and entities happens in the handlers.
package dto
