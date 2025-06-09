package mcpgenerator

import (
	"encoding/json"
	"github.com/centralmind/gateway/connectors"
	"github.com/centralmind/gateway/model"
	"github.com/centralmind/gateway/plugins"
	"github.com/centralmind/gateway/server"
	"golang.org/x/xerrors"
	"sync"
)

type MCPServer struct {
	server       *server.MCPServer
	connector    connectors.Connector
	tools        []model.Endpoint
	interceptors []plugins.Interceptor

	mu    sync.Mutex
	plugs map[string]any
}

func New(
	plugs map[string]any,
) (*MCPServer, error) {
	srv := server.NewMCPServer("mcp-data-gateway", "0.0.1")
	interceptors, err := plugins.Plugins[plugins.Interceptor](plugs)
	if err != nil {
		return nil, xerrors.Errorf("unable to init interceptors: %w", err)
	}
	return &MCPServer{
		server:       srv,
		connector:    nil,
		plugs:        plugs,
		interceptors: interceptors,
	}, nil
}

func (s *MCPServer) SetConnector(connector connectors.Connector) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var err error
	connector, err = plugins.Wrap(s.plugs, connector)
	if err != nil {
		return xerrors.Errorf("unable to init connector plugins: %w", err)
	}
	s.connector = connector
	return nil
}

func (s *MCPServer) ServeSSE(addr string, prefix string, origins []string) *server.SSEServer {
	return server.NewSSEServer(s.server, addr, prefix, origins...)
}

func (s *MCPServer) ServeStdio() *server.StdioServer {
	return server.NewStdioServer(s.server)
}

func (s *MCPServer) Server() *server.MCPServer {
	return s.server
}

func jsonify(data any) string {
	res, _ := json.Marshal(data)
	return string(res)
}
