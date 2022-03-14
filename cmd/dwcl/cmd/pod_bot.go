package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	podbot "github.com/bobbybho/k8s-deployment-watcher/testbots/podbots"
	"github.com/spf13/cobra"
)

var podBotCmd = &cobra.Command{
	Use:   "PodBots",
	Short: "PodBots command",
	Args:  cobra.NoArgs,
}

var podBotRunCmd = &cobra.Command{
	Use:   "run [BOT_COUNT] [Remote GRPC Server addr",
	Short: "pod bots run command",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		var podBotCnt int64
		var err error
		//var opt grpc.DialOption
		var remoteAddr string

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// register for signals
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGQUIT)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGHUP)
		signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

		if podBotCnt, err = strconv.ParseInt(args[0], 10, 64); err != nil {
			fmt.Fprintf(os.Stderr, "PODBOT_COUNT=%s must be an integer\n", args[0])
			os.Exit(1)
		}

		remoteAddr = args[1]
		if remoteAddr == "" {
			fmt.Fprintf(os.Stderr, "remote GRPC server addr=%s must not be empty\n", args[1])
			os.Exit(1)
		}

		podBotList := make([]podbot.PodBot, podBotCnt)

		for i := 0; i < int(podBotCnt); i++ {
			podBotName := fmt.Sprintf("podbot-%d", i)
			podBotList = append(podBotList, podbot.PodBot{Name: podBotName})
		}

		for _, pBot := range podBotList {
			pB := pBot
			go func() {
				pB.Run(ctx, remoteAddr)
			}()
		}

	waitLoop:
		for {
			select {
			case sig := <-sigs:
				log.Printf("\nReceived a signal %v\n", sig)
				cancel()
				break waitLoop
			default:
				time.Sleep(1 * time.Second)
			}
		}

	},
}

func init() {
	podBotCmd.AddCommand(podBotRunCmd)
	podBotRunCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", nameSpaceDefault, "pod namespace")
}
