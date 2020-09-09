package apis

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/jitaeyun/template-operator/pkg/apis"
	tmaxv1 "github.com/jitaeyun/template-operator/pkg/apis/tmax/v1"
	"github.com/jitaeyun/template-service-broker/internal"
	"github.com/jitaeyun/template-service-broker/pkg/server/schemas"
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
	if err := apis.AddToScheme(s); err != nil {
		log.Error(err, "cannot add Template scheme")
		respondError(w, http.StatusInternalServerError, &schemas.Error{
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
		respondError(w, http.StatusInternalServerError, &schemas.Error{
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
		respondError(w, http.StatusInternalServerError, &schemas.Error{
			Error:            "InternalServerError",
			Description:      "cannot get namespace. Check it is operated on cluster.",
			InstanceUsable:   false,
			UpdateRepeatable: false,
		})
		return
	}

	// get template to verify service class and plans exist
	template, err := internal.GetTemplate(c, types.NamespacedName{
		Namespace: ns,
		Name:      m.ServiceId,
	})

	plan := &tmaxv1.PlanSpec{}
	if err != nil || !isPlanValid(&template.TemplateSpec, m.PlanId, plan) { // cannot find template or plan
		log.Error(err, "error occurs while getting template")
		respondError(w, http.StatusBadRequest, &schemas.Error{
			Error:            "BadRequest",
			Description:      "cannot find template or plan on the namespace",
			InstanceUsable:   false,
			UpdateRepeatable: false,
		})
		return
	}

	// update template parameters using plan
	updatePlanParams(&template.TemplateSpec, plan)

	// create template instance
	if _, err = internal.CreateTemplateInstance(c, template, ns, m, serviceInstanceId); err != nil {
		log.Error(err, "error occurs while getting template")
		respondError(w, http.StatusInternalServerError, &schemas.Error{
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
	if err := apis.AddToScheme(s); err != nil {
		log.Error(err, "cannot add Template scheme")
		respondError(w, http.StatusInternalServerError, &schemas.Error{
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
		respondError(w, http.StatusInternalServerError, &schemas.Error{
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
		respondError(w, http.StatusInternalServerError, &schemas.Error{
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

	if err != nil { // cannot find template or plan
		log.Error(err, "error occurs while getting template")
		respondError(w, http.StatusBadRequest, &schemas.Error{
			Error:            "BadRequest",
			Description:      "cannot find templateInstance on the namespace",
			InstanceUsable:   false,
			UpdateRepeatable: false,
		})
		return
	}

	err = internal.DeleteTemplateInstance(c, templateInstance)
	if err != nil {
		log.Error(err, "error occurs while deleting templateInstance")
		respondError(w, http.StatusBadRequest, &schemas.Error{
			Error:            "BadRequest",
			Description:      "cannot delete templateInstance on the namespace",
			InstanceUsable:   false,
			UpdateRepeatable: false,
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
	if err := apis.AddToScheme(s); err != nil {
		log.Error(err, "cannot add Template scheme")
		respondError(w, http.StatusInternalServerError, &schemas.Error{
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
		respondError(w, http.StatusInternalServerError, &schemas.Error{
			Error:            "InternalServerError",
			Description:      "cannot connect to k8s api server",
			InstanceUsable:   false,
			UpdateRepeatable: true,
		})
		return
	}
	ns := m.Context.Namespace
	if ns == "" {
		respondError(w, http.StatusInternalServerError, &schemas.Error{
			Error:            "InternalServerError",
			Description:      "cannot get namespace. Check it is operated on cluster.",
			InstanceUsable:   false,
			UpdateRepeatable: false,
		})
		return
	}

	// get template to verify service class and plans exist
	template, err := internal.GetClusterTemplate(c, types.NamespacedName{
		Name: m.ServiceId,
	})

	plan := &tmaxv1.PlanSpec{}
	if err != nil || !isPlanValid(&template.TemplateSpec, m.PlanId, plan) { // cannot find template or plan
		log.Error(err, "error occurs while getting template")
		respondError(w, http.StatusBadRequest, &schemas.Error{
			Error:            "BadRequest",
			Description:      "cannot find template or plan on the namespace",
			InstanceUsable:   false,
			UpdateRepeatable: false,
		})
		return
	}

	// update template parameters using plan
	updatePlanParams(&template.TemplateSpec, plan)

	// create template instance
	if _, err = internal.CreateTemplateInstance(c, template, ns, m, serviceInstanceId); err != nil {
		log.Error(err, "error occurs while getting template")
		respondError(w, http.StatusInternalServerError, &schemas.Error{
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
	if err := apis.AddToScheme(s); err != nil {
		log.Error(err, "cannot add Template scheme")
		respondError(w, http.StatusInternalServerError, &schemas.Error{
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
		respondError(w, http.StatusInternalServerError, &schemas.Error{
			Error:            "InternalServerError",
			Description:      "cannot connect to k8s api server",
			InstanceUsable:   false,
			UpdateRepeatable: true,
		})
		return
	}

	name := fmt.Sprintf("%s.%s.%s", serviceId, planId, serviceInstanceId)

	// get templateinstance in all namespace
	templateInstances, err := internal.GetTemplateInstanceList(c, "")
	if err != nil { // cannot find template or plan
		log.Error(err, "error occurs while getting templateinstanceList")
		respondError(w, http.StatusBadRequest, &schemas.Error{
			Error:            "BadRequest",
			Description:      "cannot find templateInstanceList on the all namespace",
			InstanceUsable:   false,
			UpdateRepeatable: false,
		})
		return
	}

	deleteIdx := -1
	for idx, template := range templateInstances.Items {
		if template.Name == name {
			deleteIdx = idx
			break
		}
	}

	if deleteIdx < 0 {
		log.Error(err, "cannot find templateinstance")
		respondError(w, http.StatusBadRequest, &schemas.Error{
			Error:            "BadRequest",
			Description:      "cannot delete templateInstance on the namespace",
			InstanceUsable:   false,
			UpdateRepeatable: false,
		})
		return
	}

	err = internal.DeleteTemplateInstance(c, &templateInstances.Items[deleteIdx])
	if err != nil {
		log.Error(err, "error occurs while deleting templateInstance")
		respondError(w, http.StatusBadRequest, &schemas.Error{
			Error:            "BadRequest",
			Description:      "cannot delete templateInstance on the namespace",
			InstanceUsable:   false,
			UpdateRepeatable: false,
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

func respondError(w http.ResponseWriter, statusCode int, message *schemas.Error) {
	log.Error(fmt.Errorf(message.Description), "error occurred")

	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(message); err != nil {
		log.Error(err, "cannot respond")
	}
}

func isPlanValid(templateSpec *tmaxv1.TemplateSpec, planId string, plan *tmaxv1.PlanSpec) bool {
	isValid := false
	for i := range templateSpec.Plans {
		if templateSpec.Plans[i].Id == planId {
			isValid = true
			*plan = templateSpec.Plans[i]
			break
		}
	}

	return isValid
}

func updatePlanParams(templateSpec *tmaxv1.TemplateSpec, plan *tmaxv1.PlanSpec) {
	planParamMap := &plan.Schemas.ServiceInstance.Create.Parameters
	for i := range templateSpec.Parameters {
		if val, ok := (*planParamMap)[templateSpec.Parameters[i].Name]; ok {
			templateSpec.Parameters[i].Value = intstr.Parse(val)
		}
	}
}
