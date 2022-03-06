package cmd

import (
	"github.com/bobbybho/k8s-deployment-watcher/common"
	watcher "github.com/bobbybho/k8s-deployment-watcher/watcher"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
)

var deploymentCmd = &cobra.Command{
	Use:   "deployment",
	Args:  cobra.NoArgs,
	Short: "deployment commands",
	Long:  `deployment commands`,
}

var deploymentWatchCmd = &cobra.Command{
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

		dw := watcher.NewDeploymentWatcher(clientset, namespace)

		stop := make(chan struct{})
		defer close(stop)
		err = dw.Run(stop)
		if err != nil {
			klog.Fatal(err)
		}
		select {}
	},
}

func init() {
	deploymentCmd.AddCommand(deploymentWatchCmd)
	deploymentWatchCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", nameSpaceDefault, "deployment namespace")
}
