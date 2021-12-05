package types

type Topology struct {
	ID          string       `json:"id"`
	Nodes       []Node       `json:"nodes"`
	Connections []Connection `json:"connections"`
}

type Port struct {
	Name     string `json:"name"`
	TypeName string `json:"type"`
}

type Node struct {
	Name        string   `json:"name"`
	Metrics     []string `json:"metrics"`
	Description string   `json:"description,omitempty"`
	TypeName    string   `json:"type"`
	InPorts     []Port   `json:"in_ports"`
	OutPorts    []Port   `json:"out_ports"`
}

type Connection struct {
	ID       string   `json:"id"`
	Metrics  []string `json:"metrics"`
	Capacity int      `json:"capacity"`

	// node_name:port
	Source      []string `json:"src"`
	Destination []string `json:"dst"`
}
