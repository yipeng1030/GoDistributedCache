package GoDistributedCache

import (
	pb "GoDistributedCache/cachepb"
	"GoDistributedCache/consistenthash"
	"fmt"
	"github.com/golang/protobuf/proto"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/_mycache/"
	defaultReplicas = 50
)

// HTTPPool implements PeerPicker for a pool of HTTP peers.
type HTTPPool struct {
	// this peer's base URL, e.g. "https://example.net:8000"
	self string
	// base path for all requests, e.g. "/_mycache/"
	basePath string
	mu       sync.RWMutex // guards peers and httpGetters
	peers    *consistenthash.HashNodes
	// httpGetters maps remote peer to its HTTPGetter keyed by e.g. "http://10.0.0.2:8008"
	httpGetters map[string]*httpGetter
	PeerPicker
}

// NewHTTPPool initializes an HTTP pool of peers.
func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

// Log info with server name
func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

// ServeHTTP handles all HTTP requests.
// 除了原有的 /<group>/<key> 接口外，新增了 /peers 接口用于输出所有 peers 的名字及 IP 地址列表
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)
	// 去掉 basePath 得到实际的路径
	path := r.URL.Path[len(p.basePath):]

	// 新增 /peers 路由，输出 peers 的名字和 IP 地址
	if path == "peers" || path == "peers/" {
		p.mu.RLock()
		defer p.mu.RUnlock()
		var output strings.Builder
		for peer := range p.httpGetters {
			// 解析 URL
			parsedURL, err := url.Parse(peer)
			if err != nil {
				output.WriteString(fmt.Sprintf("Peer: %s (invalid URL)\n", peer))
				continue
			}
			// 提取 host 部分（可能包含端口）
			host := parsedURL.Host
			// 拆分 IP 和端口
			ip, port, err := net.SplitHostPort(host)
			if err != nil {
				// 如果拆分失败，则直接输出 host
				output.WriteString(fmt.Sprintf("Peer: %s, Host: %s\n", peer, host))
			} else {
				output.WriteString(fmt.Sprintf("Peer: %s, IP: %s, Port: %s\n", peer, ip, port))
			}
		}
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(output.String()))
		return
	}

	// 处理 /<group>/<key> 请求
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	body, err := proto.Marshal(&pb.Response{Value: view.ByteSlice()})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(body)
}

// Set updates the pool's list of peers.
func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.NewHashNodes(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

// PickPeer picks a peer according to key.
func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}

type httpGetter struct {
	baseURL string
	PeerGetter
}

func (h *httpGetter) Get(in *pb.Request, out *pb.Response) error {
	u := fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		url.QueryEscape(in.GetGroup()),
		url.QueryEscape(in.GetKey()),
	)
	res, err := http.Get(u)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %v", err)
	}

	if err = proto.Unmarshal(bytes, out); err != nil {
		return fmt.Errorf("decoding response body: %v", err)
	}

	return nil
}
