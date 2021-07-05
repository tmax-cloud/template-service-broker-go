package apis

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	tmaxv1 "github.com/tmax-cloud/template-operator/api/v1"
	"github.com/tmax-cloud/template-service-broker-go/internal"
	"github.com/tmax-cloud/template-service-broker-go/pkg/server/schemas"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Provision struct {
	client.Client
	Log logr.Logger
}

func (p *Provision) ProvisionServiceInstance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var m schemas.ServiceInstanceProvisionRequest

	// get body
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		p.Log.Error(err, "error occurs while decoding service instance body")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:            "Bad Request",
			Description:      "Cannot decode service instance request body",
			InstanceUsable:   false,
			UpdateRepeatable: false,
		}, p.Log)
		return
	}

	ns, err := internal.Namespace()
	if err != nil {
		p.Log.Error(err, "error occurs while getting namespace")
		respond(w, http.StatusInternalServerError, &schemas.Error{
			Error:            "InternalServerError",
			Description:      "cannot get namespace. Check it is operated on cluster.",
			InstanceUsable:   false,
			UpdateRepeatable: false,
		}, p.Log)
		return
	}

	// get template to verify service class and plans exist
	templateList, err := internal.GetTemplateList(p.Client, ns)
	if err != nil {
		p.Log.Error(err, "error occurs while getting templateList")
		respond(w, http.StatusInternalServerError, &schemas.Error{
			Error:            "InternalServerError",
			Description:      fmt.Sprintf("cannot find templateList on the %s namespace", ns),
			InstanceUsable:   false,
			UpdateRepeatable: false,
		}, p.Log)
		return
	}

	var template *tmaxv1.Template
	for _, tp := range templateList.Items {
		if m.ServiceId == string(tp.UID) {
			template = &tp
			break
		}
	}

	if template == nil {
		p.Log.Error(err, "error occurs while getting template")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:            "BadRequest",
			Description:      fmt.Sprintf("cannot find template %s on the %s namespace", m.ServiceId, ns),
			InstanceUsable:   false,
			UpdateRepeatable: false,
		}, p.Log)
		return
	}

	// update template parameters using plan
	if err := updatePlanParams(&m, template.TemplateSpec, string(template.UID)); err != nil {
		p.Log.Error(err, "error occurs while reflecting plan parameter")
		respond(w, http.StatusInternalServerError, &schemas.Error{
			Error:            "InternalServerError",
			Description:      "plan is invalid",
			InstanceUsable:   false,
			UpdateRepeatable: true,
		}, p.Log)
		return
	}

	// create template instance
	if _, err = internal.CreateTemplateInstance(p.Client, template, ns, m); err != nil {
		p.Log.Error(err, "error occurs while creating template instance")
		respond(w, http.StatusInternalServerError, &schemas.Error{
			Error:            "InternalServerError",
			Description:      "cannot create template instance",
			InstanceUsable:   false,
			UpdateRepeatable: true,
		}, p.Log)
		return
	}

	respond(w, http.StatusOK, schemas.ServiceInstanceProvisionResponse{}, p.Log)
}

func (p *Provision) DeprovisionServiceInstance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	query, _ := url.ParseQuery(r.URL.RawQuery)

	// extract variables
	serviceId := query["service_id"][0]
	planId := query["plan_id"][0]

	ns, err := internal.Namespace()
	if err != nil {
		p.Log.Error(err, "error occurs while getting namespace")
		respond(w, http.StatusInternalServerError, &schemas.Error{
			Error:            "InternalServerError",
			Description:      "cannot get namespace. Check it is operated on cluster.",
			InstanceUsable:   false,
			UpdateRepeatable: false,
		}, p.Log)
		return
	}

	templateInstance, err := internal.GetTemplateInstanceForDeprovision(p.Client, serviceId, planId, ns)
	// If there is no templateinstance, deprovision is complete because there is no instance to delete.
	if err != nil {
		p.Log.Info("TemplateInstance does not exist")
		respond(w, http.StatusOK, schemas.ServiceInstanceProvisionResponse{}, p.Log)
		return
	}

	if err := internal.DeleteTemplateInstance(p.Client, templateInstance); err != nil {
		p.Log.Error(err, "error occurs while deleting templateInstance")
		respond(w, http.StatusInternalServerError, &schemas.Error{
			Error:            "InternalServerError",
			Description:      "Error occurs while delete templateInstance",
			InstanceUsable:   false,
			UpdateRepeatable: false,
		}, p.Log)
		return
	}

	respond(w, http.StatusOK, schemas.ServiceInstanceProvisionResponse{}, p.Log)
}

