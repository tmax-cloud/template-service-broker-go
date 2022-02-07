package internal

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	tmaxv1 "github.com/tmax-cloud/template-operator/api/v1"
	"github.com/tmax-cloud/template-service-broker-go/pkg/server/schemas"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("k8s api")

func GetTemplate(c client.Client, name types.NamespacedName) (*tmaxv1.Template, error) {
	template := &tmaxv1.Template{}
	if err := c.Get(context.TODO(), name, template); err != nil {
		return nil, err
	}

	return template, nil
}

func GetTemplateList(c client.Client, namespace string) (*tmaxv1.TemplateList, error) {
	templates := &tmaxv1.TemplateList{}
	if err := c.List(context.TODO(), templates, client.InNamespace(namespace)); err != nil {
		return nil, err
	}

	return templates, nil
}

func GetTemplateInstance(c client.Client, name types.NamespacedName) (*tmaxv1.TemplateInstance, error) {
	templateInstance := &tmaxv1.TemplateInstance{}
	if err := c.Get(context.TODO(), name, templateInstance); err != nil {
		return nil, err
	}

	return templateInstance, nil
}

func GetTemplateInstanceForDeprovision(c client.Client, ns string, instanceId string) (*tmaxv1.TemplateInstance, error) {
	templateInstanceList, _ := GetTemplateInstanceList(c, ns)
	var templateInstance *tmaxv1.TemplateInstance

	for _, ti := range templateInstanceList.Items {
		if ti.ObjectMeta.Annotations["instance_id"] == instanceId {
			templateInstance = &ti
			break
		}
	}

	return templateInstance, nil
}

func GetTemplateInstanceList(c client.Client, namespace string) (*tmaxv1.TemplateInstanceList, error) {
	templateInstances := &tmaxv1.TemplateInstanceList{}
	if err := c.List(context.TODO(), templateInstances, client.InNamespace(namespace)); err != nil {
		return nil, err
	}

	return templateInstances, nil
}

func CreateTemplateInstance(c client.Client, obj interface{}, namespace string,
	request schemas.ServiceInstanceProvisionRequest, instanceId string) (*tmaxv1.TemplateInstance, error) {
	var parameters []tmaxv1.ParamSpec
	template := &tmaxv1.Template{}
	clusterTemplate := &tmaxv1.ClusterTemplate{}
	template.Parameters = []tmaxv1.ParamSpec{}
	clusterTemplate.Parameters = []tmaxv1.ParamSpec{}

	name := request.Context.InstanceName
	log.Info(fmt.Sprintf("service instance name: %s", name))
	log.Info(fmt.Sprintf("service instance namespace: %s", namespace))

	labels := make(map[string]string)
	labels["serviceInstanceRef"] = request.Context.InstanceName

	annotations := make(map[string]string)
	// annotations["uid"] = request.ServiceId + "." + request.PlanId // Deprecated from TSB 0.1.4
	annotations["instance_id"] = instanceId

	// form template instance
	templateInstance := &tmaxv1.TemplateInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: annotations,
		},
	}

	switch obj.(type) {
	case *tmaxv1.Template:
		template = obj.(*tmaxv1.Template)
		templateInstance.Spec.Template = &tmaxv1.ObjectInfo{}
		templateInstance.Spec.Template.Metadata.Name = template.ObjectMeta.Name
		templateInstance.Spec.Template.Parameters = template.Parameters
		//		templateInstance.Spec.Template.Objects = template.Objects  // Deprecated since template operator 0.2.0
		parameters = templateInstance.Spec.Template.Parameters
	case *tmaxv1.ClusterTemplate:
		clusterTemplate = obj.(*tmaxv1.ClusterTemplate)
		templateInstance.Spec.ClusterTemplate = &tmaxv1.ObjectInfo{}
		templateInstance.Spec.ClusterTemplate.Metadata.Name = clusterTemplate.ObjectMeta.Name
		templateInstance.Spec.ClusterTemplate.Parameters = clusterTemplate.Parameters
		//		templateInstance.Spec.ClusterTemplate.Objects = clusterTemplate.Objects // Deprecated since template operator 0.2.0
		parameters = templateInstance.Spec.ClusterTemplate.Parameters
	}

	// check if serviceInstance has required parameters or not
	for idx, param := range parameters {
		// if param in serviceInstance
		if val, ok := request.Parameters[param.Name]; ok { // if a param was given
			if param.Required {
				if val.Type == 1 && len(val.StrVal) == 0 {
					// All parameter types filled in UI console have val.Type 1 (string type)
					return nil, fmt.Errorf("parameter %s must be included", param.Name)
				}
				//[TODO]: int type일 경우 UI에서 공란을 어떻게 받는지 확인 필요
				if val.Type == 0 && val.IntVal == 0 {
					// [TODO] : Check if it has problems
					return nil, fmt.Errorf("parameter %s must be included", param.Name)
				}
				parameters[idx].Value = val

			} else {
				parameters[idx].Value = val
			}

		} else if param.Required { // if not found && the param was required
			return nil, fmt.Errorf("parameter %s must be included", param.Name)
		}
	}

	// create template instance
	err := c.Create(context.TODO(), templateInstance)
	if err == nil { // if no error occurs
		log.Info(fmt.Sprintf("template instance name: %s is created in %s namespace", templateInstance.Name, templateInstance.Namespace))
		return templateInstance, err
	}
	if !kerrors.IsAlreadyExists(err) { // if the error is not "AlreadyExists" type
		return nil, err
	}

	// if exists, return the nil
	log.Info("The same name of template instance is already existing. Please change service instance name")
	return nil, err
}

func DeleteTemplateInstance(c client.Client, templateInstance *tmaxv1.TemplateInstance) error {
	if err := c.Delete(context.TODO(), templateInstance); err != nil {
		return err
	}
	log.Info(fmt.Sprintf("template instance name: %s is deleted in %s namespace", templateInstance.Name, templateInstance.Namespace))
	return nil
}

func GetClusterTemplate(c client.Client, name types.NamespacedName) (*tmaxv1.ClusterTemplate, error) {
	clusterTemplate := &tmaxv1.ClusterTemplate{}
	if err := c.Get(context.TODO(), name, clusterTemplate); err != nil {
		return nil, err
	}

	return clusterTemplate, nil
}

func GetClusterTemplateList(c client.Client) (*tmaxv1.ClusterTemplateList, error) {
	clusterTemplates := &tmaxv1.ClusterTemplateList{}
	if err := c.List(context.TODO(), clusterTemplates); err != nil {
		return nil, err
	}

	return clusterTemplates, nil
}

func Namespace() (string, error) {
	nsPath := "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
	if FileExists(nsPath) {
		// Running in k8s cluster
		nsBytes, err := ioutil.ReadFile(nsPath)
		if err != nil {
			return "", err
		}
		return string(nsBytes), nil
	} else {
		// Not running in k8s cluster (may be running locally)
		ns := os.Getenv("NAMESPACE")
		if ns == "" {
			ns = "default"
		}
		return ns, nil
	}
}
