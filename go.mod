module template-service-broker

go 1.14

require (
	github.com/operator-framework/operator-sdk v0.19.2 // indirect
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.6.1
)

replace k8s.io/client-go => k8s.io/client-go v0.17.4