func (p *Provision) ClusterProvisionServiceInstance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var m schemas.ServiceInstanceProvisionRequest

	// get body
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		p.Log.Error(err, "error occurs while decoding service instance body")
		return
	}

	templates, err := internal.GetClusterTemplateList(p.Client)
	if err != nil {
		p.Log.Error(err, "error occurs while getting templateList")
		respond(w, http.StatusInternalServerError, &schemas.Error{
			Error:            "InternalServerError",
			Description:      "cannot find ClusterTemplateList",
			InstanceUsable:   false,
			UpdateRepeatable: false,
		}, p.Log)
		return
	}

	var template *tmaxv1.ClusterTemplate
	for _, tp := range templates.Items {
		if m.ServiceId == string(tp.UID) {
			template = &tp
			break
		}
	}

	if template == nil {
		p.Log.Error(err, "error occurs while getting template")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:            "BadRequest",
			Description:      fmt.Sprintf("cannot find ClusterTemplate %s", m.ServiceId),
			InstanceUsable:   false,
			UpdateRepeatable: false,
		}, p.Log)
		return
	}

	// update template parameters using plan
	updatePlanParams(&m, template.TemplateSpec, string(template.UID))

	// create template instance
	if _, err = internal.CreateTemplateInstance(p.Client, template, m.Context.Namespace, m); err != nil {
		p.Log.Error(err, "error occurs while creating template instance")
		respond(w, http.StatusInternalServerError, &schemas.Error{
			Error:            "InternalServerError",
			Description:      "cannot create template instance",
			InstanceUsable:   false,
			UpdateRepeatable: true,
		}, p.Log)
		return
	}

	respond(w, http.StatusOK, schemas.ServiceInstanceProvisionResponse{}, p.Log)
}

func (p *Provision) ClusterDeprovisionServiceInstance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	query, _ := url.ParseQuery(r.URL.RawQuery)

	// extract variables
	serviceId := query["service_id"][0]
	planId := query["plan_id"][0]

	// get templateinstance in all namespace
	templateInstanceList, err := internal.GetTemplateInstanceList(p.Client, "")
	if err != nil {
		p.Log.Error(err, "error occurs while getting templateinstanceList")
		respond(w, http.StatusInternalServerError, &schemas.Error{
			Error:            "InternalServerError",
			Description:      "cannot find templateInstanceList",
			InstanceUsable:   false,
			UpdateRepeatable: false,
		}, p.Log)
		return
	}

	var templateInstance *tmaxv1.TemplateInstance
	for _, ti := range templateInstanceList.Items {
		if ti.ObjectMeta.Annotations["uid"] == serviceId+"."+planId {
			templateInstance = &ti
			break
		}
	}

	// If there is no templateinstance, deprovision is complete because there is no instance to delete.
	if templateInstance == nil {
		p.Log.Info("TemplateInstance does not exist")
		respond(w, http.StatusOK, schemas.ServiceInstanceProvisionResponse{}, p.Log)
		return
	}

	if err := internal.DeleteTemplateInstance(p.Client, templateInstance); err != nil {
		p.Log.Error(err, "error occurs while deleting templateInstance")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:            "BadRequest",
			Description:      "cannot delete templateInstance on the namespace",
			InstanceUsable:   false,
			UpdateRepeatable: false,
		}, p.Log)
		return
	}

	respond(w, http.StatusOK, schemas.ServiceInstanceProvisionResponse{}, p.Log)
}

func updatePlanParams(request *schemas.ServiceInstanceProvisionRequest, templateSpec tmaxv1.TemplateSpec, templateUid string) error {
	// check if plan valid
	tokIdx := strings.LastIndex(request.PlanId, "-")
	planUid := request.PlanId[:tokIdx]
	planIdx := request.PlanId[tokIdx+1:]

	if planUid != templateUid {
		return fmt.Errorf("plan has invalid uid")
	}

	plan := tmaxv1.PlanSpec{}
	if len(templateSpec.Plans) != 0 {
		idx, _ := strconv.Atoi(planIdx)
		if idx >= len(templateSpec.Plans) {
			return fmt.Errorf("plan has invalid index")
		}
		plan = templateSpec.Plans[idx]
	}

	// reflect plan parameter
	if len(request.Parameters) == 0 {
		request.Parameters = make(map[string]intstr.IntOrString)
	}
	for key, val := range plan.Schemas.ServiceInstance.Create.Parameters {
		request.Parameters[key] = val
	}
	return nil
}
