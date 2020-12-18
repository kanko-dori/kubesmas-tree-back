package main

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"main/persistence/redis"
	"os"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)
var CLIENT_NUM = "CLIENT_NUM"

func main(){
	fmt.Println("Hello, worker is starting...")
	time.Sleep(time.Second * 5)
	redisPath := os.Getenv("REDIS_PATH")
	if redisPath == "" {
		fmt.Printf("failed to get REDIS_PATH\n")
		return
	}
	client, err := redis.New(redisPath)
	if err != nil {
		fmt.Printf("failed to get redis client: %v\n", err)
		return
	}
	defer client.Close()

	var currentNum int
	var oldNum int

	for {
		currentNum, err = redis.GetIDByToken(client, "CLIENT_NUM")
		fmt.Printf("%d clients are connecting now\n", currentNum)
		if err == redis.Nil {
			fmt.Println("CLIENT_NUM does not exist.\n")
		} else if err != nil {
			fmt.Printf("failed to call getIDByToken: %v\n", err)
		}
		if oldNum != currentNum {
			fmt.Printf("try to update replicas %d to %d\n", oldNum, currentNum)
			err = editDeployment(currentNum)
			if err != nil {
				fmt.Printf("failed to edit Deployment: %v\n", err)
			}
			oldNum = currentNum
		}
		time.Sleep(time.Second * 5)
	}


}

func editDeployment(desiredNum int) error {
	config, err := rest.InClusterConfig()
	if err != nil {
		return errors.Wrap(err, "failed to get cluster config")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return errors.Wrap(err, "failed to get clientset")
	}

	dClient := clientset.AppsV1().Deployments("kubesmas-tree")

	deployment, err := dClient.Get(context.TODO(), "nginx", metav1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to get nginx deployment")
	}
	replicas := int32(desiredNum)
	deployment.Spec.Replicas = &replicas

	_, err = dClient.Update(context.TODO(), deployment, metav1.UpdateOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to update deployment")
	}
	return nil
}