package utils

import (
	"encoding/json"
	"github.com/wybiral/pub/pkg/types"
	"net/http"
)

func JsonResponse(w http.ResponseWriter, obj interface{}) {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	err := encoder.Encode(obj)
	if err != nil {
		JsonError(w, "marshalling error")
		return
	}
}

func JsonError(w http.ResponseWriter, msg string) {
	w.WriteHeader(http.StatusInternalServerError)
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	obj := types.Error{
		Error: msg,
	}
	err := encoder.Encode(obj)
	if err != nil {
		return
	}
}
