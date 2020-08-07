package internal

import (
	tmaxv1 "github.com/youngind/hypercloud-operator/pkg/apis/tmax/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var SchemeGroupVersion = schema.GroupVersion{Group: "tmax.io", Version: "v1"}

func Client(options client.Options) (client.Client, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}
	c, err := client.New(cfg, options)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func AddKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&tmaxv1.TemplateInstance{},
		&tmaxv1.TemplateInstanceList{},
	)
	scheme.AddKnownTypes(SchemeGroupVersion,
		&tmaxv1.Template{},
		&tmaxv1.TemplateList{},
	)
	v1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
