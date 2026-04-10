// Package handlers contains the HTTP handlers for the SMO API.
//
// Each handler is a thin layer that:
//  1. parses and validates the HTTP request,
//  2. calls the appropriate use case,
//  3. translates errors to HTTP status codes via the http/errors package,
//  4. serializes the result as JSON.
//
// Handlers never contain business logic, never talk to the database
// directly, and never reference infrastructure-specific types beyond
// HTTP itself.
package handlers
