package apis

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	tmaxv1 "github.com/tmax-cloud/template-operator/api/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/kubernetes-sigs/service-catalog/pkg/controller"
	"github.com/kubernetes-sigs/service-catalog/pkg/util"
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
		service := MakeService(template.Name, &template.TemplateSpec, string(template.UID))
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
		service := MakeService(template.Name, &template.TemplateSpec, string(template.UID))
		response.Services = append(response.Services, service)
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func MakeService(templateName string, templateSpec *tmaxv1.TemplateSpec, uid string) schemas.Service {
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
		Id:          uid,
		Description: templateSpec.ShortDescription,
		Tags:        templateSpec.Tags,
		Bindable:    false,
		Metadata: map[string]string{
			"serviceClassRefName": util.GenerateSHA(controller.GenerateEscapedName(uid)),
			"imageUrl":            templateSpec.ImageUrl,
			"longDescription":     templateSpec.LongDescription,
			"urlDescription":      templateSpec.UrlDescription,
			"markdownDescription": templateSpec.MarkDownDescription,
			"providerDisplayName": templateSpec.Provider,
			"recommend":           strconv.FormatBool(templateSpec.Recommend),
		},
		PlanUpdateable: false,
	}
	//default parameter setting
	properties := make(map[string]schemas.PropertiesSpec)
	var requiredParamters []string
	for _, parameter := range templateSpec.Parameters {
		property := schemas.PropertiesSpec{
			Default:     parameter.Value,
			Description: parameter.Description,
			Type:        parameter.ValueType,
		}
		if parameter.Required {
			requiredParamters = append(requiredParamters, parameter.Name)
		}
		properties[parameter.Name] = property
	}

	//plan parameter setting & plan setting
	var Plans []schemas.PlanSpec
	for i, templatePlan := range templateSpec.Plans {
		planParameters := templatePlan.Schemas.ServiceInstance.Create.Parameters
		catalogProperties := make(map[string]schemas.PropertiesSpec)
		for key, _ := range properties {
			property := properties[key]
			if paramVal, ok := planParameters[key]; ok {
				property.Default = paramVal
				property.Fixed = true
			} else {
				property.Fixed = false
			}
			catalogProperties[key] = property
		}
		plan := schemas.PlanSpec{
			Id:          uid + "-" + strconv.Itoa(i),
			Name:        templatePlan.Name,
			Description: templatePlan.Description,
			Metadata: schemas.PlanMetadata{
				Bullets: templatePlan.Metadata.Bullets,
				Costs: schemas.Cost{
					Amount: templatePlan.Metadata.Costs.Amount,
					Unit:   templatePlan.Metadata.Costs.Unit,
				},
				DisplayName: templatePlan.Metadata.DisplayName,
			},
			Free:                   templatePlan.Free,
			Bindable:               templatePlan.Bindable,
			PlanUpdateable:         templatePlan.PlanUpdateable,
			MaximumPollingDuration: templatePlan.MaximumPollingDuration,
			MaintenanceInfo: schemas.MaintenanceInfo{
				Version:     templatePlan.MaintenanceInfo.Version,
				Description: templatePlan.MaintenanceInfo.Description,
			},
			Schemas: schemas.Schemas{
				ServiceInstance: schemas.ServiceInstanceSchema{
					Create: schemas.SchemaParameters{
						Parameters: schemas.SchemaParameterSpec{
							Properties: catalogProperties,
							Required:   requiredParamters,
						},
					},
				},
			},
		}
		if len(plan.Name) == 0 {
			plan.Name = templateName + "-" + "plan" + "-" + strconv.Itoa(i)
		}
		if len(plan.Description) == 0 {
			plan.Description = templateName + "-" + "plan" + "-" + strconv.Itoa(i)
		}
		Plans = append(Plans, plan)
	}
	service.Plans = Plans

	//default plan setting in case of no plan
	if len(service.Plans) == 0 {
		plan := schemas.PlanSpec{
			Id:          uid + "-plan-default",
			Name:        templateName + "-plan-default",
			Description: templateName + "-plan-default",
			Schemas: schemas.Schemas{
				ServiceInstance: schemas.ServiceInstanceSchema{
					Create: schemas.SchemaParameters{
						Parameters: schemas.SchemaParameterSpec{
							Properties: properties,
							Required:   requiredParamters,
						},
					},
				},
			},
		}
		service.Plans = append(service.Plans, plan)
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
