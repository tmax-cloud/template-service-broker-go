package apis

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	tmaxv1 "github.com/tmax-cloud/template-operator/api/v1"
	"github.com/tmax-cloud/template-service-broker-go/internal"
	"github.com/tmax-cloud/template-service-broker-go/pkg/server/schemas"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("provision")

func ProvisionServiceInstance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var m schemas.ServiceInstanceProvisionRequest

	// get body
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		log.Error(err, "error occurs while decoding service instance body")
		return
	}

	// extract variables
	serviceInstanceId := mux.Vars(r)["instanceId"]
	//instanceName := m.Context.InstanceName

	// initialize client
	s := scheme.Scheme
	if err := tmaxv1.AddToScheme(s); err != nil {
		log.Error(err, "cannot add Template scheme")
		respond(w, http.StatusInternalServerError, &schemas.Error{
			Error:            "InternalServerError",
			Description:      "cannot add Template scheme",
			InstanceUsable:   false,
			UpdateRepeatable: true,
		})
		return
	}
	c, err := internal.Client(client.Options{Scheme: s})
	if err != nil {
		log.Error(err, "cannot connect to k8s api server")
		respond(w, http.StatusInternalServerError, &schemas.Error{
			Error:            "InternalServerError",
			Description:      "cannot connect to k8s api server",
			InstanceUsable:   false,
			UpdateRepeatable: true,
		})
		return
	}
	ns, err := internal.Namespace()
	if err != nil {
		log.Error(err, "error occurs while getting namespace")
		respond(w, http.StatusInternalServerError, &schemas.Error{
			Error:            "InternalServerError",
			Description:      "cannot get namespace. Check it is operated on cluster.",
			InstanceUsable:   false,
			UpdateRepeatable: false,
		})
		return
	}

	// get template to verify service class and plans exist
	templates, err := internal.GetTemplateList(c, ns)
	if err != nil {
		log.Error(err, "error occurs while getting templateList")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:            "BadRequest",
			Description:      fmt.Sprintf("cannot find templateList on the %s namespace", ns),
			InstanceUsable:   false,
			UpdateRepeatable: false,
		})
		return
	}

	var template *tmaxv1.Template
	for _, tp := range templates.Items {
		if m.ServiceId == string(tp.UID) {
			template = &tp
			break
		}
	}

	if template == nil {
		log.Error(err, "error occurs while getting template")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:            "BadRequest",
			Description:      fmt.Sprintf("cannot find template %s on the %s namespace", m.ServiceId, ns),
			InstanceUsable:   false,
			UpdateRepeatable: false,
		})
		return
	}

	plan := &tmaxv1.PlanSpec{}
	if !isPlanValid(&template.TemplateSpec, m.PlanId, plan, string(template.UID)) {
		log.Error(err, "error occurs while validating plan")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:            "BadRequest",
			Description:      "cannot find plan",
			InstanceUsable:   false,
			UpdateRepeatable: false,
		})
		return
	}

	// update template parameters using plan
	updatePlanParams(&m, plan)

	// create template instance
	if _, err = internal.CreateTemplateInstance(c, template, ns, m, serviceInstanceId); err != nil {
		log.Error(err, "error occurs while creating template instance")
		respond(w, http.StatusInternalServerError, &schemas.Error{
			Error:            "InternalServerError",
			Description:      "cannot create template instance",
			InstanceUsable:   false,
			UpdateRepeatable: true,
		})
		return
	}

	response := schemas.ServiceInstanceProvisionResponse{}
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error(err, "cannot response")
		return
	}
}

