package apis

import (
	"encoding/json"
	"net/http"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("provision")

func ProvisionServiceInstance(w http.ResponseWriter, r *http.Request) {
	var m ServiceInstanceProvisionRequest
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		log.Error(err, "error occurs while decoding service instance body")
		return
	}
}

func DeprovisionServiceInstance(w http.ResponseWriter, r *http.Request) {
	var m ServiceInstanceProvisionRequest
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		log.Error(err, "error occurs while decoding service instance body")
		return
	}
}
