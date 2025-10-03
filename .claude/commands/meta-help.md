---
name: meta-help
description: 显示所有 meta-cc 相关命令和工具的使用帮助
allowed_tools: [Bash]
---

# meta-help：Meta-CC 使用指南

查看所有 meta-cc 相关功能的完整帮助信息。

```bash
# 检查 meta-cc 是否安装
if ! command -v meta-cc &> /dev/null; then
    echo "❌ 错误：meta-cc 未安装或不在 PATH 中"
    echo ""
    echo "## 安装 meta-cc"
    echo ""
    echo "### 方式 1：从源码构建"
    echo '```bash'
    echo "cd /path/to/meta-cc"
    echo "go build -o meta-cc"
    echo "sudo mv meta-cc /usr/local/bin/"
    echo '```'
    echo ""
    echo "### 方式 2：使用预编译二进制"
    echo "下载对应平台的二进制文件并放置到 PATH 中"
    echo ""
    echo "详情参见：https://github.com/yale/meta-cc"
    exit 1
fi

echo "# Meta-CC 使用指南"
echo ""
echo "meta-cc 是一个 Claude Code 元认知分析工具，帮助你分析会话历史、优化工作流程。"
echo ""
echo "---"
echo ""

# 显示 meta-cc 版本和帮助
echo "## CLI 工具"
echo ""
echo '```'
meta-cc --help
echo '```'
echo ""
echo "---"
echo ""

# Slash Commands
echo "## Slash Commands"
echo ""
echo "在 Claude Code 中可以使用以下斜杠命令："
echo ""
echo "### /meta-stats"
echo "显示当前会话的统计信息"
echo "- Turn 数量（用户/助手）"
echo "- 工具使用频率"
echo "- 错误率和会话时长"
echo ""
echo "### /meta-errors [window-size]"
echo "分析当前会话中的错误模式"
echo "- 参数：window-size（可选，默认 20）"
echo "- 检测重复错误（≥3 次）"
echo "- 提供优化建议"
echo ""
echo "### /meta-compare [project-path]"
echo "对比当前会话与其他项目"
echo "- 参数：project-path（可选）"
echo "- 效率和工具使用对比"
echo "- 识别最佳实践"
echo ""
echo "### /meta-timeline [limit]"
echo "生成会话时间线视图"
echo "- 参数：limit（可选，默认 50）"
echo "- 时序工具使用展示"
echo "- 错误分布可视化"
echo ""
echo "### /meta-help"
echo "显示此帮助信息"
echo ""
echo "---"
echo ""

# Subagent
echo "## Subagent"
echo ""
echo "### @meta-coach"
echo "元认知教练，提供对话式分析和优化建议"
echo ""
echo "**功能**："
echo "- 识别重复性低效操作"
echo "- 发现问题解决模式"
echo "- 引导反思和优化"
echo "- 协助创建 Hooks/Commands"
echo ""
echo "**使用方式**："
echo "在 Claude Code 中输入 @meta-coach 并描述你的问题或疑惑。"
echo ""
echo "**示例**："
echo '```'
echo "@meta-coach 我感觉在重复做某件事..."
echo "@meta-coach 为什么我的测试总是失败？"
echo "@meta-coach 如何提高我的开发效率？"
echo '```'
echo ""
echo "---"
echo ""

# MCP Server
echo "## MCP Server"
echo ""
echo "### meta-insight"
echo "通过 Model Context Protocol 提供 meta-cc 功能"
echo ""
echo "**可用工具**："
echo "- get_session_stats：获取会话统计"
echo "- analyze_errors：分析错误模式"
echo "- extract_tools：提取工具使用数据"
echo ""
echo "**配置方式**："
echo "在 .claude/settings.json 中添加："
echo '```json'
echo '{'
echo '  "mcpServers": {'
echo '    "meta-insight": {'
echo '      "command": "node",'
echo '      "args": [".claude/mcp-servers/meta-insight.js"],'
echo '      "transport": "stdio"'
echo '    }'
echo '  }'
echo '}'
echo '```'
echo ""
echo "---"
echo ""

# 常见用法
echo "## 常见用法示例"
echo ""
echo "### 1. 快速会话概览"
echo '```bash'
echo "/meta-stats"
echo '```'
echo ""
echo "### 2. 分析最近 30 Turns 的错误"
echo '```bash'
echo "/meta-errors 30"
echo '```'
echo ""
echo "### 3. 对比两个项目"
echo '```bash'
echo "/meta-compare /home/yale/work/other-project"
echo '```'
echo ""
echo "### 4. 查看最近 20 Turns 的时间线"
echo '```bash'
echo "/meta-timeline 20"
echo '```'
echo ""
echo "### 5. 获取个性化优化建议"
echo '```'
echo "@meta-coach 帮我分析一下我的工作模式"
echo '```'
echo ""
echo "### 6. 手动运行 CLI 命令"
echo '```bash'
echo "# 当前项目统计"
echo "meta-cc parse stats --output md"
echo ""
echo "# 分析特定项目"
echo "meta-cc --project /path/to/project analyze errors --output json"
echo ""
echo "# 分析特定会话"
echo "meta-cc --session <session-id> parse extract --type tools"
echo '```'
echo ""
echo "---"
echo ""

# 故障排查
echo "## 故障排查"
echo ""
echo "### meta-cc 命令未找到"
echo "确保 meta-cc 已安装并在 PATH 中："
echo '```bash'
echo "which meta-cc"
echo "# 应显示：/usr/local/bin/meta-cc 或类似路径"
echo '```'
echo ""
echo "### 会话文件未找到"
echo "meta-cc 使用以下策略定位会话文件："
echo "1. --session 参数（遍历所有项目）"
echo "2. --project 参数（转换为路径哈希）"
echo "3. 自动检测（当前工作目录）"
echo ""
echo "### 权限错误"
echo "确保 meta-cc 有执行权限："
echo '```bash'
echo "chmod +x /usr/local/bin/meta-cc"
echo '```'
echo ""
echo "---"
echo ""

# 相关资源
echo "## 相关资源"
echo ""
echo "- **GitHub**: https://github.com/yale/meta-cc"
echo "- **文档**: README.md + docs/troubleshooting.md"
echo "- **测试**: \`go test ./...\`"
echo ""
echo "---"
echo ""
echo "💡 **快速提示**："
echo "- 大多数命令支持 --output md|json 参数"
echo "- 使用 @meta-coach 获取交互式帮助"
echo "- 查看 docs/troubleshooting.md 了解常见问题"
```

## 使用场景

- 快速查看所有可用功能
- 学习 meta-cc 的使用方式
- 故障排查参考
- 新用户入门指南
