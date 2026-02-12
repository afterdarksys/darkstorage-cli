package ipc

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
)

type HandlerFunc func(data json.RawMessage) (*Response, error)

type Server struct {
	socketPath string
	listener   net.Listener
	handlers   map[string]HandlerFunc
	running    bool
}

func NewServer(socketPath string) *Server {
	return &Server{
		socketPath: socketPath,
		handlers:   make(map[string]HandlerFunc),
	}
}

func (s *Server) RegisterHandler(command string, handler HandlerFunc) {
	s.handlers[command] = handler
}

func (s *Server) Start() error {
	os.Remove(s.socketPath)

	listener, err := net.Listen("unix", s.socketPath)
	if err != nil {
		return err
	}

	if err := os.Chmod(s.socketPath, 0600); err != nil {
		listener.Close()
		return err
	}

	s.listener = listener
	s.running = true

	go s.acceptLoop()
	return nil
}

func (s *Server) Stop() error {
	s.running = false
	if s.listener != nil {
		s.listener.Close()
		os.Remove(s.socketPath)
	}
	return nil
}

func (s *Server) acceptLoop() {
	for s.running {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.running {
				fmt.Printf("Accept error: %v\n", err)
			}
			continue
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	var cmd Command
	if err := decoder.Decode(&cmd); err != nil {
		if err != io.EOF {
			encoder.Encode(&Response{
				Success: false,
				Error:   fmt.Sprintf("decode error: %v", err),
			})
		}
		return
	}

	handler, ok := s.handlers[cmd.Type]
	if !ok {
		encoder.Encode(&Response{
			Success: false,
			Error:   fmt.Sprintf("unknown command: %s", cmd.Type),
		})
		return
	}

	response, err := handler(cmd.Data)
	if err != nil {
		encoder.Encode(&Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	encoder.Encode(response)
}
