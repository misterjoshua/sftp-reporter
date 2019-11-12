package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// SftpEntry is an output.
type SftpEntry struct {
	Namespace string   `json:"namespace"`
	Addresses []string `json:"addresses"`
}

func getSftpServices(clientset *kubernetes.Clientset) ([]SftpEntry, error) {
	var sftpEntries []SftpEntry

	sftpEntries, err := appendServicesMatchingLabel(clientset, "app=microsite", sftpEntries)
	if err != nil {
		return []SftpEntry{}, err
	}

	sftpEntries, err = appendServicesMatchingLabel(clientset, "app=lamp", sftpEntries)
	if err != nil {
		return []SftpEntry{}, err
	}

	return sftpEntries, nil
}

func appendServicesMatchingLabel(clientset *kubernetes.Clientset, labelSelector string, sftpEntries []SftpEntry) ([]SftpEntry, error) {

	svcs, err := clientset.CoreV1().Services("").List(metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return []SftpEntry{}, err
	}

	for _, r := range svcs.Items {
		if strings.HasSuffix(r.Name, "-sftp") {
			var addresses []string

			for _, ingress := range r.Status.LoadBalancer.Ingress {
				for _, port := range r.Spec.Ports {
					address := fmt.Sprintf("%v:%v", ingress.IP, port.Port)
					addresses = append(addresses, address)
				}
			}

			sftpEntries = append(sftpEntries, SftpEntry{
				Namespace: r.Namespace,
				Addresses: addresses,
			})
		}
	}

	return sftpEntries, nil
}

var globalClient *kubernetes.Clientset

func main() {
	kubeconfig := flag.String("kubeconfig", "/home/josh/.kube/config", "kubeconfig file")
	address := flag.String("address", ":8090", "http listen address")
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		errorExit("Building config from flags", err)
	}

	globalClient, err = kubernetes.NewForConfig(config)
	if err != nil {
		errorExit("New clientset for config", err)
	}

	_, err = getSftpServices(globalClient)
	if err != nil {
		errorExit("Trying to get sftp entries before serving http", err)
	}

	log.Infof("Beginning HTTP server on listen address: %v\n", *address)
	http.HandleFunc("/sftp", handleSftpRequest)

	err = http.ListenAndServe(*address, nil)
	if err != nil {
		errorExit("Trying to listen and serve http", err)
	}
}

func errorExit(message string, err error) {
	log.Errorf("%s: %v\n", message, err)
	os.Exit(1)
}

func handleSftpRequest(w http.ResponseWriter, req *http.Request) {
	log.Infof("Serving /sftp request from %v", req.RemoteAddr)
	svcs, err := getSftpServices(globalClient)

	if err != nil {
		log.Errorf("Error getting sftp services: %v", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(svcs)
}
