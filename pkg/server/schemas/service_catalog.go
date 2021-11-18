package schemas

import "k8s.io/apimachinery/pkg/util/intstr"

type Catalog struct {
	Services []Service `json:"services"`
}

type Service struct {
	Name            string                 `json:"name"`
	Id              string                 `json:"id"`
	Description     string                 `json:"description"`
	Tags            []string               `json:"tags,omitempty"`
	Requires        []string               `json:"requires,omitempty"`
	Bindable        bool                   `json:"bindable"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	DashboardClient DashBoardClient        `json:"dashboard_client,omitempty"`
	PlanUpdateable  bool                   `json:"plan_updateable,omitempty"`
	Plans           []PlanSpec             `json:"plans"`
}

type DashBoardClient struct {
	Id          string `json:"id,omitempty"`
	Secret      string `json:"secret,omitempty"`
	RedirectUri string `json:"redirect_uri,omitempty"`
}

type PlanSpec struct {
	Id                     string          `json:"id,omitempty"`
	Name                   string          `json:"name"`
	Description            string          `json:"description,omitempty"`
	Metadata               PlanMetadata    `json:"metadata,omitempty"`
	Free                   bool            `json:"free,omitempty"`
	Bindable               bool            `json:"bindable,omitempty"`
	PlanUpdateable         bool            `json:"plan_updateable,omitempty"`
	Schemas                Schemas         `json:"schemas,omitempty"`
	MaximumPollingDuration int             `json:"maximum_polling_duration,omitempty"`
	MaintenanceInfo        MaintenanceInfo `json:"maintenance_info,omitempty"`
}

type PlanMetadata struct {
	Bullets     []string `json:"bullets,omitempty"`
	Costs       Cost     `json:"costs,omitempty"`
	DisplayName string   `json:"displayName,omitempty"`
}

type Cost struct {
	Amount int    `json:"amount"`
	Unit   string `json:"unit"`
}

type MaintenanceInfo struct {
	Version     string `json:"version"`
	Description string `json:"description,omitempty"`
}
type Schemas struct {
	ServiceInstance ServiceInstanceSchema `json:"service_instance,omitempty"`
	ServiceBinding  ServiceBindingSchema  `json:"service_binding,omitempty"`
}

type ServiceInstanceSchema struct {
	Create SchemaParameters `json:"create,omitempty"`
	Update SchemaParameters `json:"update,omitempty"`
}

type ServiceBindingSchema struct {
	Create SchemaParameters `json:"create,omitempty"`
}

type SchemaParameters struct {
	Parameters SchemaParameterSpec `json:"parameters,omitempty"`
}

type SchemaParameterSpec struct {
	Properties map[string]PropertiesSpec `json:"properties,omitempty"`
	Required   []string                  `json:"required,omitempty"`
}

type PropertiesSpec struct {
	Default     intstr.IntOrString `json:"default,omitempty"`
	Fixed       bool               `json:"fixed,omitempty"`
	Title       string             `json:"title,omitempty"`
	Description string             `json:"description,omitempty"`
	Type        string             `json:"type,omitempty"`
	Regex       string             `json:"regex,omitempty"`
}

type ParamSpec struct {
	Description string `json:"description,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	From        string `json:"from,omitempty"`
	Generate    string `json:"generate,omitempty"`
	Name        string `json:"name"`
	Required    bool   `json:"required,omitempty"`
	Value       string `json:"value,omitempty"`
	ValueType   string `json:"valueType,omitempty"`
}
