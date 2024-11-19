package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"google.golang.org/grpc"
	pb "ywc/cacheserver/cache"
)

var server cacheServer // server instace
var address [4]string
var client [2]pb.CacheClient // 2 rpc client to communicate with the other 2 rpc server
var conn [2]*grpc.ClientConn // 2 connection for 2 rpc client

func setAddress() {
	if os.Args[1] == "1" { // set address variable by server index
		address[0] = "127.0.0.1:9527" // this server's http server port
		address[1] = "127.0.0.1:9550" // this server's rpc server port

		address[2] = "127.0.0.1:9551" // another server's rpc server port
		address[3] = "127.0.0.1:9552" // another server's rpc server port
	} else if os.Args[1] == "2" {
		address[0] = "127.0.0.1:9528"
		address[1] = "127.0.0.1:9551"

		address[2] = "127.0.0.1:9550"
		address[3] = "127.0.0.1:9552"
	} else if os.Args[1] == "3" {
		address[0] = "127.0.0.1:9529"
		address[1] = "127.0.0.1:9552"

		address[2] = "127.0.0.1:9550"
		address[3] = "127.0.0.1:9551"
	} else {
		fmt.Println("only 3 cacheserver.")
	}
}


// // http Get handler
// func handleGet(w http.ResponseWriter, key string) {
// 	fmt.Println("get", key)

// 	// 检查缓存中是否存在指定的键
// 	if value, ok := server.cache[key]; ok {
// 		// 如果找到了缓存，设置响应头
// 		w.WriteHeader(http.StatusOK)
// 		w.Header().Set("Content-Type", "application/json")

// 		// 断言缓存值为 *GetReply 类型
// 		reply.Val
// 		if reply, ok := value.(&pb.GetReply); ok {
// 			// 根据 oneof 字段的类型处理不同的情况
// 			switch v := reply.GetValue().(type) {
// 			case *GetReply_IntValue:
// 				// 如果是 int32 类型
// 				fmt.Fprintf(w, "{\"%s\":%d}\n", key, v.IntValue)
// 			case *GetReply_StringValue:
// 				// 如果是 string 类型
// 				fmt.Fprintf(w, "{\"%s\":\"%s\"}\n", key, v.StringValue)
// 			case *GetReply_StringArray:
// 				// 如果是 StringArray 类型
// 				items := v.StringArray.Values
// 				// 将 StringArray 转换成 JSON 格式
// 				fmt.Fprintf(w, "{\"%s\":%v}\n", key, items)
// 			default:
// 				// 如果类型不匹配
// 				fmt.Println("Unknown type in oneof field")
// 				w.WriteHeader(http.StatusInternalServerError)
// 			}
// 		} else {
// 			// 如果缓存值不是 *GetReply 类型
// 			fmt.Println("Cached value is not a GetReply")
// 			w.WriteHeader(http.StatusInternalServerError)
// 		}
// 		return
// 	} else {
// 		// 如果缓存中没有该键，返回 404 错误
// 		w.WriteHeader(http.StatusNotFound)
// 	}
// }


// http Get handler
func handleGet(w http.ResponseWriter, key string) {
	fmt.Println("get", key)

	if _, ok := server.cache[key]; ok {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		cachedValue, ok := server.cache[key].(string)
		if !ok {
			fmt.Println("Cached value is not a string")
			return
		}
		fmt.Fprintln(w, "{\""+key+"\":\""+cachedValue+"\"}")
		return
	} else //if we can't find in this machine ,we search the another two
	{
		w.WriteHeader(http.StatusNotFound)
	}
}

// func handleGet(w http.ResponseWriter, key string) {
// 	fmt.Println("get", key)

// 	if _, ok := server.cache[key]; ok {
// 		w.WriteHeader(http.StatusOK)
// 		w.Header().Set("Content-Type", "application/json")
// 		cachedValue, ok := server.cache[key].(string)
// 		if !ok {
// 			fmt.Println("Cached value is not a string")
// 			return
// 		}
// 		fmt.Fprintln(w, "{\""+key+"\":\""+cachedValue+"\"}")
// 		return
// 	} else //if we can't find in this machine ,we search the another two
// 	{
// 		val1, err1 := CacheGet(client[0], &pb.GetRequest{Key: key})
// 		if err1 != nil {
// 			val2, err2 := CacheGet(client[1], &pb.GetRequest{Key: key})
// 			if err2 != nil {
// 				w.WriteHeader(http.StatusNotFound)
// 			} else {
// 				w.WriteHeader(http.StatusOK)
// 				w.Header().Set("Content-Type", "application/json")
// 				cachedValue, ok := val2.(string)
// 				if !ok {
// 					fmt.Println("Cached value is not a string")
// 					return
// 				}
// 				fmt.Fprintln(w, "{\""+key+"\":\""+cachedValue+"\"}")
// 				return
// 			}
// 		} else {
// 			w.WriteHeader(http.StatusOK)
// 			w.Header().Set("Content-Type", "application/json")
// 			cachedValue, ok := val1.(string)
// 			if !ok {
// 				fmt.Println("Cached value is not a string")
// 				return
// 			}
// 			fmt.Fprintln(w, "{\""+key+"\":\""+cachedValue+"\"}")
// 			return
// 		}
// 	}
// }


