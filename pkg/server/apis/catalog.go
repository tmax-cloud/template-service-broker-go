package apis

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	tmaxv1 "github.com/tmax-cloud/template-operator/pkg/apis/tmax/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/tmax-cloud/template-service-broker-go/internal"
	"github.com/tmax-cloud/template-service-broker-go/pkg/server/schemas"
)

var logCatalog = logf.Log.WithName("Catalog")

func GetCatalog(w http.ResponseWriter, r *http.Request) {

	// set response
	response := &schemas.Catalog{}
	w.Header().Set("Content-Type", "application/json")

	//add templatelist schema
	s := scheme.Scheme
	internal.AddKnownTypes(s)
	SchemeBuilder := runtime.NewSchemeBuilder()
	if err := SchemeBuilder.AddToScheme(s); err != nil {
		logCatalog.Error(err, "cannot add TemplateList scheme")
	}

	// connect k8s client
	c, err := internal.Client(client.Options{Scheme: s})
	if err != nil {
		logCatalog.Error(err, "cannot connect k8s api server")
	}

	namespace, err := internal.Namespace()
	if err != nil {
		logCatalog.Error(err, "cannot get namespace")
	}

	// get templatelist
	templateList, err := internal.GetTemplateList(c, namespace)
	if err != nil {
		logCatalog.Error(err, "cannot get templateList info")
	}

	for _, template := range templateList.Items {
		//make service
		service := MakeService(template.Name, &template.TemplateSpec)
		response.Services = append(response.Services, service)
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func GetClusterCatalog(w http.ResponseWriter, r *http.Request) {

	// set response
	response := &schemas.Catalog{}
	w.Header().Set("Content-Type", "application/json")

	//add templatelist schema
	s := scheme.Scheme
	internal.AddKnownTypes(s)
	SchemeBuilder := runtime.NewSchemeBuilder()
	if err := SchemeBuilder.AddToScheme(s); err != nil {
		logCatalog.Error(err, "cannot add TemplateList scheme")
	}

	// connect k8s client
	c, err := internal.Client(client.Options{Scheme: s})
	if err != nil {
		logCatalog.Error(err, "cannot connect k8s api server")
	}

	// get templatelist
	clusterTemplateList, err := internal.GetClusterTemplateList(c)
	if err != nil {
		logCatalog.Error(err, "cannot get clustertemplateList info")
	}

	for _, template := range clusterTemplateList.Items {
		//make service
		service := MakeService(template.Name, &template.TemplateSpec)
		response.Services = append(response.Services, service)
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func MakeService(templateName string, templateSpec *tmaxv1.TemplateSpec) schemas.Service {
	if templateSpec.ShortDescription == "" {
		templateSpec.ShortDescription = templateName
	}

	if templateSpec.ImageUrl == "" {
		templateSpec.ImageUrl = "https://folo.co.kr/img/gm_noimage.png"
	}

	if templateSpec.LongDescription == "" {
		templateSpec.LongDescription = templateName
	}

	if templateSpec.MarkDownDescription == "" {
		templateSpec.MarkDownDescription = templateName
	}

	if templateSpec.Provider == "" {
		templateSpec.Provider = "tmax"
	}
	//create service struct
	service := schemas.Service{
		Name:        templateName,
		Id:          templateName,
		Description: templateSpec.ShortDescription,
		Tags:        templateSpec.Tags,
		Bindable:    false,
		Metadata: map[string]string{
			"imageUrl":            templateSpec.ImageUrl,
			"longDescription":     templateSpec.LongDescription,
			"urlDescription":      templateSpec.UrlDescription,
			"markdownDescription": templateSpec.MarkDownDescription,
			"providerDisplayName": templateSpec.Provider,
			"recommend":           strconv.FormatBool(templateSpec.Recommend),
		},
		PlanUpdateable: false,
		Plans:          templateSpec.Plans,
	}
	//Bindable check
	for _, object := range templateSpec.Objects {
		var raw map[string]interface{}
		if err := json.Unmarshal(object.Raw, &raw); err != nil {
			logCatalog.Error(err, "cannot get object info")
		}
		//get kind, namespace, name of object
		kind := raw["kind"].(string)
		if strings.Contains(kind, "Service") || strings.Contains(kind, "Secret") {
			service.Bindable = true
			break
		}
	}
	return service
}
