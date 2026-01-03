# LineNews - 新闻时间线与知识图谱分析系统

## 项目概述

LineNews 是一个基于 Go 语言开发的轻量级 Web 服务，专注于新闻事件的时间线梳理和知识图谱构建。系统整合了多种 AI 模型和搜索服务，为用户提供全面的事件分析和可视化展示。

## 核心功能

### 1. 新闻时间线生成
- 自动按照时间顺序梳理新闻事件
- 提取事件的关键信息：时间、地点、人物、标题、摘要
- 支持关键词搜索和事件追踪

### 2. 知识图谱构建
- 基于时间线数据构建实体关系图谱
- 可视化展示事件、人物、地点、主题之间的关联
- 支持图谱的动态生成和交互展示

## 技术架构

### 后端技术栈
- **语言**: Go 1.24+
- **Web 框架**: Gin
- **HTTP 客户端**: net/http (标准库)
- **JSON 处理**: encoding/json (标准库)
- **并发控制**: sync (标准库)

### 项目结构
```
lineNews/
├── main.go                    # 主入口文件
├── agent/                     # AI Agent 层
│   ├── prompt/               # 提示词模板
│   ├── tool/                 # LLM 调用工具
│   ├── workflow/             # 工作流逻辑
│   ├── agent.go              # Agent 主入口
│   └── types.go              # 类型定义
├── http/                      # HTTP 层
│   ├── router.go             # 路由配置
│   ├── mock.go               # Mock 数据
│   └── controller/           # 控制器层
│       ├── timeline.go       # 时间链和图谱控制器
│       ├── deepsearch.go     # 深度搜索控制器
│       ├── baike.go          # 百科控制器
│       ├── arkchat.go        # Ark Chat 控制器
│       └── health.go         # 健康检查控制器
├── model/                     # 模型层
│   ├── baidudeepsearch.go    # 百度深度搜索封装
│   ├── baidubaike.go         # 百度百科封装
│   ├── deepseek.go           # DeepSeek API
│   └── ark.go             # Ark API
├── static/                    # 静态资源
│   ├── index.html            # 首页
│   └── graph.html            # 图谱可视化页面
└── vendor/                    # 依赖包
```

### 3. 深度搜索能力
- 集成百度深度搜索 API
- 支持智能问答和深度内容挖掘
- 提供自定义搜索参数配置

### 4. 百科知识检索
- 提供百度百科知识检索功能
- 支持按关键词和词条ID搜索
- 丰富事件背景知识补充

### 前端技术栈
- **前端框架**: 原生 HTML + JavaScript
- **样式**: 纯 CSS
- **DOM 操作**: 原生 DOM API
- **图表库**: ECharts (用于知识图谱可视化)
- **时间线展示**: 自定义布局，节点按时间倒序排列

## API 接口

### 核心 API
- `GET /api/timeline?keyword={关键词}` - 获取新闻时间线
- `GET /api/graph?keyword={关键词}` - 获取知识图谱
- `GET /api/health` - 服务健康检查

### 搜索 API
- `GET /api/deepsearch/search?query={查询}` - 百度深度搜索
- `POST /api/deepsearch/custom` - 自定义深度搜索
- `GET /api/baike/search?keyword={关键词}` - 百科搜索
- `GET /api/baike/lemma?lemma_id={ID}` - 按词条ID搜索

### Chat API
- `GET /api/arkchat/chat?message={消息}` - Ark Chat 对话
- `GET /api/arkchat/stream?message={消息}` - Ark 流式对话

### 静态资源
- `/` - 首页
- `/static/*` - 静态文件服务

## 部署与运行

### 环境要求
- Go 1.24 或更高版本
- 网络连接（用于调用外部 API）

### 本地运行
```bash
# 克隆项目
git clone <repository-url>

# 进入项目目录
cd lineNews

# 安装依赖
go mod tidy

# 运行服务
go run main.go
```

### 访问地址
- 服务地址: `http://localhost:8080`
- 前端页面: `http://localhost:8080/`

## 特色功能


### 2. 模块化设计
- Model 层：外部 API 封装，统一接口规范
- Controller 层：业务逻辑处理，参数校验
- Router 层：路由配置，中间件管理

### 3. 错误处理与容错
- 完善的错误日志记录
- 失败时自动降级到 Mock 数据
- 统一的错误响应格式

### 4. CORS 支持
- 支持从 file:// 等来源访问接口
- 允许跨域请求，便于前端集成

## 扩展能力

### 已集成服务
- 百度千帆大模型平台
- 百度深度搜索
- 百度百科 API
- DeepSeek API
- 字节豆包 API

### 可扩展方向
- 新增更多 AI 模型接入
- 扩展数据源集成
- 增强图谱分析能力
- 添加用户认证功能
- 实现流式响应支持

## 开发规范

### 代码结构
- Model 层：负责外部 API 调用封装
- Controller 层：处理业务逻辑和参数校验
- Router 层：定义路由和中间件配置

### API 设计
- 统一响应格式
- 标准错误处理
- 清晰的路由分组

## 贡献指南

1. Fork 项目
2. 创建功能分支
3. 提交代码变更
4. 发起 Pull Request

## 许可证

[MIT License](LICENSE)

## 联系方式

如有问题或建议，请通过项目 Issues 进行反馈。