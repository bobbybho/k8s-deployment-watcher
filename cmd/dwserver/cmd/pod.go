package cmd

import (
	"github.com/bobbybho/k8s-deployment-watcher/common"
	watcher "github.com/bobbybho/k8s-deployment-watcher/watcher"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
)

var podCmd = &cobra.Command{
	Use:   "pod",
	Args:  cobra.NoArgs,
	Short: "deployment commands",
	Long:  `deployment commands`,
}

var podWatchCmd = &cobra.Command{
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

		q := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), common.PodQueue)
		pw := watcher.NewPodWatcher(clientset, namespace, q)

		stop := make(chan struct{})
		defer close(stop)
		err = pw.Run(stop)
		if err != nil {
			klog.Fatal(err)
		}
		select {}
	},
}

func init() {
	podCmd.AddCommand(podWatchCmd)
	podWatchCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", nameSpaceDefault, "pod namespace")
}
