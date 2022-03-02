package common

import (
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// ClientConfig loads kubectl configuration
func ClientConfig(configPath string) (*rest.Config, error) {
	return clientcmd.BuildConfigFromFlags("", configPath)
}
