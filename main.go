package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const THRESHOLD_MINUTES = 2

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	kubeconfig := os.Getenv("KUBECONFIG")

	mode := flag.String("mode", "once", "Mode of operation: 'once' or 'repeat'")
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatal().Err(err).Msgf("Error building config with %s", kubeconfig)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal().Err(err).Msgf("Error building clientset")
	}

	checkPendingPods(clientset)

	if *mode == "repeat" {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				log.Debug().Msg("Checking for pending pods")
				go checkPendingPods(clientset)
			}
		}
	}

}

func checkPendingPods(clientset *kubernetes.Clientset) {
	options := metav1.ListOptions{
		FieldSelector: "status.phase=Pending",
	}
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), options)
	if err != nil {
		log.Err(err).Msg("Error getting pods")
		return
	}
	now := time.Now()
	for _, pod := range pods.Items {
		age := now.Sub(pod.CreationTimestamp.Time)
		if age > THRESHOLD_MINUTES*time.Minute {
			log.Info().
				Str("Namespace", pod.Namespace).
				Str("Pod Name", pod.Name).
				Str("Node", getNode(pod.Spec.Affinity)).
				Dur("Age", age).
				Msgf("Pod pending for more than %d minutes", THRESHOLD_MINUTES)
		}
	}
}

func getNode(aff *v1.Affinity) string {
	if aff != nil {
		if aff.NodeAffinity != nil {
			if aff.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil {
				for _, term := range aff.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
					for _, matchField := range term.MatchFields {
						if matchField.Key == "metadata.name" {
							return matchField.Values[0]
						}
					}
				}
			}
		}
	}
	return ""
}
