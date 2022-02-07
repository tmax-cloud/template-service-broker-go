package apis

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-logr/logr"
	"github.com/tmax-cloud/template-service-broker-go/internal"
	"github.com/tmax-cloud/template-service-broker-go/pkg/server/schemas"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Binding struct {
	client.Client
	Log logr.Logger
}

// [TODO]: instance.status.(cluster)template 으로 변경해야 하는지 확인필요
func (b *Binding) BindingServiceInstance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var m schemas.ServiceBindingRequest

	//get request body
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		b.Log.Error(err, "error occurs while decoding service binding body")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:            "Bad Request",
			Description:      "Cannot decode service binding request body",
			InstanceUsable:   false,
			UpdateRepeatable: false,
		}, b.Log)
		return
	}

	//get templateinstance name & namespace
	instanceName := m.Context.InstanceName
	instanceNameSpace, err := internal.Namespace()
	if err != nil {
		b.Log.Error(err, "cannot get namespace")
		respond(w, http.StatusInternalServerError, &schemas.Error{
			Error:            "InternalServerError",
			Description:      "cannot get namespace. Check it is operated on cluster.",
			InstanceUsable:   false,
			UpdateRepeatable: false,
		}, b.Log)
	}

	// get templateinstance info
	templateInstance, err := internal.GetTemplateInstance(b.Client, types.NamespacedName{Name: instanceName, Namespace: instanceNameSpace})
	if err != nil {
		b.Log.Error(err, "cannot get templateinstance info")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:            "BadRequest",
			Description:      fmt.Sprintf("cannot find templateinstance on the %s namespace", instanceNameSpace),
			InstanceUsable:   false,
			UpdateRepeatable: false,
		}, b.Log)
		return
	}

	//set reponse
	response := &schemas.ServiceBindingResponse{}
	if err := b.getBindingInfo(templateInstance.Spec.Template.Objects, instanceNameSpace, response); err != nil {
		b.Log.Error(err, "Error occurs while get binding info")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:            "BadRequest",
			Description:      "Error occurs while get binding info",
			InstanceUsable:   false,
			UpdateRepeatable: false,
		}, b.Log)
		return
	}

	respond(w, http.StatusOK, response, b.Log)
}

func (b *Binding) ClusterBindingServiceInstance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var m schemas.ServiceBindingRequest

	//get request body
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		b.Log.Error(err, "error occurs while decoding service binding body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//get templateinstance name & namespace
	instanceName := m.Context.InstanceName
	instanceNameSpace := m.Context.Namespace

	// get templateinstance info
	templateInstance, err := internal.GetTemplateInstance(b.Client, types.NamespacedName{Name: instanceName, Namespace: instanceNameSpace})
	if err != nil {
		b.Log.Error(err, "cannot get templateinstance info")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:            "BadRequest",
			Description:      fmt.Sprintf("cannot find templateinstance on the %s namespace", instanceNameSpace),
			InstanceUsable:   false,
			UpdateRepeatable: false,
		}, b.Log)
		return
	}

	//set reponse
	response := &schemas.ServiceBindingResponse{}
	if err := b.getBindingInfo(templateInstance.Spec.ClusterTemplate.Objects, instanceNameSpace, response); err != nil {
		b.Log.Error(err, "Error occurs while get binding info")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:            "BadRequest",
			Description:      "Error occurs while get binding info",
			InstanceUsable:   false,
			UpdateRepeatable: false,
		}, b.Log)
		return
	}

	respond(w, http.StatusOK, response, b.Log)
}

func (b *Binding) getBindingInfo(objects []runtime.RawExtension, ns string, response *schemas.ServiceBindingResponse) error {
	response.Credentials = make(map[string]interface{})

	for _, object := range objects {
		var unmarshaledObject map[string]interface{}
		if err := json.Unmarshal(object.Raw, &unmarshaledObject); err != nil {
			b.Log.Error(err, "Error occurs while unmarshal object info")
			return err
		}

		//get kind, namespace, name of object
		kind := unmarshaledObject["kind"].(string)
		name := unmarshaledObject["metadata"].(map[string]interface{})["name"].(string)
		if kind == "Service" {
			//set endpoint in case of service
			service := &corev1.Service{}
			if err := b.Client.Get(context.TODO(), types.NamespacedName{Namespace: ns, Name: name}, service); err != nil {
				b.Log.Error(err, "error occurs while get service info")
				return err
			}
			if service.Spec.Type == corev1.ServiceTypeLoadBalancer {
				ports := []string{}
				for _, port := range service.Spec.Ports {
					ports = append(ports, strconv.FormatInt(int64(port.Port), 10))
				}
				for _, ingress := range service.Status.LoadBalancer.Ingress {
					endpoint := schemas.ServiceBindingEndpoint{
						Host:  ingress.IP,
						Ports: ports,
					}
					response.Endpoints = append(response.Endpoints, endpoint)
				}
			}
		}
		if kind == "Secret" {
			//set credentials in case of secret
			secret := &corev1.Secret{}
			if err := b.Client.Get(context.TODO(), types.NamespacedName{Namespace: ns, Name: name}, secret); err != nil {
				b.Log.Error(err, "error occurs while get secret info")
				return err
			}
			for key, val := range secret.Data {
				response.Credentials[key] = string(val)
			}
		}
	}

	//set credential if Endpoint is not empty
	if len(response.Endpoints) != 0 {
		response.Credentials["endpoints"] = response.Endpoints
	}

	return nil
}

func (b *Binding) UnBindingServiceInstance(w http.ResponseWriter, r *http.Request) {
	respond(w, http.StatusOK, schemas.ServiceInstanceProvisionResponse{}, b.Log)
}
