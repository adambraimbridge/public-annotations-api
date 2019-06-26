package annotations

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
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
		if _, err = w.Write([]byte(msg)); err != nil {
			log.WithError(err).Errorf("Error while writing response: %s", msg)
		}
		return
	}
	if !found {
		w.WriteHeader(http.StatusNotFound)
		msg := fmt.Sprintf(`{"message":"No annotations found for content with uuid %s."}`, uuid)
		if _, err = w.Write([]byte(msg)); err != nil {
			log.WithError(err).Errorf("Error while writing response: %s", msg)
		}
		return
	}

	lifecycleFilter := newLifecycleFilter()
	predicateFilter := NewAnnotationsPredicateFilter()
	chain := newAnnotationsFilterChain(lifecycleFilter, predicateFilter)

	annotations = chain.doNext(annotations)

	w.Header().Set("Cache-Control", CacheControlHeader)
	w.WriteHeader(http.StatusOK)

	if err = json.NewEncoder(w).Encode(annotations); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg := fmt.Sprintf(`{"message":"Error parsing annotations for content with uuid %s, err=%s"}`, uuid, err.Error())
		if _, err = w.Write([]byte(msg)); err != nil {
			log.WithError(err).Errorf("Error while writing response: %s", msg)
		}
	}
}
