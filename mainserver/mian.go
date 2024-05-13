package main

import (
	_ "github.com/go-sql-driver/mysql"
	"google.golang.org/grpc"
	"log"
	"net"
	"time"
	"user_growth/conf"
	"user_growth/dbhelper"
	"user_growth/pb"
	"user_growth/ugserver"
)

func initDb() {
	time.Local = time.UTC
	conf.LoadConfigs()
	dbhelper.InitDb()
}

func main() {
	initDb()
	lis, err := net.Listen("tcp", ":1111")
	if err != nil {
		log.Fatalf("Failed too listen :%s", err.Error())
	}
	s := grpc.NewServer()
	// 注册服务
	pb.RegisterUserCoinServer(s, &ugserver.UgCoinServer{})
	pb.RegisterUserGradeServer(s, &ugserver.UgGradeServer{})
	// 启动服务
	log.Printf("server listening at %v\n", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to server: %v", err)
	}
}
