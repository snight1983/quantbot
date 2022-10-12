package main

import (
	"fmt"
	"log"
	"net"
	"quantbot/server"
	"quantbot/stocks"
	"quantbot/utils"

	"google.golang.org/grpc"

	pb "quantbot/proto/pb"
)

/*****************************************
  数据采集和清洗
	1.历史数据采集
	  1.1 东财(5分钟K线数据),回溯一周。
		    前复权、后复权、不复权三类。

	2.实时数据,计算5分钟K线
	  2.1 东财
		2.2 雪球
		2.3 163
		2.4 新浪(反爬,暂不启用)
*****************************************/

func main() {
	stocks.StartCollect()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", utils.Conf.Port))
	if err != nil {
		log.Fatalf("failed to listen :%v", err)
	}

	grpc := grpc.NewServer()
	pb.RegisterRegeditServer(grpc, &server.EndPointRegeditService{})
	if err := grpc.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
