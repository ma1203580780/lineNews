package agent

// Event 事件数据结构
type Event struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Time     string   `json:"time"`
	Location string   `json:"location"`
	People   []string `json:"people"`
	Summary  string   `json:"summary"`
}

// TimelineResponse 时间链响应
type TimelineResponse struct {
	Keyword string  `json:"keyword"`
	Events  []Event `json:"events"`
}

// GraphNode 图谱节点
type GraphNode struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
}

// GraphLink 图谱连接
type GraphLink struct {
	Source   string `json:"source"`
	Target   string `json:"target"`
	Relation string `json:"relation"`
}

// GraphResponse 图谱响应
type GraphResponse struct {
	Keyword string      `json:"keyword"`
	Nodes   []GraphNode `json:"nodes"`
	Links   []GraphLink `json:"links"`
}
