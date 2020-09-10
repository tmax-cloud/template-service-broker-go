module github.com/jitaeyun/template-service-broker

go 1.14

require (
	github.com/gorilla/mux v1.7.2
	github.com/jitaeyun/template-operator v0.0.0-20200910080510-df75d18b246b
	github.com/operator-framework/operator-sdk v0.17.1
	github.com/tidwall/gjson v1.6.0
	k8s.io/api v0.17.4
	k8s.io/apimachinery v0.17.4
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.5.2
)

replace k8s.io/client-go => k8s.io/client-go v0.17.4
