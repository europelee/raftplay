package utils

import (
	"net"
	"net/http"
	"strconv"

	restful "github.com/emicklei/go-restful"
)

var (
	APIServerInst *APIServer
)

// APIServer
type APIServer struct {
	*restful.Container
}

func NewAPIServer() *APIServer {
	APIServerInst = &APIServer{
		Container: restful.NewContainer(),
	}
	return APIServerInst
}

// ListenAndServeAPIServer initializes a server to respond to HTTP network requests on rdsync.
func ListenAndServeAPIServer(
	ip string,
	port uint) {
	s := &http.Server{
		Addr:           net.JoinHostPort(ip, strconv.FormatUint(uint64(port), 10)),
		Handler:        APIServerInst,
		MaxHeaderBytes: 1 << 20,
	}
	s.ListenAndServe()
}
