package http

import (
	"fmt"
	"lineNews/agent"
)

// MockTimeline 生成 mock 时间链数据
func MockTimeline(keyword string) agent.TimelineResponse {
	return agent.TimelineResponse{
		Keyword: keyword,
		Events: []agent.Event{
			{
				ID:       "1",
				Title:    fmt.Sprintf("%s 相关新闻一：事件起源", keyword),
				Time:     "2023-01-10",
				Location: "北京",
				People:   []string{"张三", "李四"},
				Summary:  fmt.Sprintf("围绕 %s 的最初报道和背景信息。", keyword),
			},
			{
				ID:       "2",
				Title:    fmt.Sprintf("%s 相关新闻二：事态发展", keyword),
				Time:     "2023-03-05",
				Location: "上海",
				People:   []string{"王五"},
				Summary:  fmt.Sprintf("%s 相关事件在区域内的进一步发酵与反应。", keyword),
			},
			{
				ID:       "3",
				Title:    fmt.Sprintf("%s 相关新闻三：官方回应", keyword),
				Time:     "2023-05-20",
				Location: "广州",
				People:   []string{"官方发言人"},
				Summary:  fmt.Sprintf("有关部门针对 %s 发布官方说明与政策。", keyword),
			},
			{
				ID:       "4",
				Title:    fmt.Sprintf("%s 相关新闻四：后续影响", keyword),
				Time:     "2023-08-01",
				Location: "深圳",
				People:   []string{"媒体", "专家"},
				Summary:  fmt.Sprintf("%s 对社会、产业或公众情绪产生的长期影响分析。", keyword),
			},
		},
	}
}

// MockGraph 生成 mock 知识图谱数据
func MockGraph(keyword string) agent.GraphResponse {
	nodes := []agent.GraphNode{
		{ID: "e1", Name: fmt.Sprintf("%s 核心事件", keyword), Category: "事件"},
		{ID: "e2", Name: fmt.Sprintf("%s 延伸事件", keyword), Category: "事件"},
		{ID: "p1", Name: "张三", Category: "人物"},
		{ID: "p2", Name: "李四", Category: "人物"},
		{ID: "l1", Name: "北京", Category: "地点"},
		{ID: "l2", Name: "上海", Category: "地点"},
		{ID: "t1", Name: fmt.Sprintf("%s 政策", keyword), Category: "主题"},
	}

	links := []agent.GraphLink{
		{Source: "e1", Target: "p1", Relation: "相关人物"},
		{Source: "e1", Target: "l1", Relation: "发生地点"},
		{Source: "e1", Target: "t1", Relation: "涉及主题"},
		{Source: "e2", Target: "p2", Relation: "相关人物"},
		{Source: "e2", Target: "l2", Relation: "发生地点"},
		{Source: "e2", Target: "t1", Relation: "政策影响"},
		{Source: "e1", Target: "e2", Relation: "事件演化"},
	}

	return agent.GraphResponse{
		Keyword: keyword,
		Nodes:   nodes,
		Links:   links,
	}
}
