module github.com/tmax-cloud/template-service-broker-go

go 1.14

require (
	github.com/gorilla/mux v1.7.2
	github.com/operator-framework/operator-sdk v0.17.1
	github.com/tidwall/gjson v1.6.0
	github.com/tmax-cloud/template-operator v0.0.0-20201014062702-3775627ef3a2
	k8s.io/api v0.17.4
	k8s.io/apimachinery v0.17.4
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.5.2
)

replace k8s.io/client-go => k8s.io/client-go v0.17.4
