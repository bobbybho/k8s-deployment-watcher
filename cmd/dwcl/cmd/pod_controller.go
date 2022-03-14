package cmd

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/bobbybho/k8s-deployment-watcher/common"
	"github.com/bobbybho/k8s-deployment-watcher/controller"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	podserver "github.com/bobbybho/k8s-deployment-watcher/grpc/server/pod"
	pb "github.com/bobbybho/k8s-deployment-watcher/proto"
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
			errc       chan error
		)

		if kubeConfig, err = common.ClientConfig(kubeConfigPath); err != nil {
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

		addr := "0.0.0.0:8088"

		// TODO: the listening addr should be read from a config
		lis, err := net.Listen("tcp", addr)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		s := grpc.NewServer()
		pb.RegisterPodStatIntfServer(s, &podserver.PodServer{})

		go func() {
			log.Printf("GRPC server is listening on %v", addr)
			errc <- s.Serve(lis)
		}()

		stop := make(chan struct{})
		defer close(stop)
		pc.Run(stop)

	waitloop:
		for {
			select {
			case sig := <-sigs:
				log.Printf("Received a signal: %v\n", sig)
				break waitloop
			case err := <-errc:
				log.Printf("Received error from gRPC server: %v\n", err.Error())
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
