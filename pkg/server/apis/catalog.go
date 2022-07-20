package apis

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	tmaxv1 "github.com/tmax-cloud/template-operator/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubernetes-sigs/service-catalog/pkg/controller"
	"github.com/kubernetes-sigs/service-catalog/pkg/util"
	"github.com/tmax-cloud/template-service-broker-go/internal"
	"github.com/tmax-cloud/template-service-broker-go/pkg/server/schemas"
)

type Catalog struct {
	client.Client
	Log logr.Logger
}

func (c *Catalog) GetCatalog(w http.ResponseWriter, r *http.Request) {
	// set response
	response := &schemas.Catalog{}
	w.Header().Set("Content-Type", "application/json")

	namespace, err := internal.Namespace()
	if err != nil {
		c.Log.Error(err, "cannot get namespace")
		respond(w, http.StatusInternalServerError, &schemas.Error{
			Error:            "InternalServerError",
			Description:      "cannot get namespace. Check it is operated on cluster.",
			InstanceUsable:   false,
			UpdateRepeatable: false,
		}, c.Log)
		return
	}

	// get templatelist
	templateList, err := internal.GetTemplateList(c.Client, namespace)
	if err != nil {
		c.Log.Error(err, "error occurs while getting templateList")
		respond(w, http.StatusInternalServerError, &schemas.Error{
			Error:            "InternalServerError",
			Description:      fmt.Sprintf("cannot find templateList on the %s namespace", namespace),
			InstanceUsable:   false,
			UpdateRepeatable: false,
		}, c.Log)
		return
	}

	for _, template := range templateList.Items {
		//make service
		service := c.MakeService(template.Name, &template.TemplateSpec, string(template.UID))
		response.Services = append(response.Services, service)
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (c *Catalog) GetClusterCatalog(w http.ResponseWriter, r *http.Request) {
	// set response
	response := &schemas.Catalog{}
	w.Header().Set("Content-Type", "application/json")

	// get templatelist
	clusterTemplateList, err := internal.GetClusterTemplateList(c.Client)
	if err != nil {
		c.Log.Error(err, "error occurs while getting templateList")
		respond(w, http.StatusInternalServerError, &schemas.Error{
			Error:            "InternalServerError",
			Description:      "cannot find cluster-templateList",
			InstanceUsable:   false,
			UpdateRepeatable: false,
		}, c.Log)
	}

	for _, template := range clusterTemplateList.Items {
		//make service
		service := c.MakeService(template.Name, &template.TemplateSpec, string(template.UID))
		response.Services = append(response.Services, service)
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (c *Catalog) MakeService(templateName string, templateSpec *tmaxv1.TemplateSpec, uid string) schemas.Service {
	//create service struct
	service := schemas.Service{
		Name:        templateName,
		Id:          uid,
		Description: templateSpec.ShortDescription,
		Tags:        templateSpec.Tags,
		Bindable:    false,
		Metadata: map[string]interface{}{
			"serviceClassRefName": util.GenerateSHA(controller.GenerateEscapedName(uid)),
			"imageUrl":            templateSpec.ImageUrl,
			"longDescription":     templateSpec.LongDescription,
			"urlDescription":      templateSpec.UrlDescription,
			"markdownDescription": templateSpec.MarkDownDescription,
			"providerDisplayName": templateSpec.Provider,
			"categories":          templateSpec.Categories,
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
			Regex:       parameter.Regex,
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
		for key := range properties {
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
			c.Log.Error(err, "cannot get object info")
		}
		//get kind, namespace, name of object
		kind, ok := raw["kind"].(string)
		if !ok {
			c.Log.Info("Checking bindablity is failed")
			break
		}

		if strings.Contains(kind, "Service") || strings.Contains(kind, "Secret") {
			service.Bindable = true
			break
		}
	}
	return service
}

func respond(w http.ResponseWriter, statusCode int, body interface{}, log logr.Logger) {
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Error(err, "Error occurs while encoding response body")
	}
}
