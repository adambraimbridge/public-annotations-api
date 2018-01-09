package annotations

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

var (
	AnnotationsDriver  driver
	CacheControlHeader string
)

// methodNotAllowedHandler handles 405
func MethodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	return
}

func GetAnnotations(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if uuid == "" {
		http.Error(w, "uuid required", http.StatusBadRequest)
		return
	}
	annotations, found, err := AnnotationsDriver.read(uuid)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		msg := fmt.Sprintf(`{"message":"Error getting annotations for content with uuid %s, err=%s"}`, uuid, err.Error())
		w.Write([]byte(msg))
		return
	}
	if !found {
		w.WriteHeader(http.StatusNotFound)
		msg := fmt.Sprintf(`{"message":"No annotations found for content with uuid %s."}`, uuid)
		w.Write([]byte(msg))
		return
	}

	w.Header().Set("Cache-Control", CacheControlHeader)
	w.WriteHeader(http.StatusOK)

	if err = json.NewEncoder(w).Encode(annotations); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg := fmt.Sprintf(`{"message":"Error parsing annotations for content with uuid %s, err=%s"}`, uuid, err.Error())
		w.Write([]byte(msg))
	}
}

func GetFilteredAnnotations(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]
	platformVersion := vars["platformVersion"]

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if uuid == "" {
		http.Error(w, "uuid required", http.StatusBadRequest)
		return
	}
	annotations, found, err := AnnotationsDriver.filteredRead(uuid, platformVersion)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		msg := fmt.Sprintf(`{"message":"Error getting annotations for content with uuid %s for platformVersion %s, err=%s"}`, uuid, platformVersion, err.Error())
		w.Write([]byte(msg))
		return
	}
	if !found {
		w.WriteHeader(http.StatusNotFound)
		msg := fmt.Sprintf(`{"message":"No annotations found for content with uuid %s for platformVersion %s."}`, uuid, platformVersion)
		w.Write([]byte(msg))
		return
	}

	w.Header().Set("Cache-Control", CacheControlHeader)
	w.WriteHeader(http.StatusOK)

	if err = json.NewEncoder(w).Encode(annotations); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg := fmt.Sprintf(`{"message":"Error parsing annotations for content with uuid %s for platformVersion %s, err=%s"}`, uuid, platformVersion, err.Error())
		w.Write([]byte(msg))
	}
}
