// Package rr is a package for convinient reading of json data from an http request and writing json to reponse.
package rr

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// JSONResponse struct for writing http reponse.
type JSONResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// ReadRespond the main struct that stores all methods for reading and writing response.
type ReadRespond struct {
	maxBytes int
}

// ReadRespondOption type for ReadRespond initialization options.
type ReadRespondOption func(*ReadRespond)

// WithMaxBytes constructor option functions for setting maxBytes option.
func WithMaxBytes(maxBytes int) ReadRespondOption {
	return func(r *ReadRespond) {
		r.maxBytes = maxBytes
	}
}

// NewReadRespond ReadRespond constructor funciton.
func NewReadRespond(options ...ReadRespondOption) *ReadRespond {
	rr := &ReadRespond{}

	for _, option := range options {
		option(rr)
	}
	return rr
}

// ReadJSON method reads json data from request to specified struct and returns an error if something went wrong.
func (rr *ReadRespond) ReadJSON(w http.ResponseWriter, r *http.Request, data any) error {
	if rr.maxBytes > 0 {
		r.Body = http.MaxBytesReader(w, r.Body, int64(rr.maxBytes))
	}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	decoder.DisallowUnknownFields()

	if err := decoder.Decode(data); err != nil {
		return err
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("body must contain a single JSON object")
	}

	return nil
}

// WriteJSON writes json response with provided data and status, and additional headers if specified. Returns an error if something went wrong.
func (rr *ReadRespond) WriteJSON(w http.ResponseWriter, status int, data any, headers ...http.Header) error {
	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(data)
}

// WriteJSONError writes json response with provided error message and additional status code if specified. 
// Default stauts is 400. Returns an error if something went wrong.
func (rr *ReadRespond) WriteJSONError(w http.ResponseWriter, err error, status ...int) error {

	statusCode := http.StatusBadRequest
	if len(status) > 0 {
		statusCode = status[0]
	}

	response := &JSONResponse{
		Error:   true,
		Message: err.Error(),
	}

	return rr.WriteJSON(w, statusCode, response)
}

