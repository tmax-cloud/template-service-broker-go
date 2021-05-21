package apis

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/tmax-cloud/template-service-broker-go/internal"
	"github.com/tmax-cloud/template-service-broker-go/pkg/server/schemas"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var logBind = logf.Log.WithName("binding")

func BindingServiceInstance(w http.ResponseWriter, r *http.Request) {
	var m schemas.ServiceBindingRequest

	//set reponse
	response := &schemas.ServiceBindingResponse{}
	response.Credentials = make(map[string]interface{})
	w.Header().Set("Content-Type", "application/json")

	//get request body
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		logBind.Error(err, "error occurs while decoding service binding body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//get url param
	vars := mux.Vars(r)

	//get templateinstance name & namespace
	instanceName := m.ServiceId + "." + m.PlanId + "." + vars["instance_id"]
	instanceNameSpace, err := internal.Namespace()
	if err != nil {
		logBind.Error(err, "cannot get namespace")
		respond(w, http.StatusInternalServerError, &schemas.Error{
			Error:            "InternalServerError",
			Description:      "cannot get namespace. Check it is operated on cluster.",
			InstanceUsable:   false,
			UpdateRepeatable: false,
		})
	}

	//add template & template instance schema
	s := scheme.Scheme
	internal.AddKnownTypes(s)
	SchemeBuilder := runtime.NewSchemeBuilder()
	if err := SchemeBuilder.AddToScheme(s); err != nil {
		logBind.Error(err, "cannot add Template/Templateinstance scheme")
		respond(w, http.StatusInternalServerError, &schemas.Error{
			Error:            "InternalServerError",
			Description:      "cannot add Template/Templateinstance scheme",
			InstanceUsable:   false,
			UpdateRepeatable: true,
		})
		return
	}

	// connect k8s client
	c, err := internal.Client(client.Options{Scheme: s})
	if err != nil {
		logBind.Error(err, "cannot connect k8s api server")
		respond(w, http.StatusInternalServerError, &schemas.Error{
			Error:            "InternalServerError",
			Description:      "cannot connect to k8s api server",
			InstanceUsable:   false,
			UpdateRepeatable: true,
		})
		return
	}

	// get templateinstance info
	templateInstance, err := internal.GetTemplateInstance(c, types.NamespacedName{Name: instanceName, Namespace: instanceNameSpace})
	if err != nil {
		logBind.Error(err, "cannot get templateinstance info")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:            "BadRequest",
			Description:      "cannot find templateinstance on the namespace",
			InstanceUsable:   false,
			UpdateRepeatable: false,
		})
		return
	}

	// parse object info in template info
	for _, object := range templateInstance.Spec.Template.Objects {
		var raw map[string]interface{}
		if err := json.Unmarshal(object.Raw, &raw); err != nil {
			logBind.Error(err, "cannot get object info")
			respond(w, http.StatusBadRequest, &schemas.Error{
				Error:            "BadRequest",
				Description:      "cannot find object info of template",
				InstanceUsable:   false,
				UpdateRepeatable: false,
			})
			return
		}

		//get kind, namespace, name of object
		kind := raw["kind"].(string)
		name := raw["metadata"].(map[string]interface{})["name"].(string)
		namespace := instanceNameSpace

		if strings.Compare(kind, "Service") == 0 {
			//set endpoint in case of service
			service := &corev1.Service{}
			if err := c.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: name}, service); err != nil {
				logBind.Error(err, "error occurs while getting service info")
				respond(w, http.StatusBadRequest, &schemas.Error{
					Error:            "BadRequest",
					Description:      "cannot get service info of template",
					InstanceUsable:   false,
					UpdateRepeatable: false,
				})
				return
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

		} else if strings.Compare(kind, "Secret") == 0 {
			//set credentials in case of secret
			secret := &corev1.Secret{}
			if err := c.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: name}, secret); err != nil {
				logBind.Error(err, "error occurs while getting secret info")
				respond(w, http.StatusBadRequest, &schemas.Error{
					Error:            "BadRequest",
					Description:      "cannot get secret info of template",
					InstanceUsable:   false,
					UpdateRepeatable: false,
				})
				return
			}
			for key, val := range secret.Data {
				response.Credentials[key] = string(val)
			}
		}
	}
	//set credential if not empty Endpoints
	if len(response.Endpoints) != 0 {
		response.Credentials["endpoints"] = response.Endpoints
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func UnBindingServiceInstance(w http.ResponseWriter, r *http.Request) {
	response := &schemas.AsyncOperation{}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func ClusterBindingServiceInstance(w http.ResponseWriter, r *http.Request) {
	var m schemas.ServiceBindingRequest

	//set reponse
	response := &schemas.ServiceBindingResponse{}
	response.Credentials = make(map[string]interface{})
	w.Header().Set("Content-Type", "application/json")

	//get request body
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		logBind.Error(err, "error occurs while decoding service binding body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//get url param
	vars := mux.Vars(r)

	//get templateinstance name & namespace
	instanceName := m.ServiceId + "." + m.PlanId + "." + vars["instance_id"]
	instanceNameSpace := m.Context.Namespace

	if instanceNameSpace == "" {
		logBind.Info("cannot get instanceNamespace")
		respond(w, http.StatusInternalServerError, &schemas.Error{
			Error:            "InternalServerError",
			Description:      "cannot get namespace. Check it is operated on cluster.",
			InstanceUsable:   false,
			UpdateRepeatable: false,
		})
		return
	}

	//add template & template instance schema
	s := scheme.Scheme
	internal.AddKnownTypes(s)
	SchemeBuilder := runtime.NewSchemeBuilder()
	if err := SchemeBuilder.AddToScheme(s); err != nil {
		logBind.Error(err, "cannot add Template/Templateinstance scheme")
		respond(w, http.StatusInternalServerError, &schemas.Error{
			Error:            "InternalServerError",
			Description:      "cannot add Template/Templateinstance scheme",
			InstanceUsable:   false,
			UpdateRepeatable: true,
		})
		return
	}

	// connect k8s client
	c, err := internal.Client(client.Options{Scheme: s})
	if err != nil {
		logBind.Error(err, "cannot connect k8s api server")
		respond(w, http.StatusInternalServerError, &schemas.Error{
			Error:            "InternalServerError",
			Description:      "cannot connect to k8s api server",
			InstanceUsable:   false,
			UpdateRepeatable: true,
		})
		return
	}

	// get templateinstance info
	templateInstance, err := internal.GetTemplateInstance(c, types.NamespacedName{Name: instanceName, Namespace: instanceNameSpace})
	if err != nil {
		logBind.Error(err, "cannot get templateinstance info")
		respond(w, http.StatusBadRequest, &schemas.Error{
			Error:            "BadRequest",
			Description:      "cannot find templateinstance on the namespace",
			InstanceUsable:   false,
			UpdateRepeatable: false,
		})
		return
	}

	// parse object info in template info
	for _, object := range templateInstance.Spec.ClusterTemplate.Objects {
		var raw map[string]interface{}
		if err := json.Unmarshal(object.Raw, &raw); err != nil {
			logBind.Error(err, "cannot get object info")
			respond(w, http.StatusBadRequest, &schemas.Error{
				Error:            "BadRequest",
				Description:      "cannot find object info of template",
				InstanceUsable:   false,
				UpdateRepeatable: false,
			})
			return
		}

		//get kind, namespace, name of object
		kind := raw["kind"].(string)
		name := raw["metadata"].(map[string]interface{})["name"].(string)
		namespace := instanceNameSpace

		if strings.Compare(kind, "Service") == 0 {
			//set endpoint in case of service
			service := &corev1.Service{}
			if err := c.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: name}, service); err != nil {
				logBind.Error(err, "error occurs while getting service info")
				respond(w, http.StatusBadRequest, &schemas.Error{
					Error:            "BadRequest",
					Description:      "cannot get service info of template",
					InstanceUsable:   false,
					UpdateRepeatable: false,
				})
				return
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

		} else if strings.Compare(kind, "Secret") == 0 {
			//set credentials in case of secret
			secret := &corev1.Secret{}
			if err := c.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: name}, secret); err != nil {
				logBind.Error(err, "error occurs while getting secret info")
				respond(w, http.StatusBadRequest, &schemas.Error{
					Error:            "BadRequest",
					Description:      "cannot get secret info of template",
					InstanceUsable:   false,
					UpdateRepeatable: false,
				})
				return
			}
			for key, val := range secret.Data {
				response.Credentials[key] = string(val)
			}
		}
	}
	//set credential if not empty Endpoints
	if len(response.Endpoints) != 0 {
		response.Credentials["endpoints"] = response.Endpoints
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
