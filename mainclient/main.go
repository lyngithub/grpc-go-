package main

import (
	"context"
	"flag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"time"
	"user_growth/pb"
)

func main() {
	//连接到服务
	add := flag.String("addr", "localhost:1111", "address")
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	conn, err := grpc.Dial(*add, opts...)
	if err != nil {
		log.Fatalf("did not connect : %v", err)
	}
	defer conn.Close()
	// 请求服务
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	// 新建客户端
	cCoin := pb.NewUserCoinClient(conn)
	cGrade := pb.NewUserGradeClient(conn)
	// 测试1：UserCoinServer.ListTasks
	r1, err := cCoin.ListTasks(ctx, &pb.ListTasksRequest{})
	if err != nil {
		log.Printf("cCoin.ListTasks error=%v\n", err)
	} else {
		log.Printf("cCoin.ListTasks: %+v\n", r1.GetDatalist())
	}
	// 测试2：UserGradeServer.ListGrades
	r2, err := cGrade.ListGrades(ctx, &pb.ListGradesRequest{})
	if err != nil {
		log.Printf("cGrade.ListGrades error=%v\n", err)
	} else {
		log.Printf("cGrade.ListGrades: %+v\n", r2.GetDatalist())
	}
	// 测试3：修改积分
	r3, err := cCoin.UserCoinChange(ctx, &pb.UserCoinChangeRequest{
		Uid:  0,
		Task: "abc",
		Coin: 0,
	})
	if err != nil {
		log.Printf("cCoin.UserCoinChange error=%v\n", err)
	} else {
		log.Printf("cCoin.UserCoinChange: %+v\n", r3.GetUser())
	}
}
