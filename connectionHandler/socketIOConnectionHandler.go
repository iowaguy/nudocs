package connectionHandler

import (
	"github.com/googollee/go-socket.io"
	"log"
	"net/http"
	"strconv"
)

const (
	EMIT_TO_CLIENT      = "server-event"
	RECEIVE_FROM_CLIENT = "client-event"
)

var ValueToEmit chan string

type Server struct {
	socketio.Server
}

func newServer(transportNames []string) (*Server, error) {
	ret, err := socketio.NewServer(transportNames)
	if err != nil {
		return nil, err
	}
	return &Server{*ret}, nil
}

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	OriginList := r.Header["Origin"]
	Origin := ""
	if len(OriginList) > 0 {
		Origin = OriginList[0]
	}
	w.Header().Add("Access-Control-Allow-Origin", Origin)
	w.Header().Add("Access-Control-Allow-Credentials", "true")
	s.Server.ServeHTTP(w, r)
}

type connectionHandler func(string, socketio.Socket)
type eventHandler func(string)

func AcceptSocketIoConnectionOn(port int, cHandler connectionHandler, eHandler eventHandler) {
	portStr := strconv.Itoa(port)
	server, err := newServer(nil)
	if err != nil {
		log.Println("Socker.io connection failed")
		log.Fatal(err)
	}
	server.On("connection", func(so socketio.Socket) {
		cHandler("connection", so)
		so.On(RECEIVE_FROM_CLIENT, func(msg string) {
			eHandler(msg)
		})
		so.On("disconnection", func() {
			cHandler("disconnection", so)
		})
	})
	server.On("error", func(so socketio.Socket, err error) {
		cHandler("error", so)
	})
	http.Handle("/", server)
	log.Fatal(http.ListenAndServe(":"+portStr, nil))
}
