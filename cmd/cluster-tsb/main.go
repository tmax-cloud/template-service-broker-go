package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tmax-cloud/template-service-broker-go/pkg/server/apis"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const (
	port                  = 8081
	apiPathPrefix         = "/v2/"
	serviceCatalogPrefix  = "/catalog"
	serviceInstancePrefix = "/service_instances/{instanceId}"
	serviceBindingPrefix  = "/service_instances/{instance_id}/service_bindings/{binding_id}"
)

var log = logf.Log.WithName("TSB-main")

func main() {
	logf.SetLogger(zap.Logger(true))
	log.Info("initializing server....")

	router := mux.NewRouter()
	apiRouter := router.PathPrefix(apiPathPrefix).Subrouter()

	//catalog
	apiRouter.HandleFunc(serviceCatalogPrefix, apis.GetClusterCatalog).Methods("GET")

	//provision
	apiRouter.HandleFunc(serviceInstancePrefix, apis.ClusterProvisionServiceInstance).Methods("PUT")
	apiRouter.HandleFunc(serviceInstancePrefix, apis.ClusterDeprovisionServiceInstance).Methods("DELETE")

	//binding
	apiRouter.HandleFunc(serviceBindingPrefix, apis.ClusterBindingServiceInstance).Methods("PUT")
	apiRouter.HandleFunc(serviceBindingPrefix, apis.UnBindingServiceInstance).Methods("DELETE")

	http.Handle("/", router)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Error(err, "failed to initialize a server")
	}
}
