package schemas

type Context struct {
	ClusterId    string `json:"clusterid"`
	InstanceName string `json:"instance_name"`
	Namespace    string `json:"namespace"`
	Platform     string `json:"platform"`
}

type Error struct {
	Error            string `json:"error,omitempty"`
	Description      string `json:"description,omitempty"`
	InstanceUsable   bool   `json:"instance_usable,omitempty"`
	UpdateRepeatable bool   `json"update_repeatable,omitempty"`
}
