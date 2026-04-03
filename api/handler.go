package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/NeerajRijhwani/peer-cdn/internal/tracker"
)

type apierror struct {
	status_code int
	message     string
	err         error
}

type apiresponse struct {
	status_code int
	message     string
	data        interface{}
}

type apiHandler func(http.ResponseWriter, *http.Request) (apiresponse, apierror)

func ApiError(status_code int, message string, errors error) apierror {
	Error := apierror{
		status_code: status_code,
		message:     message,
		err:         errors}
	debug.PrintStack()
	return Error
}

func ApiResponse(status_code int, message string, data interface{}) apiresponse {
	response := apiresponse{
		status_code: status_code,
		message:     message,
		data:        data,
	}
	return response
}

func ApiHandler(fn apiHandler, tracker *tracker.Tracker) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		type contextKey string
		const myStructKey contextKey = "tracker"
		r = r.WithContext(context.WithValue(r.Context(), myStructKey, tracker))
		response, errors := fn(w, r)
		if (errors != apierror{}) {
			err := fmt.Sprintf("error : %v \n\n Message : %s \n\n Status Code: %d ", errors.err, errors.message, errors.status_code)
			http.Error(w, err, errors.status_code)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}
