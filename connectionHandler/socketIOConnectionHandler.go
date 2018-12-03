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

func NewServer(transportNames []string) (*Server, error) {
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

func AcceptConnectionOn(port int) {
	portStr := strconv.Itoa(port)
	server, err := NewServer(nil)
	if err != nil {
		log.Println("Socker.io connection failed")
		log.Fatal(err)
	}
	server.On("connection", func(so socketio.Socket) {
		log.Println("on connection")
		so.On(RECEIVE_FROM_CLIENT, func(msg string) {
			//TODO: handle event
		})
		so.On("disconnection", func() {
			log.Println("on disconnect")
		})
	})
	server.On("error", func(so socketio.Socket, err error) {
		log.Println("error:", err)
	})
	http.Handle("/", server)
	log.Fatal(http.ListenAndServe(":"+portStr, nil))
}
