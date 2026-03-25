//go:build debug

package main

type Config struct {
	AppName        string `json:"appName"`
	KubeConfig     string `json:"kubeConfig"`
	Namespace      string `json:"namespace"`
	Appconfig      string `json:"appConfig"`
	Deployconfig   string `json:"deployConfig"`
	StartupScript  string `json:"startupScript"`
	HeartbeatUrl   string `json:"heartbeatUrl"`
	OrigDeployName string `json:"origDeployName"`
}

type RuntimeState struct {
	PodName     string `json:"podName"`
	DeployName  string `json:"deployName"`
	ProjectName string `json:"projectName"`
}

type NhctlOutput struct {
	DeployName string
	PodName    string
}
