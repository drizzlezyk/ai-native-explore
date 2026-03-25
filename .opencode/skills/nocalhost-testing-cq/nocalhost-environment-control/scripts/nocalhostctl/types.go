//go:build debug

package main

type Config struct {
	DeveloperName  string `json:"developerName"`
	KubeConfig     string `json:"kubeConfig"`
	Namespace      string `json:"namespace"`
	Appconfig      string `json:"appConfig"`
	Deployconfig   string `json:"deployConfig"`
	StartupScript  string `json:"startupScript"`
	BuildScript    string `json:"buildScript"`
	HeartbeatUrl   string `json:"heartbeatUrl"`
	OrigDeployName string `json:"origDeployName"`
	BinaryName     string `json:"binaryName"`
	ProjectPath    string `json:"projectPath"`
	RemotePort     string `json:"remotePort"`
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
