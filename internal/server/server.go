// Package server used to create http server and handle req/res
package server

import (
	"fmt"
	"net"
	"strconv"
	"sync/atomic"

	"github.com/IntDavydov/httpfromtcp/internal/request"
	"github.com/IntDavydov/httpfromtcp/internal/response"
)

type Handler func(w *response.Writer, req *request.Request)

type Server struct {
	addr     int
	handler  Handler
	listener net.Listener
	isClosed atomic.Bool
}

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return nil, err
	}

	server := &Server{
		addr:     port,
		handler:  handler,
		listener: listener,
	}

	go server.listen()

	return server, nil
}

func (s *Server) Close() error {
	s.isClosed.Store(true)
	if s.listener != nil {
		return s.listener.Close()
	}
	return s.listener.Close()
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.isClosed.Load() {
				return
			}

			fmt.Println("\nerror accepting connection: ", err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	fmt.Println("Connection from: ", conn.LocalAddr())
	defer conn.Close()

	w := response.NewWriter(conn)

	req, err := request.RequestFromReader(conn)
	if err != nil {
		response.HandleServerError(w, response.BadRequest, err)
		return
	}

	s.handler(w, req)
}
