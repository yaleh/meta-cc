---
name: meta-suggest-next
description: 根据会话上下文和项目状态建议下一步 prompt
allowed_tools: [Bash]
---

# meta-suggest-next：智能 Prompt 建议

基于当前会话上下文和项目状态，建议下一步最佳 prompt。

```bash
#!/bin/bash

# Step 1: 获取最近用户意图
recent_intents=$(meta-cc query user-messages --match "." --limit 5 --with-context 2 --output json)

# Step 2: 获取项目状态
project_state=$(meta-cc query project-state --include-incomplete-tasks --output json)

# Step 3: 获取成功工作流模式
workflows=$(meta-cc query tool-sequences --min-occurrences 3 --successful-only --with-metrics --output json)

# 将数据传递给 Claude 进行语义分析
cat <<EOF
## 上下文数据

### 最近用户意图
\`\`\`json
$recent_intents
\`\`\`

### 项目状态
\`\`\`json
$project_state
\`\`\`

### 成功工作流模式
\`\`\`json
$workflows
\`\`\`

## 任务

基于以上数据，分析并建议下一步最佳 prompt。考虑：

1. **未完成任务优先级**：从 incomplete_stages 中选择最紧急/重要的
2. **用户意图连续性**：根据 recent_intents 推断下一步自然延续
3. **成功模式复用**：建议使用已验证的高成功率工作流

输出格式（Markdown）：

### 建议 Prompt 1（优先级：高）
**Prompt**: [具体的、可执行的 prompt]

**理由**：
- [基于数据的分析，引用具体数据点]

**预期工作流**: [建议使用的工具序列，基于 workflows 数据]

### 建议 Prompt 2（优先级：中）
...

### 建议 Prompt 3（优先级：低/可选）
...
EOF
```
