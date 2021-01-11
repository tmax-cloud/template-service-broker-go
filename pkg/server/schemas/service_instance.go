package schemas

import "k8s.io/apimachinery/pkg/util/intstr"

type ServiceInstanceProvisionRequest struct {
	ServiceId        string                        `json:"service_id"`
	PlanId           string                        `json:"plan_id"`
	Context          Context                       `json:"context,omitempty"`
	OrganizationGuid string                        `json:"organization_guid"`
	SpaceGuid        string                        `json:"space_guid"`
	Parameters       map[string]intstr.IntOrString `json:"parameters,omitempty"`
}

type ServiceInstanceProvisionResponse struct {
	DashboardUrl string                  `json:"dashboard_url,omitempty"`
	Operation    string                  `json:"operation,omitempty"`
	Metadata     ServiceInstanceMetadata `json:"metadata,omitempty"`
}

type ServiceInstanceMetadata struct {
	Labels map[string]string `json:"labels,omitempty"`
}
