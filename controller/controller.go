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
