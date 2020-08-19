package apis

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jwkim1993/template-service-broker/internal"
	"github.com/jwkim1993/template-service-broker/pkg/server/schemas"
	"github.com/youngind/hypercloud-operator/pkg/apis"
	tmaxv1 "github.com/youngind/hypercloud-operator/pkg/apis/tmax/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/scheme"
	"net/http"
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
	if err != nil || !isPlanValid(template, m.PlanId, plan) { // cannot find template or plan
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
	updatePlanParams(template, plan)

	// create template instance
	if _, err = internal.CreateTemplateInstance(c, template, m, serviceInstanceId); err != nil {
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
	var m schemas.ServiceInstanceProvisionRequest
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		log.Error(err, "error occurs while decoding service instance body")
		return
	}

	//c, err := internal.Client(client.Options{})
	//if err != nil {
	//	log.Error(err, "cannot connect to k8s api server")
	//	return
	//}

	// get template instance to delete

	// delete the template instance
}

func respondError(w http.ResponseWriter, statusCode int, message *schemas.Error) {
	log.Error(fmt.Errorf(message.Description), "error occurred")

	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(message); err != nil {
		log.Error(err, "cannot respond")
	}
}

func isPlanValid(template *tmaxv1.Template, planId string, plan *tmaxv1.PlanSpec) bool {
	isValid := false
	for i := range template.Plans {
		if template.Plans[i].Id == planId {
			isValid = true
			*plan = template.Plans[i]
			break
		}
	}

	return isValid
}

func updatePlanParams(template *tmaxv1.Template, plan *tmaxv1.PlanSpec) {
	planParamMap := &plan.Schemas.ServiceInstance.Create.Parameters
	for i := range template.Parameters {
		if val, ok := (*planParamMap)[template.Parameters[i].Name]; ok {
			template.Parameters[i].Value = intstr.Parse(val)
		}
	}
}
