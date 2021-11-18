// Package k8s provides functions to interact with
// the kubernetes API to create a secret that contains
// app role credentials for vault backups
package k8s

import (
	"context"
	"flag"
	"path/filepath"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	coreV1Types "k8s.io/client-go/kubernetes/typed/core/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var secretsClient coreV1Types.SecretInterface

func CreateK8sSecret(){
	initClient()

	secret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foobarbaz",
			Namespace: "default",
		},
		StringData: map[string]string{
			"ROLE_ID": "something",
			"SECRET_ID": "somethingelse",
		},
	}

	opts := metav1.CreateOptions{}

	secretsClient.Create(context.TODO(), &secret, opts)
}



func initClient() {
	var err error
	var config *rest.Config
  var clientset *kubernetes.Clientset
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
			panic(err.Error())
		}
	}

	// create the clientset
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
			panic(err.Error())
	}

	// create the secretsClient
	secretsClient = clientset.CoreV1().Secrets("default")
}
