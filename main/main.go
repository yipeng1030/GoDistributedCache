package main

import (
	"GoDistributedCache"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func createGroup() *GoDistributedCache.Group {
	return GoDistributedCache.NewGroup("scores", 2<<10, GoDistributedCache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

func startCacheServer(addr string, dnsServiceName string, gee *GoDistributedCache.Group) {
	peers := GoDistributedCache.NewHTTPPool(addr)
	gee.RegisterPeers(peers)
	log.Println("GoDistributedCache is running at", addr)

	// 定时查询 DNS 动态更新 peers 列表
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			<-ticker.C
			ips, err := net.LookupHost(dnsServiceName)
			if err != nil {
				log.Printf("DNS lookup error for %s: %v", dnsServiceName, err)
				continue
			}
			var dynamicAddrs []string
			// 假设所有节点都在同一个端口，例如 8001
			for _, ip := range ips {
				peerAddr := fmt.Sprintf("http://%s:8001", ip)
				dynamicAddrs = append(dynamicAddrs, peerAddr)
			}
			peers.Set(dynamicAddrs...)
			log.Printf("Updated peers: %v", dynamicAddrs)
		}
	}()

	// addr[7:] 去掉 "http://" 前缀，作为监听地址
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

func startAPIServer(apiAddr string, gee *GoDistributedCache.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := gee.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(append(view.ByteSlice(), []byte("\n")...))
		}))
	// 新增 /peers 接口，返回当前 HTTPPool 中的 peer 信息
	http.Handle("/peers", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		output := gee.GetPeers()
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(output))
	}))
	log.Println("fontend server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

func main() {
	apiAddr := "http://0.0.0.0:9999"
	gee := createGroup()
	go startAPIServer(apiAddr, gee)

	// 假设使用 DNS 服务发现的域名，需在 k8s 中配置好 Headless Service
	dnsServiceName := "mycache-headless.default.svc.cluster.local"
	podIP := os.Getenv("MY_POD_IP")
	selfAddr := fmt.Sprintf("http://%s:8001", podIP)

	startCacheServer(selfAddr, dnsServiceName, gee)
}
