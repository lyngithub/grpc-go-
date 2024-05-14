package main

import (
	"context"
	"log"
	"net/http"
	"time"
	"user_growth/conf"
	"user_growth/dbhelper"
	"user_growth/pb"
	"user_growth/ugserver"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
)

func initDb() {
	time.Local = time.UTC
	conf.LoadConfigs()
	dbhelper.InitDb()
}

var AllowOrigin = map[string]bool{
	"http://a.site.com": true,
	"http://b.site.com": true,
	"http://web.com":    true,
}

func mainGin() {
	// 连接到grpc服务的客户端
	conn, err := grpc.Dial("localhost:1111", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect : %v", err.Error())
	}
	defer conn.Close()
	clientCoin := pb.NewUserCoinClient(conn)
	clientGrade := pb.NewUserGradeClient(conn)

	var router = gin.New()
	router.GET("/hello", func(context *gin.Context) {
		context.String(http.StatusOK, "hello")
	})
	v1Group := router.Group("/v1", func(ctx *gin.Context) {
		// 支持跨域
		origin := ctx.GetHeader("Origin")
		if AllowOrigin[origin] {
			ctx.Header("Access-Control-Allow-Origin", origin)
			ctx.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTION")
			ctx.Header("Access-Control-Allow-Headers", "*")
			ctx.Header("Access-Control-Allow-Credentials", "true")
		}
		ctx.Next()
	})
	// 用户积分服务的方法
	gUserCoin := v1Group.Group("/v1/UserGrowth.UserCoin")
	gUserCoin.GET("/listTasks", func(context *gin.Context) {
		out, err := clientCoin.ListTasks(context, &pb.ListTasksRequest{})
		if err != nil {
			context.JSON(http.StatusInternalServerError, map[string]interface{}{
				"code": 2,
				"msg":  err.Error(),
			})
		} else {
			context.JSON(http.StatusOK, out)
		}
	})
	gUserCoin.POST("/UserCoinChange", func(context *gin.Context) {
		body := &pb.UserCoinChangeRequest{}
		err := context.BindJSON(body)
		if err != nil {
			context.JSON(http.StatusInternalServerError, map[string]interface{}{
				"code": 2,
				"msg":  err.Error(),
			})
		} else if out, err := clientCoin.UserCoinChange(context, body); err != nil {
			context.JSON(http.StatusInternalServerError, map[string]interface{}{
				"code": 2,
				"msg":  err.Error(),
			})
		} else {
			context.JSON(http.StatusOK, out)
		}
	})
	// 用户等级服务的方法
	gUserGrade := v1Group.Group("/v1/UserGrowth.UserGrade")
	gUserGrade.GET("/listGrades", func(context *gin.Context) {
		out, err := clientGrade.ListGrades(context, &pb.ListGradesRequest{})
		if err != nil {
			context.JSON(http.StatusInternalServerError, map[string]interface{}{
				"code": 2,
				"msg":  err.Error(),
			})
		} else {
			context.JSON(http.StatusOK, out)
		}
	})
	// 为Http/2配置参数
	h2Handler := h2c.NewHandler(router, &http2.Server{})
	// 配置http服务
	server := &http.Server{
		Addr:    ":9999",
		Handler: h2Handler,
	}
	// 启动http服务
	server.ListenAndServe()
}

func mainGateway() {
	// 初始化数据库实例
	initDb()
	s := grpc.NewServer()
	// 注册服务
	pb.RegisterUserCoinServer(s, &ugserver.UgCoinServer{})
	pb.RegisterUserGradeServer(s, &ugserver.UgGradeServer{})
	reflection.Register(s)
	// grpc-gateway 注册服务
	serveMuxOpt := []runtime.ServeMuxOption{
		runtime.WithOutgoingHeaderMatcher(func(s string) (string, bool) {
			return s, true
		}),
		runtime.WithMetadata(func(ctx context.Context, request *http.Request) metadata.MD {
			origin := request.Header.Get("Origin")
			if AllowOrigin[origin] {
				md := metadata.New(map[string]string{
					"Access-Control-Allow-Origin":      origin,
					"Access-Control-Allow-Methods":     "GET,POST,PUT,DELETE,OPTION",
					"Access-Control-Allow-Headers":     "*",
					"Access-Control-Allow-Credentials": "true",
				})
				grpc.SetHeader(ctx, md)
			}
			return nil
		}),
	}
	mux := runtime.NewServeMux(serveMuxOpt...)
	ctx := context.Background()
	if err := pb.RegisterUserCoinHandlerServer(ctx, mux, &ugserver.UgCoinServer{}); err != nil {
		log.Printf("Faile to RegisterUserCoinHandlerServer error=%v", err)
	}
	if err := pb.RegisterUserGradeHandlerServer(ctx, mux, &ugserver.UgGradeServer{}); err != nil {
		log.Printf("Faile to RegisterUserGradeHandlerServer error=%v", err)
	}
	httpMux := http.NewServeMux()
	httpMux.Handle("/v1/UserGrowth", mux)
	// 配置http服务
	server := &http.Server{
		Addr: ":8081",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("http.HandlerFunc url=%s", r.URL)
			mux.ServeHTTP(w, r)
		}),
	}
	// 启动http服务
	log.Printf("server.ListenAdnServe(%s)", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("ListenAndServe error=%v", err)
	}
}

func main() {
	mainGateway()
	//mainGin()
}
