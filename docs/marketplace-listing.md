# meta-cc - Workflow Analysis for Claude Code

![License](https://img.shields.io/github/license/yaleh/meta-cc)
![Version](https://img.shields.io/github/v/release/yaleh/meta-cc)
![Platform](https://img.shields.io/badge/platform-Linux%20%7C%20macOS%20%7C%20Windows-blue)

**Transform your Claude Code session logs into actionable workflow insights.**

## What is meta-cc?

meta-cc is a metacognition tool that analyzes your Claude Code session history to help you understand and optimize your development workflow. By parsing session logs, detecting patterns, and providing AI-powered recommendations, meta-cc turns raw data into productivity insights.

## Key Features

### 📊 Comprehensive Analytics
- **Session Statistics**: Detailed metrics on tool usage, errors, and workflow patterns
- **Error Detection**: Identify repetitive errors and anti-patterns
- **File Access Tracking**: Understand which files are accessed most frequently
- **Time Series Analysis**: Track productivity metrics over time (hourly/daily/weekly)

### 🎯 Workflow Optimization
- **Pattern Recognition**: Detect common tool sequences and workflow bottlenecks
- **Prompt Analysis**: Learn from your most successful prompts
- **Quality Scoring**: Assess response quality and iteration efficiency
- **Habit Insights**: Discover productivity patterns and areas for improvement

### 🤖 AI-Powered Coaching
- **@meta-coach**: Interactive subagent providing personalized workflow recommendations
- **Context-Aware**: Analyzes your specific project history for tailored advice
- **Multi-Turn Conversations**: Deep dive into specific workflow aspects

### 📈 Visual Dashboards
- **ASCII Charts**: Terminal-friendly visualizations of metrics
- **Timeline Views**: Project evolution over time
- **Focus Analysis**: Attention distribution across files and tasks

## Components

### 10 Slash Commands
- `/meta-stats` - Quick session statistics
- `/meta-errors` - Error pattern analysis
- `/meta-timeline` - Project evolution timeline
- `/meta-viz` - Visual analytics dashboard
- `/meta-habits` - Productivity habit insights
- `/meta-quality-scan` - Quality assessment with scorecard
- `/meta-focus-analyzer` - Attention pattern analysis
- `/meta-guide` - Intelligent guidance and recommendations
- `/meta-next` - Generate next-step prompts
- `/meta-prompt` - Refine prompts using historical patterns

### 3 Subagents
- **@meta-coach** - Comprehensive workflow analysis and coaching
- **@meta-query** - Complex query orchestration with Unix pipelines
- **@project-planner** - Project planning assistance

### 14 MCP Query Tools
Programmatic access to session data for autonomous analysis:
- `get_session_stats` - Session-level metrics
- `query_tools` - Tool call filtering
- `query_user_messages` - User input pattern analysis
- `query_assistant_messages` - Response quality assessment
- `query_conversation` - Full turn analysis
- `query_files` - File operation statistics
- `query_context` - Error context extraction
- `query_tool_sequences` - Workflow pattern detection
- `query_file_access` - File access history
- `query_project_state` - Project evolution tracking
- `query_successful_prompts` - High-quality prompt patterns
- `query_tools_advanced` - SQL-like query expressions
- `query_time_series` - Time-based metric analysis
- `cleanup_temp_files` - Temporary file management

## Installation

```
/plugin marketplace add yaleh/meta-cc
/plugin install meta-cc
```

Restart Claude Code to activate all components.

## Screenshots

### Installation Demo
![Installation Demo](screenshots/installation-demo.gif)
*One-command installation via Claude Code marketplace*

### Feature Showcase
![meta-coach Analysis](screenshots/meta-coach-analysis.png)
*Interactive workflow analysis with @meta-coach subagent*

![meta-viz Dashboard](screenshots/meta-viz-dashboard.png)
*Visual analytics dashboard with ASCII charts*

> Note: Screenshots are placeholders pending manual capture. See [docs/screenshots/README.md](screenshots/README.md) for creation instructions.

## Quick Start

After installation, try these commands:

1. **Get session overview**:
   ```
   /meta-stats
   ```

2. **Analyze errors**:
   ```
   /meta-errors
   ```

3. **Interactive coaching**:
   ```
   @meta-coach analyze my workflow
   ```

4. **Visual dashboard**:
   ```
   /meta-viz
   ```

## Use Cases

### For Solo Developers
- Understand your coding patterns and improve efficiency
- Identify repetitive errors and learn from mistakes
- Track productivity trends over time

### For Team Leads
- Analyze team workflow patterns
- Identify common pain points across projects
- Share best practices based on successful patterns

### For Learning
- Review your Claude Code learning journey
- Track skill progression over time
- Optimize your prompting strategies

## Platform Support

- **Linux**: x86_64, ARM64
- **macOS**: Intel, Apple Silicon
- **Windows**: x86_64 (via Git Bash)

## Documentation

- [Complete Documentation](https://github.com/yaleh/meta-cc/blob/develop/docs/)
- [Installation Guide](https://github.com/yaleh/meta-cc/blob/develop/docs/installation.md)
- [Examples & Usage](https://github.com/yaleh/meta-cc/blob/develop/docs/examples-usage.md)
- [Troubleshooting](https://github.com/yaleh/meta-cc/blob/develop/docs/troubleshooting.md)

## Links

- [GitHub Repository](https://github.com/yaleh/meta-cc)
- [Issue Tracker](https://github.com/yaleh/meta-cc/issues)
- [Changelog](https://github.com/yaleh/meta-cc/blob/develop/CHANGELOG.md)

## License

MIT License - see [LICENSE](https://github.com/yaleh/meta-cc/blob/develop/LICENSE) for details.

## Author

Yale Huang ([@yaleh](https://github.com/yaleh))
