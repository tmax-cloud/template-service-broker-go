package internal

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	tmaxv1 "github.com/jwkim1993/hypercloud-operator/pkg/apis/tmax/v1"
	"github.com/jwkim1993/template-service-broker/pkg/server/schemas"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
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

func GetTemplateInstanceList(c client.Client, namespace string) (*tmaxv1.TemplateInstanceList, error) {
	templateInstances := &tmaxv1.TemplateInstanceList{}
	if err := c.List(context.TODO(), templateInstances, client.InNamespace(namespace)); err != nil {
		return nil, err
	}

	return templateInstances, nil
}

func CreateTemplateInstance(c client.Client, obj interface{}, namespace string,
	request schemas.ServiceInstanceProvisionRequest, serviceInstanceId string) (*tmaxv1.TemplateInstance, error) {
	var parameters []tmaxv1.ParamSpec
	template := &tmaxv1.Template{}
	clusterTemplate := &tmaxv1.ClusterTemplate{}
	template.Parameters = []tmaxv1.ParamSpec{}
	clusterTemplate.Parameters = []tmaxv1.ParamSpec{}

	switch obj.(type) {
	case *tmaxv1.Template:
		template = obj.(*tmaxv1.Template)
		parameters = template.Parameters[0:]
	case *tmaxv1.ClusterTemplate:
		clusterTemplate = obj.(*tmaxv1.ClusterTemplate)
		parameters = clusterTemplate.Parameters[0:]
	}

	for idx, param := range parameters {
		// if param in plan
		if val, ok := request.Parameters[param.Name]; ok { // if a param was given
			parameters[idx].Value = intstr.Parse(val)
		} else if param.Required { // if not found && the param was required
			return nil, errors.New(fmt.Sprintf("parameter %s must be included", param.Name))
		}
	}

	name := fmt.Sprintf("%s.%s.%s", request.ServiceId, request.PlanId, serviceInstanceId)
	log.Info(fmt.Sprintf("service instance name: %s", name))

	// form template instance
	templateInstance := &tmaxv1.TemplateInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: tmaxv1.TemplateInstanceSpec{
			Template:        *template,
			ClusterTemplate: *clusterTemplate,
		},
	}

	// create template instance
	err := c.Create(context.TODO(), templateInstance)
	if err == nil { // if no error occurs
		return templateInstance, err
	}
	if !kerrors.IsAlreadyExists(err) { // if the error is not "AlreadyExists" type
		return nil, err
	}

	// if exists, return the existing template instance
	log.Info("the template instance already existed. return existing template")
	existingTemplateInstance, err := GetTemplateInstance(c, types.NamespacedName{
		Namespace: templateInstance.Namespace,
		Name:      templateInstance.Name,
	})
	if err != nil {
		return nil, err
	}

	return existingTemplateInstance, nil
}

func DeleteTemplateInstance(c client.Client, templateInstance *tmaxv1.TemplateInstance) error {
	if err := c.Delete(context.TODO(), templateInstance); err != nil {
		return err
	}
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
