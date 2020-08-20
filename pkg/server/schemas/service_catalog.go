package schemas

import (
	tmaxv1 "github.com/jwkim1993/hypercloud-operator/pkg/apis/tmax/v1"
)

type Catalog struct {
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
	Plans           []tmaxv1.PlanSpec `json:"plans`
}

type DashBoardClient struct {
	Id          string `json:"id,omitempty"`
	Secret      string `json:"secret,omitempty"`
	RedirectUri string `json:"redirect_uri,omitempty"`
}
