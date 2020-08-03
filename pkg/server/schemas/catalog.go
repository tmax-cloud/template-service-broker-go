package schemas

type CatalogGetResponse struct {
	Services []Service `json:"services`
}

type Service struct {
	Name            string            `json:"name"`
	Id              string            `json:"id"`
	Description     string            `json:"description"`
	Tags            []string          `json:"tags,omitempty"`
	Requires        []string          `json:"requires,omitempty"`
	Bindable        bool              `json:"bindable"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	DashboardClient DashBoardClient   `json:"dashboard_client,omitempty`
	PlanUpdateable  bool              `json:"plan_updateable,omitempty"`
	Plans           Plan              `json:"plans`
}

type DashBoardClient struct {
	Id          string `json:"id,omitempty"`
	Secret      string `json:"secret,omitempty"`
	RedirectUri string `json:"redirect_uri,omitempty"`
}

type Plan struct {
	Id              string            `json:"id"`
	Name            string            `json:"name"`
	Description     string            `json:"description"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	MaintenanceInfo MaintenanceInfo   `json:"maintenance_info,omitempty"`
	Free            bool              `json:"free,omitempty"`
	Bindable        bool              `json:"bindable,omitempty"`
	Schemas         Schemas           `json:schemas,omitempty`
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
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}
