//go:build debug

package main

type Config struct {
	XiheUsername  string `json:"xihe_username"`
	KubeConfig    string `json:"kubeconfig"`
	Namespace     string `json:"namespace"`
	Appconfig     string `json:"appconfig"`
	Deployconfig  string `json:"deployconfig"`
	StartupScript string `json:"startup_script"`
	HeartbeatUrl  string `json:"heartbeat_url"`
}

type RuntimeState struct {
	PodName     string `json:"pod_name"`
	DeployName  string `json:"deploy_name"`
	ProjectName string `json:"project_name"`
}

type NhctlOutput struct {
	DeployName string
	PodName    string
}
