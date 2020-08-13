module github.com/jwkim1993/template-service-broker

go 1.14

require (
	github.com/gorilla/mux v1.7.2
	github.com/operator-framework/operator-sdk v0.17.1
	github.com/tidwall/gjson v1.6.0
	github.com/youngind/hypercloud-operator v0.0.0-20200806070920-93ed5a774a00
	k8s.io/api v0.17.4
	k8s.io/apimachinery v0.17.4
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.5.2
)

replace k8s.io/client-go => k8s.io/client-go v0.17.4
