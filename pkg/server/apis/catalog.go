package apis

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	tmaxv1 "github.com/youngind/hypercloud-operator/pkg/apis/tmax/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"template-service-broker/internal"
	"template-service-broker/pkg/server/schemas"
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
		service := MakeService(&template)

		//if empty plan, make default plan
		if len(template.Plans) == 0 {
			plan := tmaxv1.PlanSpec{
				Id:          template.Name + "-plan-default",
				Name:        template.Name + "-plan-default",
				Description: template.Name + "-plan-default",
			}
			service.Plans = append(service.Plans, plan)
		}
		response.Services = append(response.Services, service)
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func MakeService(template *tmaxv1.Template) schemas.Service {
	//default value setting, if not set
	if template.ShortDescription == "" {
		template.ShortDescription = template.Name
	}
	if template.ImageUrl == "" {
		template.ImageUrl = "https://folo.co.kr/img/gm_noimage.png"
	}
	if template.LongDescription == "" {
		template.LongDescription = template.Name
	}
	if template.UrlDescription == "" {
		template.UrlDescription = template.Name
	}
	if template.MarkDownDescription == "" {
		template.MarkDownDescription = template.Name
	}
	if template.Provider == "" {
		template.Provider = "tmax"
	}
	//create service struct
	service := schemas.Service{
		Name:        template.Name,
		Id:          template.Name,
		Description: template.ShortDescription,
		Tags:        template.Tags,
		Bindable:    false,
		Metadata: map[string]string{
			"imageUrl":            template.ImageUrl,
			"longDescription":     template.LongDescription,
			"urlDescription":      template.UrlDescription,
			"markdownDescription": template.MarkDownDescription,
			"providerDisplayName": template.Provider,
			"recommend":           strconv.FormatBool(template.Recommand),
		},
		PlanUpdateable: false,
		Plans:          template.Plans,
	}
	//Bindable check
	for _, object := range template.Objects {
		var raw map[string]interface{}
		if err := json.Unmarshal(object.Raw, &raw); err != nil {
			logBind.Error(err, "cannot get object info")
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
