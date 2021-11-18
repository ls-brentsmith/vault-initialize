// Package k8s provides functions to interact with
// the kubernetes API to create a secret that contains
// app role credentials for vault backups
package k8s

import (
	"context"
	"flag"
	"log"
	"path/filepath"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var clientset *kubernetes.Clientset

func CreateK8sSecret(namespace, secretName, roleID, secretID  string){
	initClient()

	// create the secretsClient
	secretsClient := clientset.CoreV1().Secrets(namespace)
	secret, err := secretsClient.Get(context.TODO(), secretName, metav1.GetOptions{})

	// Add Secret Data
	secret.StringData = map[string]string{
		"ROLE_ID": roleID,
		"SECRET_ID": secretID,
	}

	if errors.IsNotFound(err) {
		// Create the secret
		secret.ObjectMeta = metav1.ObjectMeta{
			Name: secretName,
		}

		secret, err = secretsClient.Create(context.TODO(), secret, metav1.CreateOptions{})
	} else {
		// Update the secret
		secret, err = secretsClient.Update(context.TODO(), secret, metav1.UpdateOptions{})
	}

	if err != nil {
		log.Panic(err)
	}

	log.Printf("Succesfull created secret: %v", secret.Name)
}

func initClient() {
	var err error
	var config *rest.Config
	var kubeconfig *string

	// creates the in-cluster config
	config, err = rest.InClusterConfig()
	if err != nil {
		// assume Out of Cluster Config
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		} else {
			kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
		}
		flag.Parse()

		// use the current context in kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			log.Panic(err.Error())
		}
	}

	// create the clientset
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
			log.Panic(err.Error())
	}
}