// http Set handler
func handleSet(w http.ResponseWriter, jsonstr string) {

	// reg := regexp.MustCompile(`{\s*"(.*)"\s*:\s*"(.*)"\s*}`)
	// reg := regexp.MustCompile(`{\s*"(.*?)"\s*:\s*(\[(.*?)\]|"(.*?)"|(\d+))\s*}`)
	reg := regexp.MustCompile(`{\s*"(.*)"\s*:\s*"(.*)"|\[(.*)\]|(\d+)\s*}`)
	if reg == nil {
		fmt.Println("regexp err")
		return
	}
	result := reg.FindAllStringSubmatch(jsonstr, -1)
	key, value := result[0][1], result[0][2]

	fmt.Println("set", key, ":", value)

	server.cache[key] = value
	CacheSet(client[0], &pb.SetRequest{Key: key, Value: value})
    CacheSet(client[1], &pb.SetRequest{Key: key, Value: value})

	w.WriteHeader(http.StatusOK)

}

// http Delete handler
func handleDelete(w http.ResponseWriter, key string) {
	fmt.Println("delete", key)
	if _, ok := server.cache[key]; ok {
        delete(server.cache, key)
        CacheDelete(client[0], &pb.DeleteRequest{Key: key})
        CacheDelete(client[1], &pb.DeleteRequest{Key: key})

        w.WriteHeader(http.StatusOK)
        fmt.Fprintln(w, "1")
        return
    }

    w.WriteHeader(http.StatusOK)
    fmt.Fprintln(w, "0")
	// if _, ok := server.cache[key]; ok {
	// 	delete(server.cache, key)
	// 	w.WriteHeader(http.StatusOK)
	// 	fmt.Fprintln(w, "1") //delete successfully
	// 	return
	// } else{
	//     err1 := CacheDelete(client[0],&pb.DeleteRequest{Key:key})
	//     if err1 != nil{
	//         err2 := CacheDelete(client[1],&pb.DeleteRequest{Key:key})
	//         if err2 != nil{
	//             w.WriteHeader(http.StatusOK)
	//             fmt.Fprintln(w, "0")//no need to delete
	//             return
	//         } else {
	//             w.WriteHeader(http.StatusOK)
	//             fmt.Fprintln(w, "1")//delete successfully
	//             return
	//         }
	//     } else {
	//         w.WriteHeader(http.StatusOK)
	//         fmt.Fprintln(w, "1")//delete successfully
	//         return
	//     }
	// }

}

// http request handler
func handleHttpRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		handleGet(w, r.URL.String()[1:])
	} else if r.Method == http.MethodPost {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Unable to read request body.", http.StatusInternalServerError)
			return
		}
		handleSet(w, string(body))
	} else if r.Method == http.MethodDelete {
		handleDelete(w, r.URL.String()[1:])
	} else {
		http.Error(w, "Unsupport http request.", http.StatusMethodNotAllowed)
	}
}

// cacheserver type
type cacheServer struct {
	pb.UnimplementedCacheServer
	cache map[string]interface{}
}

// rpc server Get handler
func (s *cacheServer) GetCache(ctx context.Context, req *pb.GetRequest) (*pb.GetReply, error) {
	if val, ok := s.cache[req.Key]; ok {
		reply := &pb.GetReply{Key: req.Key}
		
		// 根据值的类型设置相应的 oneof 字段
		switch v := val.(type) {
		case int32:
			reply.Value = &pb.GetReply_IntValue{IntValue: v}
		case string:
			reply.Value = &pb.GetReply_StringValue{StringValue: v}
		case []string:
			reply.Value = &pb.GetReply_StringArray{StringArray: &pb.StringArray{Values: v}}
		default:
			return nil, fmt.Errorf("unsupported value type")
		}
		
		return reply, nil
	}
	return nil, fmt.Errorf("key not found in cache")
}

// rpc server Set handler
func (s *cacheServer) SetCache (ctx context.Context, req *pb.SetRequest) (*pb.SetReply, error) {
    s.cache[req.Key] = req.Value
    return &pb.SetReply{}, nil
}

// rpc server Delete handler
func (s *cacheServer) DeleteCache(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteReply, error) {
	if _, ok := s.cache[req.Key]; ok {
		delete(s.cache, req.Key)
		return &pb.DeleteReply{Num: 1}, nil
	}
	return &pb.DeleteReply{Num: 0}, nil
}

func startHttpServer() {
	http.HandleFunc("/", handleHttpRequest)
	fmt.Println("Listening http on", address[0])
	err := http.ListenAndServe(address[0], nil)
	if err != nil {
		fmt.Println("Listten failed:", err)
	}
}

func startRpcServer() {
	fmt.Println("Listening rpc on", address[1])
	lis, err := net.Listen("tcp", address[1])
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	server = cacheServer{cache: make(map[string]interface{})}
	pb.RegisterCacheServer(grpcServer, &server)
	grpcServer.Serve(lis)
}



func main() {
	if len(os.Args) != 2 {
		fmt.Println("please specify server index(1-3).")
		return
	}

	setAddress()
	go startHttpServer()
	go startRpcServer()
	setupClient()

	select {}
}
