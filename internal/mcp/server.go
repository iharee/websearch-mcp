package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
)

type ToolHandler func(ctx context.Context, args map[string]interface{}) (*ToolCallResult, error)

type registeredTool struct {
	Definition Tool
	Handler    ToolHandler
}

type Server struct {
	mu       sync.RWMutex
	tools    []registeredTool
	sessions map[string]bool
}

func New() *Server {
	return &Server{
		sessions: make(map[string]bool),
	}
}

func (s *Server) RegisterTool(def Tool, handler ToolHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tools = append(s.tools, registeredTool{Definition: def, Handler: handler})
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, nil, ErrParse, "failed to read request body")
		return
	}

	var req Request
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, nil, ErrParse, "invalid JSON")
		return
	}

	if req.JSONRPC != JSONRPCVersion {
		writeError(w, req.ID, ErrInvalid, "jsonrpc must be 2.0")
		return
	}

	sessionID := r.Header.Get("Mcp-Session-Id")

	// initialize is the only method allowed without a session
	if req.Method != "initialize" && sessionID != "" {
		s.mu.RLock()
		valid := s.sessions[sessionID]
		s.mu.RUnlock()
		if !valid {
			writeError(w, req.ID, ErrInvalid, "invalid or expired session")
			return
		}
	}

	var resp *Response

	switch req.Method {
	case "initialize":
		resp = s.handleInitialize(req.ID, req.Params, w)
	case "notifications/initialized":
		// notification has no id, no response
		w.WriteHeader(http.StatusAccepted)
		return
	case "tools/list":
		resp = s.handleToolsList(req.ID, sessionID)
	case "tools/call":
		resp = s.handleToolsCall(r.Context(), req.ID, req.Params)
	case "ping":
		resp = &Response{
			JSONRPC: JSONRPCVersion,
			ID:      req.ID,
			Result:  map[string]interface{}{},
		}
	default:
		writeError(w, req.ID, ErrMethod, fmt.Sprintf("unknown method: %s", req.Method))
		return
	}

	if resp == nil {
		return // handleInitialize sets up the response itself
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleInitialize(id json.RawMessage, params json.RawMessage, w http.ResponseWriter) *Response {
	var initParams InitializeParams
	if err := json.Unmarshal(params, &initParams); err != nil {
		writeError(w, id, ErrParams, "invalid initialize params")
		return nil
	}

	sessionID := uuid.New().String()

	s.mu.Lock()
	s.sessions[sessionID] = true
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Mcp-Session-Id", sessionID)

	result := InitializeResult{
		ProtocolVersion: ProtocolVersion,
		Capabilities: ServerCapabilities{
			Tools: &ToolsCapability{ListChanged: false},
		},
		ServerInfo: Implementation{
			Name:    "websearch-mcp-server",
			Version: "0.1.0",
		},
	}

	resp := &Response{
		JSONRPC: JSONRPCVersion,
		ID:      id,
		Result:  result,
	}

	json.NewEncoder(w).Encode(resp)
	return nil
}

func (s *Server) handleToolsList(id json.RawMessage, _ string) *Response {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tools := make([]Tool, len(s.tools))
	for i, t := range s.tools {
		tools[i] = t.Definition
	}

	return &Response{
		JSONRPC: JSONRPCVersion,
		ID:      id,
		Result:  ToolsListResult{Tools: tools},
	}
}

func (s *Server) handleToolsCall(ctx context.Context, id json.RawMessage, params json.RawMessage) *Response {
	var callParams ToolCallParams
	if err := json.Unmarshal(params, &callParams); err != nil {
		return &Response{
			JSONRPC: JSONRPCVersion,
			ID:      id,
			Error:   &RPCError{Code: ErrParams, Message: "invalid tools/call params"},
		}
	}

	s.mu.RLock()
	var handler ToolHandler
	for _, t := range s.tools {
		if t.Definition.Name == callParams.Name {
			handler = t.Handler
			break
		}
	}
	s.mu.RUnlock()

	if handler == nil {
		return &Response{
			JSONRPC: JSONRPCVersion,
			ID:      id,
			Result: ToolCallResult{
				Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("unknown tool: %s", callParams.Name)}},
				IsError: true,
			},
		}
	}

	result, err := handler(ctx, callParams.Arguments)
	if err != nil {
		log.Printf("tool %s error: %v", callParams.Name, err)
		return &Response{
			JSONRPC: JSONRPCVersion,
			ID:      id,
			Result: ToolCallResult{
				Content: []ContentItem{{Type: "text", Text: err.Error()}},
				IsError: true,
			},
		}
	}

	return &Response{
		JSONRPC: JSONRPCVersion,
		ID:      id,
		Result:  result,
	}
}

func writeError(w http.ResponseWriter, id json.RawMessage, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	resp := &Response{
		JSONRPC: JSONRPCVersion,
		ID:      id,
		Error:   &RPCError{Code: code, Message: message},
	}
	json.NewEncoder(w).Encode(resp)
}
