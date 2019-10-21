package raftoperation

import (
	"fmt"
	"net/http"

	"github.com/europelee/raftplay/pkg/election"
	utils "github.com/europelee/raftplay/pkg/utils"

	restful "github.com/emicklei/go-restful"
)

type Node struct {
	ID string `json:"id"`
}

var (
	ElecOpHandler *Handler
)

const (
	operPath = "/operations"
)

func CreateHandler(o election.Operation) *Handler {
	ElecOpHandler = &Handler{operation: o}
	return ElecOpHandler
}

// Handler as resource
type Handler struct {
	operation election.Operation
}

func (h *Handler) Register(server *utils.APIServer) {
	ws := &restful.WebService{}
	ws.Path(operPath).
		Produces(restful.MIME_JSON).
		Consumes(restful.MIME_JSON)
	endpoints := []struct {
		path    string
		method  string
		handler restful.RouteFunction
	}{
		{"/leaderShip", "POST", h.transferLeaderShip},
		{"/addVoter", "POST", h.addVoter},
		{"/remSvr", "POST", h.remServer},
	}

	for _, e := range endpoints {
		ws.Route(ws.
			Method(e.method).
			Path(e.path).
			To(e.handler))
	}
	server.Add(ws)
}

func (h *Handler) transferLeaderShip(request *restful.Request, response *restful.Response) {
	node := new(Node)
	err := request.ReadEntity(&node)
	if err == nil {
		err = h.operation.TransferLeaderShip(node.ID, node.ID)
		if err == nil {
			response.WriteHeaderAndEntity(http.StatusCreated, node)
		}

	}
	if err != nil {
		fmt.Println(err.Error())
		response.WriteErrorString(http.StatusBadRequest, err.Error())
	}
}

func (h *Handler) addVoter(request *restful.Request, response *restful.Response) {
	node := new(Node)
	err := request.ReadEntity(&node)
	if err == nil {
		err = h.operation.Join(node.ID, node.ID)
		if err == nil {
			response.WriteHeaderAndEntity(http.StatusCreated, node)
		}

	}
	if err != nil {
		fmt.Println(err.Error())
		response.WriteErrorString(http.StatusBadRequest, err.Error())
	}
}

func (h *Handler) remServer(request *restful.Request, response *restful.Response) {
	node := new(Node)
	err := request.ReadEntity(&node)
	if err == nil {
		err = h.operation.RemoveSvr(node.ID)
		if err == nil {
			response.WriteHeaderAndEntity(http.StatusCreated, node)
		}

	}
	if err != nil {
		fmt.Println(err.Error())
		response.WriteErrorString(http.StatusBadRequest, err.Error())
	}
}
