package stats

import (
	"fmt"
	"net/http"

	"github.com/europelee/raftplay/pkg/election"
	utils "github.com/europelee/raftplay/pkg/utils"

	restful "github.com/emicklei/go-restful"
)

var (
	StatHandler *Handler
)

const (
	statsPath = "/stats"
)

func CreateHandler(o election.Operation) *Handler {
	StatHandler = &Handler{operation: o}
	return StatHandler
}

// Handler as resource
type Handler struct {
	operation election.Operation
}

func (h *Handler) Register(server *utils.APIServer) {
	ws := &restful.WebService{}
	ws.Path(statsPath).
		Produces(restful.MIME_JSON).
		Consumes(restful.MIME_JSON)
	endpoints := []struct {
		path    string
		method  string
		handler restful.RouteFunction
	}{
		{"/keyWatch", "POST", h.createKeyWatch},
		{"/keyWatch/{db}/{key}", "DELETE", h.removeKeyWatch},
		{"/keyWatch/{db}/{key}", "GET", h.queryKeyWatch},
	}

	for _, e := range endpoints {
		ws.Route(ws.
			Method(e.method).
			Path(e.path).
			To(e.handler))
	}
	server.Add(ws)
}

func (h *Handler) createKeyWatch(request *restful.Request, response *restful.Response) {
	utils.Logger.Info(fmt.Sprintf("createKeyWatch[%v]", request))
	key := new(Key)
	err := request.ReadEntity(&key)
	if err == nil {
		WatchObj.RegisterKeyWatch(key.String())
		response.WriteHeaderAndEntity(http.StatusCreated, key)
	} else {
		utils.Logger.Error(err.Error())
		response.WriteErrorString(http.StatusBadRequest, err.Error())
	}
}
