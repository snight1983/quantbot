package server

import (
	"context"
	pb "quantbot/proto/pb"
)

type EndPointRegeditService struct {
}

func (userService *EndPointRegeditService) EndPointRegedit(ctx context.Context, in *pb.RegEditRQ) (*pb.RegEditRS, error) {

	return &pb.RegEditRS{}, nil
}