func DeprovisionServiceInstance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	query, _ := url.ParseQuery(r.URL.RawQuery)

	// extract variables
	serviceInstanceId := mux.Vars(r)["instanceId"]
	serviceId := query["service_id"][0]
	planId := query["plan_id"][0]

	// initialize client
	s := scheme.Scheme
	if err := tmaxv1.AddToScheme(s); err != nil {
		log.Error(err, "cannot add Template scheme")
		respond(w, http.StatusInternalServerError, &schemas.Error{
			Error:            "InternalServerError",
			Description:      "cannot add Template scheme",
			InstanceUsable:   false,
			UpdateRepeatable: true,
		})
		return
	}
	c, err := internal.Client(client.Options{Scheme: s})
	if err != nil {
		log.Error(err, "cannot connect to k8s api server")
		respond(w, http.StatusInternalServerError, &schemas.Error{
			Error:            "InternalServerError",
			Description:      "cannot connect to k8s api server",
			InstanceUsable:   false,
			UpdateRepeatable: true,
		})
		return
	}

	ns, err := internal.Namespace()
	if err != nil {
		log.Error(err, "error occurs while getting namespace")
		respond(w, http.StatusInternalServerError, &schemas.Error{
			Error:            "InternalServerError",
			Description:      "cannot get namespace. Check it is operated on cluster.",
			InstanceUsable:   false,
			UpdateRepeatable: false,
		})
		return
	}

	name := fmt.Sprintf("%s.%s.%s", serviceId, planId, serviceInstanceId)

	// get template to verify service class and plans exist
	templateInstance, err := internal.GetTemplateInstance(c, types.NamespacedName{
		Namespace: ns,
		Name:      name,
	})

	// If there is no templateinstance, deprovision is complete because there is no instance to delete.
	if err != nil {
		log.Info(fmt.Sprintf("TemplateInstance %s does not exist", name))
		respond(w, http.StatusOK, schemas.ServiceInstanceProvisionResponse{})
		return
	}

	if err := internal.DeleteTemplateInstance(c, templateInstance); err != nil {
		log.Error(err, "error occurs while deleting templateInstance")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:            "BadRequest",
			Description:      "cannot delete templateInstance on the namespace",
			InstanceUsable:   false,
			UpdateRepeatable: false,
		})
		return
	}

	respond(w, http.StatusOK, schemas.ServiceInstanceProvisionResponse{})
}

func ClusterProvisionServiceInstance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var m schemas.ServiceInstanceProvisionRequest

	// get body
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		log.Error(err, "error occurs while decoding service instance body")
		return
	}

	// extract variables
	serviceInstanceId := mux.Vars(r)["instanceId"]

	// initialize client
	s := scheme.Scheme
	if err := tmaxv1.AddToScheme(s); err != nil {
		log.Error(err, "cannot add Template scheme")
		respond(w, http.StatusInternalServerError, &schemas.Error{
			Error:            "InternalServerError",
			Description:      "cannot add Template scheme",
			InstanceUsable:   false,
			UpdateRepeatable: true,
		})
		return
	}
	c, err := internal.Client(client.Options{Scheme: s})
	if err != nil {
		log.Error(err, "cannot connect to k8s api server")
		respond(w, http.StatusInternalServerError, &schemas.Error{
			Error:            "InternalServerError",
			Description:      "cannot connect to k8s api server",
			InstanceUsable:   false,
			UpdateRepeatable: true,
		})
		return
	}
	ns := m.Context.Namespace
	if ns == "" {
		respond(w, http.StatusInternalServerError, &schemas.Error{
			Error:            "InternalServerError",
			Description:      "cannot get namespace. Check it is operated on cluster.",
			InstanceUsable:   false,
			UpdateRepeatable: false,
		})
		return
	}

	// get template to verify service class and plans exist
	templates, err := internal.GetClusterTemplateList(c)
	if err != nil {
		log.Error(err, "error occurs while getting templateList")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:            "BadRequest",
			Description:      "cannot find ClusterTemplateList",
			InstanceUsable:   false,
			UpdateRepeatable: false,
		})
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
		log.Error(err, "error occurs while getting template")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:            "BadRequest",
			Description:      fmt.Sprintf("cannot find ClusterTemplate %s", m.ServiceId),
			InstanceUsable:   false,
			UpdateRepeatable: false,
		})
		return
	}

	plan := &tmaxv1.PlanSpec{}
	if !isPlanValid(&template.TemplateSpec, m.PlanId, plan, string(template.UID)) {
		log.Error(err, "error occurs while validating plan")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:            "BadRequest",
			Description:      fmt.Sprintf("cannot find plan %s", m.PlanId),
			InstanceUsable:   false,
			UpdateRepeatable: false,
		})
		return
	}
	// update template parameters using plan
	updatePlanParams(&m, plan)

	// create template instance
	if _, err = internal.CreateTemplateInstance(c, template, ns, m, serviceInstanceId); err != nil {
		log.Error(err, "error occurs while getting template")
		respond(w, http.StatusInternalServerError, &schemas.Error{
			Error:            "InternalServerError",
			Description:      "cannot create template instance",
			InstanceUsable:   false,
			UpdateRepeatable: true,
		})
		return
	}

	response := schemas.ServiceInstanceProvisionResponse{}
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error(err, "cannot response")
		return
	}
}

