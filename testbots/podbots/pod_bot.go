package podbot

import (
	"context"
	"io"
	"log"

	pb "github.com/bobbybho/k8s-deployment-watcher/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type PodBot struct {
	Name string
}

func NewPodBot(name string) *PodBot {
	return &PodBot{
		Name: name,
	}
}

// Connect opens a grpc connection
func GRPCConnect(address string, opt ...grpc.DialOption) (*grpc.ClientConn, error) {

	if len(opt) == 0 {
		opt = []grpc.DialOption{grpc.WithTransportCredentials(credentials.NewTLS(nil))}
	}
	opt = append(opt, grpc.WithBlock())
	return grpc.Dial(address, opt...)
}

// Run ...
func (p *PodBot) Run(ctx context.Context, remoteAddr string, opt ...grpc.DialOption) error {
	var conn *grpc.ClientConn
	var err error

	if conn, err = GRPCConnect(remoteAddr, opt...); err != nil {
		log.Printf("Failed to connect to remote addr %v error: %v\n", remoteAddr, err.Error())
	}

	go p.RunPodStat(ctx, conn)

	<-ctx.Done()
	return nil
}

// RunPodStat
func (p *PodBot) RunPodStat(ctx context.Context, conn *grpc.ClientConn) {
	client := pb.NewPodStatIntfClient(conn)

	listenRequest := pb.PodStatRequest{}
	listenRequest.Clientid = p.Name

	stream, err := client.ListenPodStatus(ctx, &listenRequest)
	if err != nil {
		log.Fatalf("Failed to listen to podstatus stream. podbot=%v error=%v\n", p.Name, err.Error())

	}

listenLoop:
	for {
		select {
		case <-ctx.Done():
			log.Printf("[%v] received done signal\n", p.Name)
			break listenLoop
		default:
			podStatusReply, err := stream.Recv()
			if err == io.EOF {
				break listenLoop
			}

			if err != nil {
				log.Printf("Failed to receive podstat stream: %v\n", err.Error())
				return
			}

			log.Printf("Receive PodStatusReply %+v\n", podStatusReply)

		}
	}
}
