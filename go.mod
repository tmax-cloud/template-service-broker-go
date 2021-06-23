module github.com/tmax-cloud/template-service-broker-go

go 1.13

require (
	github.com/go-logr/logr v0.1.0
	github.com/gorilla/mux v1.8.0
	github.com/kubernetes-sigs/service-catalog v0.3.1
	github.com/prometheus/common v0.4.1
	github.com/tmax-cloud/template-operator v0.0.0-20210324013110-d35a77c12859
	k8s.io/api v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v0.18.6
	sigs.k8s.io/controller-runtime v0.6.2
)
