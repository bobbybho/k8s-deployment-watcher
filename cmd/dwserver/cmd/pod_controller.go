package cmd

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bobbybho/k8s-deployment-watcher/controller"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var podControllerCmd = &cobra.Command{
	Use:   "pod-controller",
	Args:  cobra.NoArgs,
	Short: "pod controller commands",
	Long:  `pod controller commands`,
}

var podControllerWatchCmd = &cobra.Command{
	Use:   "watch-endpoints [namespace]",
	Args:  cobra.NoArgs,
	Short: "watch deployment point endpoints",
	Long:  `watch deployment point endpoints`,
	Run: func(cmd *cobra.Command, args []string) {

		var (
			kubeConfig *rest.Config
			err        error
		)

		if kubeConfig, err = rest.InClusterConfig(); err != nil {
			panic(err.Error())
		}

		clientset, err := kubernetes.NewForConfig(kubeConfig)
		if err != nil {
			panic(err.Error())
		}

		// register for signals
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGQUIT)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGHUP)
		signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

		pc := controller.NewPodController(clientset, namespace)

		stop := make(chan struct{})
		defer close(stop)
		pc.Run(stop)

	waitloop:
		for {
			select {
			case sig := <-sigs:
				log.Printf("Received a signal: %v\n", sig)
				break waitloop
			}
		}
		log.Println("Bye")
	},
}

func init() {
	podControllerCmd.AddCommand(podControllerWatchCmd)
	podControllerWatchCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", nameSpaceDefault, "pod namespace")
}
