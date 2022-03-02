package cmd

import (
	"github.com/spf13/cobra"
	"gitlab.com/bobbyho10/k8s-deployment-watcher/common"
	"gitlab.com/bobbyho10/k8s-deployment-watcher/controller"
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

		if kubeConfig, err = common.ClientConfig(kubeConfigPath); err != nil {
			panic(err.Error())
		}

		clientset, err := kubernetes.NewForConfig(kubeConfig)
		if err != nil {
			panic(err.Error())
		}

		pc := controller.NewPodController(clientset, namespace)

		stop := make(chan struct{})
		defer close(stop)
		pc.Run(stop)
		select {}
	},
}

func init() {
	podControllerCmd.AddCommand(podControllerWatchCmd)
	podControllerWatchCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", nameSpaceDefault, "pod namespace")
}
