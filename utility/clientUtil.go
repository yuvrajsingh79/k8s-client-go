package utility

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	// apiextension "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	// myresourceclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// GetClientConfig first tries to get a config object which uses the service account kubernetes gives to pods,
// if it is called from a process running in a kubernetes environment.
// Otherwise, it tries to build config from a default kubeconfig filepath if it fails, it fallback to the default config.
// Once it get the config, it returns the same.
func GetClientConfig() (*rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Printf("Unable to create config. Error: %+v", err)
		err1 := err
		kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			err = fmt.Errorf("InClusterConfig as well as BuildConfigFromFlags Failed. Error in InClusterConfig: %+v\nError in BuildConfigFromFlags: %+v", err1, err)
			return nil, err
		}
	}

	return config, nil
}

// GetClientset first tries to get a config object which uses the service account kubernetes gives to pods,
// if it is called from a process running in a kubernetes environment.
// Otherwise, it tries to build config from a default kubeconfig filepath if it fails, it fallback to the default config.
// Once it get the config, it creates a new Clientset for the given config and returns the clientset.
func GetClientset() (*kubernetes.Clientset, error) {
	config, err := GetClientConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		err = fmt.Errorf("Failed creating clientset. Error: %+v", err)
		return nil, err
	}

	return clientset, nil
}

//GetKubeClient functions helps to connect kubeclient for Custom Resource
// func GetKubeClient() (apiextension.Interface, myresourceclientset.Interface, kubernetes.Interface) {
// 	// construct the path to resolve to `~/.kube/config`
// 	kubeConfigPath := os.Getenv("HOME") + "/.kube/config"

// 	// create the config from the path
// 	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
// 	if err != nil {
// 		log.Fatalf("getClusterConfig: %v", err)
// 	}

// 	// generate the client based off of the config
// 	client, err := apiextension.NewForConfig(config)
// 	if err != nil {
// 		log.Fatalf("getClusterConfig: %v", err)
// 	}

// 	kubeClient, err := kubernetes.NewForConfig(config)
// 	if err != nil {
// 		log.Fatalf("getClusterConfig: %v", err)
// 	}

// 	myresourceClient, err := myresourceclientset.NewForConfig(config)
// 	if err != nil {
// 		log.Fatalf("getClusterConfig: %v", err)
// 	}

// 	log.Info("Successfully constructed k8s client")
// 	return client, myresourceClient, kubeClient
// }

// PrettyString returns the prettified string of the interface supplied. (If it can)
func PrettyString(in interface{}) string {
	jsonStr, err := json.MarshalIndent(in, "", "    ")
	if err != nil {
		err := fmt.Errorf("Unable to marshal, Error: %+v", err)
		if err != nil {
			fmt.Printf("Unable to marshal, Error: %+v\n", err)
		}
		return fmt.Sprintf("%+v", in)
	}
	return string(jsonStr)
}
