package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog"
)

func collectDataAndAnnotatePod(config *rest.Config, dirPath string) {
	annotations := getAnnotationsFromFiles(dirPath)
	updateAnnotations(config, annotations)
}

func updateAnnotations(config *rest.Config, annotations map[string]string) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	podsClient := clientset.CoreV1().Pods(namespace)

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of Pod before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		pod, err := podsClient.Get(context.TODO(), podName, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				log.Fatalf("Pod %s was not found in namespace %s", podName, namespace)
			} else {
				log.Fatalf("Failed to get latest version of Pod: %v", err)
			}
		}

		for key, value := range annotations {
			pod.ObjectMeta.Annotations[key] = value
		}

		klog.V(5).Infof("Updating annotations: %+v", annotations)

		_, updateErr := podsClient.Update(context.TODO(), pod, metav1.UpdateOptions{})

		if updateErr != nil {
			klog.V(5).Infof("Error updating pod. will retry. Error: %s", updateErr)
		} else {
			klog.V(5).Info("Pod annotated successfully")
		}

		return updateErr
	})

	if retryErr != nil {
		panic(fmt.Errorf("Update failed: %v", retryErr))
	}
}

func getAnnotationsFromFiles(dirPath string) map[string]string {
	annotations := map[string]string{}

	klog.V(5).Infof("getting annotations from files in %s", dirPath)

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			// Ignore sub directories
			return nil
		}

		_, filename := filepath.Split(path)
		key := annotationsPrefix + filename
		annotations[key] = readFileContent(path)
		klog.V(5).Infof("found annotation %s=%s", key, annotations[key])

		return nil
	})

	if err != nil {
		panic(err)
	}

	return annotations
}

func readFileContent(filename string) string {
	klog.V(5).Infof("Reading contents of file %s", filename)

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		klog.Fatalf("Failed to read file %s. Error: %s", filename, err)
	}

	value := string(b)
	return value
}
