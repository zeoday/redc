package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	redc "red-cloud/mod"
	"red-cloud/mod/gologger"
	"red-cloud/utils/sshutil"
	"strings"

	"github.com/spf13/cobra"
)

// MCP Protocol Version
const MCPVersion = "2024-11-05"

// MCP Message types
type MCPRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

type MCPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// MCP Server capabilities and info
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Implementation struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type InitializeResult struct {
	ProtocolVersion string          `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      ServerInfo      `json:"serverInfo"`
}

type ServerCapabilities struct {
	Tools     *ToolsCapability     `json:"tools,omitempty"`
	Resources *ResourcesCapability `json:"resources,omitempty"`
}

type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

type ResourcesCapability struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}

// Tool definitions
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema ToolSchema  `json:"inputSchema"`
}

type ToolSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required,omitempty"`
}

type Property struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
}

// Tool call parameters
type CallToolParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

type ToolResult struct {
	Content []ContentItem `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

type ContentItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Resource definitions
type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

type ResourceContents struct {
	URI      string        `json:"uri"`
	MimeType string        `json:"mimeType,omitempty"`
	Contents []ContentItem `json:"contents"`
}

// MCP Server
type MCPServer struct {
	project *redc.RedcProject
}

func NewMCPServer(project *redc.RedcProject) *MCPServer {
	return &MCPServer{
		project: project,
	}
}

func (s *MCPServer) HandleRequest(req *MCPRequest) *MCPResponse {
	resp := &MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
	}

	switch req.Method {
	case "initialize":
		resp.Result = InitializeResult{
			ProtocolVersion: MCPVersion,
			Capabilities: ServerCapabilities{
				Tools: &ToolsCapability{
					ListChanged: false,
				},
				Resources: &ResourcesCapability{
					Subscribe:   false,
					ListChanged: false,
				},
			},
			ServerInfo: ServerInfo{
				Name:    "redc",
				Version: redc.Version,
			},
		}

	case "tools/list":
		resp.Result = map[string]interface{}{
			"tools": s.getTools(),
		}

	case "tools/call":
		var params CallToolParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			resp.Error = &MCPError{
				Code:    -32602,
				Message: "Invalid params",
				Data:    err.Error(),
			}
		} else {
			result, err := s.executeTool(params.Name, params.Arguments)
			if err != nil {
				resp.Result = ToolResult{
					Content: []ContentItem{{
						Type: "text",
						Text: fmt.Sprintf("Error: %v", err),
					}},
					IsError: true,
				}
			} else {
				resp.Result = result
			}
		}

	case "resources/list":
		resp.Result = map[string]interface{}{
			"resources": s.getResources(),
		}

	case "resources/read":
		var params struct {
			URI string `json:"uri"`
		}
		if err := json.Unmarshal(req.Params, &params); err != nil {
			resp.Error = &MCPError{
				Code:    -32602,
				Message: "Invalid params",
				Data:    err.Error(),
			}
		} else {
			result, err := s.readResource(params.URI)
			if err != nil {
				resp.Error = &MCPError{
					Code:    -32603,
					Message: "Failed to read resource",
					Data:    err.Error(),
				}
			} else {
				resp.Result = result
			}
		}

	case "ping":
		resp.Result = map[string]interface{}{}

	default:
		resp.Error = &MCPError{
			Code:    -32601,
			Message: "Method not found",
		}
	}

	return resp
}

func (s *MCPServer) getTools() []Tool {
	return []Tool{
		{
			Name:        "list_templates",
			Description: "List all available redc templates/images",
			InputSchema: ToolSchema{
				Type:       "object",
				Properties: map[string]Property{},
			},
		},
		{
			Name:        "list_cases",
			Description: "List all running cases in the current project",
			InputSchema: ToolSchema{
				Type:       "object",
				Properties: map[string]Property{},
			},
		},
		{
			Name:        "create_case",
			Description: "Create a new case from a template",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]Property{
					"template": {
						Type:        "string",
						Description: "Template name (e.g., 'aliyun/ecs')",
					},
					"name": {
						Type:        "string",
						Description: "Case name (optional, auto-generated if not provided)",
					},
					"env": {
						Type:        "object",
						Description: "Environment variables for the template (optional)",
					},
				},
				Required: []string{"template"},
			},
		},
		{
			Name:        "start_case",
			Description: "Start a case by ID",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]Property{
					"case_id": {
						Type:        "string",
						Description: "Case ID to start",
					},
				},
				Required: []string{"case_id"},
			},
		},
		{
			Name:        "stop_case",
			Description: "Stop a running case by ID",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]Property{
					"case_id": {
						Type:        "string",
						Description: "Case ID to stop",
					},
				},
				Required: []string{"case_id"},
			},
		},
		{
			Name:        "kill_case",
			Description: "Kill (destroy) a case by ID",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]Property{
					"case_id": {
						Type:        "string",
						Description: "Case ID to kill",
					},
				},
				Required: []string{"case_id"},
			},
		},
		{
			Name:        "get_case_status",
			Description: "Get the status of a specific case",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]Property{
					"case_id": {
						Type:        "string",
						Description: "Case ID to check",
					},
				},
				Required: []string{"case_id"},
			},
		},
		{
			Name:        "exec_command",
			Description: "Execute a command on a case",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]Property{
					"case_id": {
						Type:        "string",
						Description: "Case ID to execute command on",
					},
					"command": {
						Type:        "string",
						Description: "Command to execute",
					},
				},
				Required: []string{"case_id", "command"},
			},
		},
	}
}

func (s *MCPServer) getResources() []Resource {
	return []Resource{
		{
			URI:         "redc://templates",
			Name:        "Available Templates",
			Description: "List of all available redc templates",
			MimeType:    "application/json",
		},
		{
			URI:         "redc://cases",
			Name:        "Running Cases",
			Description: "List of all cases in the current project",
			MimeType:    "application/json",
		},
		{
			URI:         "redc://config",
			Name:        "Configuration",
			Description: "Current redc configuration",
			MimeType:    "application/json",
		},
	}
}

func (s *MCPServer) readResource(uri string) (interface{}, error) {
	switch uri {
	case "redc://templates":
		dirs, err := redc.ScanTemplateDirs(redc.TemplateDir, redc.MaxTfDepth)
		if err != nil {
			return nil, err
		}
		data, _ := json.MarshalIndent(dirs, "", "  ")
		return ResourceContents{
			URI:      uri,
			MimeType: "application/json",
			Contents: []ContentItem{{
				Type: "text",
				Text: string(data),
			}},
		}, nil

	case "redc://cases":
		cases, err := redc.LoadProjectCases(s.project.ProjectName)
		if err != nil {
			return nil, err
		}
		caseList := make([]map[string]string, 0)
		for _, c := range cases {
			caseList = append(caseList, map[string]string{
				"id":     c.GetId(),
				"name":   c.Name,
				"status": c.State,
				"type":   c.Type,
			})
		}
		data, _ := json.MarshalIndent(caseList, "", "  ")
		return ResourceContents{
			URI:      uri,
			MimeType: "application/json",
			Contents: []ContentItem{{
				Type: "text",
				Text: string(data),
			}},
		}, nil

	case "redc://config":
		config := map[string]interface{}{
			"project":      redc.Project,
			"user":         redc.U,
			"template_dir": redc.TemplateDir,
			"redc_path":    redc.RedcPath,
		}
		data, _ := json.MarshalIndent(config, "", "  ")
		return ResourceContents{
			URI:      uri,
			MimeType: "application/json",
			Contents: []ContentItem{{
				Type: "text",
				Text: string(data),
			}},
		}, nil

	default:
		return nil, fmt.Errorf("unknown resource URI: %s", uri)
	}
}

func (s *MCPServer) executeTool(name string, args map[string]interface{}) (ToolResult, error) {
	switch name {
	case "list_templates":
		return s.toolListTemplates()

	case "list_cases":
		return s.toolListCases()

	case "create_case":
		template, ok := args["template"].(string)
		if !ok {
			return ToolResult{}, fmt.Errorf("missing or invalid 'template' parameter")
		}
		caseName, _ := args["name"].(string)
		env, _ := args["env"].(map[string]interface{})
		return s.toolCreateCase(template, caseName, env)

	case "start_case":
		caseID, ok := args["case_id"].(string)
		if !ok {
			return ToolResult{}, fmt.Errorf("missing or invalid 'case_id' parameter")
		}
		return s.toolStartCase(caseID)

	case "stop_case":
		caseID, ok := args["case_id"].(string)
		if !ok {
			return ToolResult{}, fmt.Errorf("missing or invalid 'case_id' parameter")
		}
		return s.toolStopCase(caseID)

	case "kill_case":
		caseID, ok := args["case_id"].(string)
		if !ok {
			return ToolResult{}, fmt.Errorf("missing or invalid 'case_id' parameter")
		}
		return s.toolKillCase(caseID)

	case "get_case_status":
		caseID, ok := args["case_id"].(string)
		if !ok {
			return ToolResult{}, fmt.Errorf("missing or invalid 'case_id' parameter")
		}
		return s.toolGetCaseStatus(caseID)

	case "exec_command":
		caseID, ok := args["case_id"].(string)
		if !ok {
			return ToolResult{}, fmt.Errorf("missing or invalid 'case_id' parameter")
		}
		command, ok := args["command"].(string)
		if !ok {
			return ToolResult{}, fmt.Errorf("missing or invalid 'command' parameter")
		}
		return s.toolExecCommand(caseID, command)

	default:
		return ToolResult{}, fmt.Errorf("unknown tool: %s", name)
	}
}

// Tool implementations
func (s *MCPServer) toolListTemplates() (ToolResult, error) {
	dirs, err := redc.ScanTemplateDirs(redc.TemplateDir, redc.MaxTfDepth)
	if err != nil {
		return ToolResult{}, err
	}

	output := "Available templates:\n"
	for _, dir := range dirs {
		output += fmt.Sprintf("- %s\n", dir)
	}

	return ToolResult{
		Content: []ContentItem{{
			Type: "text",
			Text: output,
		}},
	}, nil
}

func (s *MCPServer) toolListCases() (ToolResult, error) {
	cases, err := redc.LoadProjectCases(s.project.ProjectName)
	if err != nil {
		return ToolResult{}, err
	}
	
	output := fmt.Sprintf("Cases in project '%s':\n", redc.Project)
	
	if len(cases) == 0 {
		output += "No cases found.\n"
	} else {
		for _, c := range cases {
			output += fmt.Sprintf("- ID: %s, Name: %s, Status: %s, Type: %s\n", c.GetId(), c.Name, c.State, c.Type)
		}
	}

	return ToolResult{
		Content: []ContentItem{{
			Type: "text",
			Text: output,
		}},
	}, nil
}

func (s *MCPServer) toolCreateCase(template string, name string, env map[string]interface{}) (ToolResult, error) {
	// Create case using the RedcProject.CaseCreate method
	vars := make(map[string]string)
	if env != nil {
		for k, v := range env {
			vars[k] = fmt.Sprintf("%v", v)
		}
	}
	
	c, err := s.project.CaseCreate(template, redc.U, name, vars)
	if err != nil {
		return ToolResult{}, fmt.Errorf("failed to create case: %v", err)
	}

	output := fmt.Sprintf("Case created successfully:\n")
	output += fmt.Sprintf("- ID: %s\n", c.GetId())
	output += fmt.Sprintf("- Name: %s\n", c.Name)
	output += fmt.Sprintf("- Template: %s\n", template)
	output += fmt.Sprintf("\nUse 'start_case' with ID '%s' to start the case.\n", c.GetId())

	return ToolResult{
		Content: []ContentItem{{
			Type: "text",
			Text: output,
		}},
	}, nil
}

func (s *MCPServer) toolStartCase(caseID string) (ToolResult, error) {
	c, err := s.project.GetCase(caseID)
	if err != nil {
		return ToolResult{}, fmt.Errorf("case not found: %v", err)
	}

	if err := c.TfApply(); err != nil {
		return ToolResult{}, fmt.Errorf("failed to start case: %v", err)
	}

	output := fmt.Sprintf("Case '%s' (%s) started successfully.\n", c.Name, c.GetId())
	return ToolResult{
		Content: []ContentItem{{
			Type: "text",
			Text: output,
		}},
	}, nil
}

func (s *MCPServer) toolStopCase(caseID string) (ToolResult, error) {
	c, err := s.project.GetCase(caseID)
	if err != nil {
		return ToolResult{}, fmt.Errorf("case not found: %v", err)
	}

	if err := c.Stop(); err != nil {
		return ToolResult{}, fmt.Errorf("failed to stop case: %v", err)
	}

	output := fmt.Sprintf("Case '%s' (%s) stopped successfully.\n", c.Name, c.GetId())
	return ToolResult{
		Content: []ContentItem{{
			Type: "text",
			Text: output,
		}},
	}, nil
}

func (s *MCPServer) toolKillCase(caseID string) (ToolResult, error) {
	c, err := s.project.GetCase(caseID)
	if err != nil {
		return ToolResult{}, fmt.Errorf("case not found: %v", err)
	}

	if err := c.Kill(); err != nil {
		return ToolResult{}, fmt.Errorf("failed to kill case: %v", err)
	}

	output := fmt.Sprintf("Case '%s' (%s) killed successfully.\n", c.Name, c.GetId())
	return ToolResult{
		Content: []ContentItem{{
			Type: "text",
			Text: output,
		}},
	}, nil
}

func (s *MCPServer) toolGetCaseStatus(caseID string) (ToolResult, error) {
	c, err := s.project.GetCase(caseID)
	if err != nil {
		return ToolResult{}, fmt.Errorf("case not found: %v", err)
	}

	output := fmt.Sprintf("Case Status:\n")
	output += fmt.Sprintf("- ID: %s\n", c.GetId())
	output += fmt.Sprintf("- Name: %s\n", c.Name)
	output += fmt.Sprintf("- Status: %s\n", c.State)
	output += fmt.Sprintf("- Template: %s\n", c.Type)

	return ToolResult{
		Content: []ContentItem{{
			Type: "text",
			Text: output,
		}},
	}, nil
}

func (s *MCPServer) toolExecCommand(caseID string, command string) (ToolResult, error) {
	c, err := s.project.GetCase(caseID)
	if err != nil {
		return ToolResult{}, fmt.Errorf("case not found: %v", err)
	}

	// Get SSH config and client
	sshConfig, err := c.GetSSHConfig()
	if err != nil {
		return ToolResult{}, fmt.Errorf("failed to get SSH config: %v", err)
	}

	client, err := sshutil.NewClient(sshConfig)
	if err != nil {
		return ToolResult{}, fmt.Errorf("failed to create SSH client: %v", err)
	}
	defer client.Close()

	// Execute command and capture output
	var outputBuf strings.Builder
	session, err := client.Client.NewSession()
	if err != nil {
		return ToolResult{}, fmt.Errorf("failed to create SSH session: %v", err)
	}
	defer session.Close()

	session.Stdout = &outputBuf
	session.Stderr = &outputBuf

	if err := session.Run(command); err != nil {
		return ToolResult{}, fmt.Errorf("failed to execute command: %v\nOutput: %s", err, outputBuf.String())
	}

	output := fmt.Sprintf("Command executed on case '%s' (%s):\n", c.Name, c.GetId())
	output += fmt.Sprintf("\nOutput:\n%s", outputBuf.String())

	return ToolResult{
		Content: []ContentItem{{
			Type: "text",
			Text: output,
		}},
	}, nil
}

// STDIO transport
func runStdioServer(server *MCPServer) {
	decoder := json.NewDecoder(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	for {
		var req MCPRequest
		if err := decoder.Decode(&req); err != nil {
			if err == io.EOF {
				break
			}
			gologger.Error().Msgf("Failed to decode request: %v", err)
			continue
		}

		resp := server.HandleRequest(&req)
		if err := encoder.Encode(resp); err != nil {
			gologger.Error().Msgf("Failed to encode response: %v", err)
		}
	}
}

// SSE transport
func runSSEServer(server *MCPServer, addr string) {
	http.HandleFunc("/sse", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		// Handle POST requests with JSON-RPC messages
		if r.Method == "POST" {
			var req MCPRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			resp := server.HandleRequest(&req)
			data, _ := json.Marshal(resp)
			
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
			return
		}

		// Keep connection alive
		fmt.Fprintf(w, "data: {\"type\":\"connected\"}\n\n")
		flusher.Flush()
	})

	http.HandleFunc("/message", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req MCPRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resp := server.HandleRequest(&req)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "redc MCP Server v%s\n", redc.Version)
		fmt.Fprintf(w, "Endpoints:\n")
		fmt.Fprintf(w, "  POST /message - Send JSON-RPC messages\n")
		fmt.Fprintf(w, "  GET  /sse     - SSE endpoint (for streaming)\n")
	})

	gologger.Info().Msgf("🚀 MCP SSE Server listening on %s", addr)
	gologger.Info().Msgf("   POST endpoint: http://%s/message", addr)
	gologger.Info().Msgf("   SSE endpoint:  http://%s/sse", addr)
	
	if err := http.ListenAndServe(addr, nil); err != nil {
		gologger.Fatal().Msgf("Failed to start server: %v", err)
	}
}

// MCP commands
var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Model Context Protocol (MCP) server",
	Long: `Start an MCP server to expose redc functionality via the Model Context Protocol.
Supports both STDIO and SSE (Server-Sent Events) transport modes.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var mcpStdioCmd = &cobra.Command{
	Use:   "stdio",
	Short: "Start MCP server with STDIO transport",
	Long: `Start an MCP server using STDIO transport.
The server reads JSON-RPC requests from stdin and writes responses to stdout.
This mode is suitable for integration with AI assistants and tools.`,
	Run: func(cmd *cobra.Command, args []string) {
		server := NewMCPServer(redcProject)
		gologger.Info().Msg("🚀 Starting MCP STDIO Server...")
		gologger.Info().Msgf("   Protocol version: %s", MCPVersion)
		gologger.Info().Msgf("   Project: %s", redc.Project)
		runStdioServer(server)
	},
}

var mcpSSECmd = &cobra.Command{
	Use:   "sse [address]",
	Short: "Start MCP server with SSE transport",
	Long: `Start an MCP server using SSE (Server-Sent Events) transport.
The server listens on the specified address (default: localhost:8080).

Example:
  redc mcp sse localhost:8080
  redc mcp sse :9000`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		addr := "localhost:8080"
		if len(args) > 0 {
			addr = args[0]
			// Handle cases like ":8080" or "8080"
			if strings.HasPrefix(addr, ":") || !strings.Contains(addr, ":") {
				if strings.HasPrefix(addr, ":") {
					addr = "localhost" + addr
				} else {
					addr = "localhost:" + addr
				}
			}
		}

		server := NewMCPServer(redcProject)
		gologger.Info().Msgf("Protocol version: %s", MCPVersion)
		gologger.Info().Msgf("Project: %s", redc.Project)
		runSSEServer(server, addr)
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)
	mcpCmd.AddCommand(mcpStdioCmd)
	mcpCmd.AddCommand(mcpSSECmd)
}
