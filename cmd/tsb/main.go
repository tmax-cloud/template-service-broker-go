package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	tmaxv1 "github.com/tmax-cloud/template-operator/api/v1"
	"github.com/tmax-cloud/template-service-broker-go/internal"
	"github.com/tmax-cloud/template-service-broker-go/pkg/server/apis"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

	s := scheme.Scheme
	if err := tmaxv1.AddToScheme(s); err != nil {
		panic(err)
	}
	c, err := internal.Client(client.Options{Scheme: s})
	if err != nil {
		panic(err)
	}

	//catalog
	catalog := apis.Catalog{
		Client: c,
		Log:    logf.Log.WithName("Catalog"),
	}
	apiRouter.HandleFunc(serviceCatalogPrefix, catalog.GetCatalog).Methods("GET")

	//provision
	provision := apis.Provision{
		Client: c,
		Log:    logf.Log.WithName("Provision"),
	}
	apiRouter.HandleFunc(serviceInstancePrefix, provision.ProvisionServiceInstance).Methods("PUT")
	apiRouter.HandleFunc(serviceInstancePrefix, provision.DeprovisionServiceInstance).Methods("DELETE")

	//binding
	apiRouter.HandleFunc(serviceBindingPrefix, apis.BindingServiceInstance).Methods("PUT")
	apiRouter.HandleFunc(serviceBindingPrefix, apis.UnBindingServiceInstance).Methods("DELETE")

	http.Handle("/", router)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Error(err, "failed to initialize a server")
	}
}