func ClusterDeprovisionServiceInstance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	query, _ := url.ParseQuery(r.URL.RawQuery)

	// extract variables
	serviceInstanceId := mux.Vars(r)["instanceId"]
	serviceId := query["service_id"][0]
	planId := query["plan_id"][0]

	// initialize client
	s := scheme.Scheme
	if err := tmaxv1.AddToScheme(s); err != nil {
		log.Error(err, "cannot add Template scheme")
		respond(w, http.StatusInternalServerError, &schemas.Error{
			Error:            "InternalServerError",
			Description:      "cannot add Template scheme",
			InstanceUsable:   false,
			UpdateRepeatable: true,
		})
		return
	}
	c, err := internal.Client(client.Options{Scheme: s})
	if err != nil {
		log.Error(err, "cannot connect to k8s api server")
		respond(w, http.StatusInternalServerError, &schemas.Error{
			Error:            "InternalServerError",
			Description:      "cannot connect to k8s api server",
			InstanceUsable:   false,
			UpdateRepeatable: true,
		})
		return
	}

	name := fmt.Sprintf("%s.%s.%s", serviceId, planId, serviceInstanceId)

	// get templateinstance in all namespace
	templateInstanceList, err := internal.GetTemplateInstanceList(c, "")
	if err != nil {
		log.Error(err, "error occurs while getting templateinstanceList")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:            "BadRequest",
			Description:      "cannot find templateInstanceList on the all namespace",
			InstanceUsable:   false,
			UpdateRepeatable: false,
		})
		return
	}

	var templateInstance *tmaxv1.TemplateInstance
	for _, ti := range templateInstanceList.Items {
		if ti.Name == name {
			templateInstance = &ti
			break
		}
	}

	// If there is no templateinstance, deprovision is complete because there is no instance to delete.
	if templateInstance == nil {
		log.Info(fmt.Sprintf("TemplateInstance %s does not exist", name))
		respond(w, http.StatusOK, schemas.ServiceInstanceProvisionResponse{})
		return
	}

	if err := internal.DeleteTemplateInstance(c, templateInstance); err != nil {
		log.Error(err, "error occurs while deleting templateInstance")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:            "BadRequest",
			Description:      "cannot delete templateInstance on the namespace",
			InstanceUsable:   false,
			UpdateRepeatable: false,
		})
		return
	}

	respond(w, http.StatusOK, schemas.ServiceInstanceProvisionResponse{})
}

func respond(w http.ResponseWriter, statusCode int, body interface{}) {
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Error(err, "Error occurs while encoding response body")
	}
}

func isPlanValid(templateSpec *tmaxv1.TemplateSpec, planId string, plan *tmaxv1.PlanSpec, uid string) bool {
	tokIdx := strings.LastIndex(planId, "-")
	planUid := planId[:tokIdx]
	planIdx := planId[tokIdx+1:]
	if len(templateSpec.Plans) == 0 {
		return true
	}
	if planUid != uid {
		return false
	}
	n, _ := strconv.Atoi(planIdx)
	if n < len(templateSpec.Plans) {
		*plan = templateSpec.Plans[n]
		return true
	}
	return false
}

func updatePlanParams(request *schemas.ServiceInstanceProvisionRequest, plan *tmaxv1.PlanSpec) {
	if len(request.Parameters) == 0 {
		request.Parameters = make(map[string]intstr.IntOrString)
	}
	for key, val := range plan.Schemas.ServiceInstance.Create.Parameters {
		request.Parameters[key] = val
	}
}
