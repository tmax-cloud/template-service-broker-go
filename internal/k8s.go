package internal

import (
	"context"
	"io/ioutil"
	"os"

	tmaxv1 "github.com/youngind/hypercloud-operator/pkg/apis/tmax/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetTemplateInstance(c client.Client, name types.NamespacedName) (*tmaxv1.TemplateInstance, error) {
	templateInstance := &tmaxv1.TemplateInstance{}
	if err := c.Get(context.TODO(), name, templateInstance); err != nil {
		return nil, err
	}

	return templateInstance, nil
}

func GetTemplateInstanceList(c client.Client) (*tmaxv1.TemplateInstanceList, error) {

	templateInstanceList := &tmaxv1.TemplateInstanceList{}
	if err := c.List(context.TODO(), templateInstanceList); err != nil {
		return nil, err
	}

	return templateInstanceList, nil
}

func GetTemplate(c client.Client, name types.NamespacedName) (*tmaxv1.Template, error) {
	template := &tmaxv1.Template{}
	if err := c.Get(context.TODO(), name, template); err != nil {
		return nil, err
	}

	return template, nil
}

func GetTemplateList(c client.Client) (*tmaxv1.TemplateList, error) {

	templates := &tmaxv1.TemplateList{}
	if err := c.List(context.TODO(), templates); err != nil {
		return nil, err
	}

	return templates, nil
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
