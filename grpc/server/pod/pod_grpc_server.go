package podserver

import (
	"context"
	"log"

	pc "github.com/bobbybho/k8s-deployment-watcher/controller"
	pb "github.com/bobbybho/k8s-deployment-watcher/proto"
)

// PodServer ...
type PodServer struct {
	pb.UnimplementedPodStatIntfServer
	PodController *pc.PodController
}

// ListenPodStatus ...
func (p *PodServer) ListenPodStatus(r *pb.PodStatRequest, stream pb.PodStatIntf_ListenPodStatusServer) error {
	clientID := r.GetClientid()

	ch := p.PodController.OpenChannel(clientID)

	if ch != nil {
		for {
			select {
			case <-stream.Context().Done():
				log.Printf("stream.Context.Done(): clientID: %v\n", clientID)
			case msg := <-ch:
				if err := stream.Send(&msg); err != nil {
					p.PodController.CloseChannel(clientID)
					log.Printf("Failed to send podstat err=%v\n", err.Error())
					return err
				}
			}
		}
	} else {
		log.Printf("podstat channel is empty\n")
	}

	return nil
}

// GetAllPodStatus ...
func (p *PodServer) GetAllPodStatus(r *pb.PodStatRequest, stream pb.PodStatIntf_GetAllPodStatusServer) error {
	// TODO: implement this function
	return nil
}

//GetPodStatusByName ...
func (p *PodServer) GetPodStatusByName(ctx context.Context, r *pb.PodStatRequest) (*pb.PodStatReply, error) {
	// TODO: implement this function
	return &pb.PodStatReply{}, nil
}
