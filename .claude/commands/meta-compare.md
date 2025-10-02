---
name: meta-compare
description: 比较当前会话与其他项目会话的统计数据
allowed_tools: [Bash]
argument-hint: [project-path]
---

# meta-compare：跨项目会话对比

对比当前会话与指定项目的会话统计数据。

```bash
# 检查 meta-cc 是否安装
if ! command -v meta-cc &> /dev/null; then
    echo "❌ 错误：meta-cc 未安装或不在 PATH 中"
    echo ""
    echo "请安装 meta-cc："
    echo "  1. 下载或构建 meta-cc 二进制文件"
    echo "  2. 将其放置在 PATH 中（如 /usr/local/bin/meta-cc）"
    echo "  3. 确保可执行权限：chmod +x /usr/local/bin/meta-cc"
    echo ""
    echo "详情参见：https://github.com/yale/meta-cc"
    exit 1
fi

# 获取当前项目统计
echo "## 当前项目会话统计"
echo ""
meta-cc parse stats --output md
echo ""
echo "---"
echo ""

# 如果提供了对比项目路径
if [ -n "${1}" ]; then
    echo "## 对比项目会话统计：${1}"
    echo ""
    meta-cc --project "${1}" parse stats --output md
    echo ""
    echo "---"
    echo ""
    echo "## 分析建议"
    echo ""
    echo "请对比两个项目的："
    echo "- Turn 数量和会话时长（效率差异）"
    echo "- 工具使用频率（工作方式差异）"
    echo "- 错误率（代码质量差异）"
    echo "- Top Tools 排名（技术栈差异）"
else
    echo "💡 提示：使用 /meta-compare <项目路径> 来对比其他项目"
    echo ""
    echo "例如："
    echo "  /meta-compare /home/yale/work/NarrativeForge"
    echo "  /meta-compare /home/yale/work/claude-tmux"
fi
```

## 说明

此命令用于对比不同项目的会话统计，帮助：

- **识别效率差异**：哪个项目的开发效率更高？
- **发现工作模式**：不同项目中工具使用习惯的差异
- **评估代码质量**：错误率是否与项目复杂度相关？
- **技术栈分析**：不同技术栈下的开发模式

## 参数

- `project-path`（可选）：要对比的项目路径

## 输出内容

### 当前项目统计
- Turn 数量、工具调用次数、错误率、会话时长
- Top 5 工具使用频率

### 对比项目统计（如果提供）
- 相同的统计指标
- 便于直接对比

### 分析建议
- 关键对比维度提示

## 使用场景

1. **学习最佳实践**：对比成功项目和困难项目的差异
2. **技术选型参考**：不同技术栈的开发体验对比
3. **效率评估**：量化不同项目的开发效率
4. **团队协作**：对比团队成员的工作模式

## 示例输出

```markdown
## 当前项目会话统计

- **Total Turns**: 2,563
- **Tool Calls**: 971
- **Error Rate**: 0.0%
- **Session Duration**: 524.3 minutes

### Top Tools
| Tool | Count | Percentage |
|------|-------|------------|
| Bash | 484 | 49.8% |
| Read | 156 | 16.1% |

---

## 对比项目会话统计：/home/yale/work/NarrativeForge

- **Total Turns**: 2,032
- **Tool Calls**: 750
- **Error Rate**: 0.0%
- **Session Duration**: 2,438.1 minutes

### Top Tools
| Tool | Count | Percentage |
|------|-------|------------|
| Bash | 271 | 36.1% |
| Edit | 154 | 20.5% |

---

## 分析建议

对比发现：
- 当前项目会话时长更短但 Turn 数更多（效率更高）
- 当前项目 Bash 使用占比更高（自动化程度更高？）
- 对比项目 Edit 使用占比更高（手动编辑更多？）
```

## 相关命令

- `/meta-stats`：查看当前会话统计
- `/meta-errors`：分析错误模式
- `@meta-coach`：获取个性化优化建议
