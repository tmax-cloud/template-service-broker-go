package schemas

type ServiceInstanceProvisionRequest struct {
	ServiceId        string            `json:"service_id"`
	PlanId           string            `json:"plan_id"`
	Context          map[string]string `json:"context,omitempty"`
	OrganizationGuid string            `json:"organization_guid"`
	SpaceGuid        string            `json:"space_guid"`
	Parameters       map[string]string `json:"parameters,omitempty"`
}

type ServiceInstanceProvisionResponse struct {
	DashboardUrl string                  `json:"dashboard_url,omitempty"`
	Operation    string                  `json:"operation,omitempty"`
	Metadata     ServiceInstanceMetadata `json:"metadata,omitempty"`
}

type ServiceInstanceMetadata struct {
	Labels map[string]string `json:"labels,omitempty"`
}
