package controller

import (
	"client-go/k8s-client-go/utility"
	"context"
	"fmt"
	"net/http"
	"strings"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//GetPodList fetches all the running pods in the k8s
func GetPodList(w http.ResponseWriter, r *http.Request) {
	clientset, err := utility.GetClientset()
	if err != nil {
		panic(err)
	}

	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), meta_v1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	svc, err := clientset.CoreV1().Services("").List(context.TODO(), meta_v1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("\n\nThere are %d pods in the cluster\n\n", len(pods.Items))
	for _, pod := range pods.Items {
		fmt.Println(utility.PrettyString(pod))
	}

	fmt.Println(strings.Repeat("*", 80))
	fmt.Printf("\nThere are %d services in the cluster\n\n", len(svc.Items))
	for _, svc := range svc.Items {
		fmt.Println(utility.PrettyString(svc))
	}
}

// func CreateCRD(w http.ResponseWriter, r *http.Request) {
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
