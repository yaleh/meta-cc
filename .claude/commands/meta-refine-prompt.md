---
name: meta-refine-prompt
description: 将口语化 prompt 改写为清晰、完备的 prompt
allowed_tools: [Bash]
argument-hint: "<user_prompt>"
---

# meta-refine-prompt：Prompt 优化器

将用户的口语化、不严谨的 prompt 改写为更清晰、更完备的 prompt。

```bash
#!/bin/bash

USER_PROMPT="${1:-请输入要优化的 prompt}"

# Step 1: 获取项目上下文
project_state=$(meta-cc query project-state --output json)

# Step 2: 获取成功 prompts 参考
successful_prompts=$(meta-cc query successful-prompts --limit 10 --output json)

# 传递给 Claude 进行改写
cat <<EOF
## 原始 Prompt
> $USER_PROMPT

## 项目上下文
\`\`\`json
$project_state
\`\`\`

## 成功 Prompt 参考
\`\`\`json
$successful_prompts
\`\`\`

## 任务

基于以上数据，将原始 prompt 改写为结构化、完备的 prompt。

### 改写要求

1. **明确目标**：
   - 使用具体动词（实现、添加、修复、重构）
   - 指定对象和范围（具体文件、功能、模块）

2. **提供上下文**：
   - 关联当前项目状态（基于 project_state）
   - 说明为什么要做这个任务

3. **设定边界**：
   - 明确约束（代码行数、时间限制、依赖范围）
   - 列出交付物（文件、测试、文档）

4. **验收标准**：
   - 可测试的完成标志
   - 质量指标（测试覆盖率、性能要求）

### 输出格式

#### 优化后的 Prompt

[完整的、结构化的 prompt，包含上述 4 个要素]

#### 改进说明

- **原问题**：[原 prompt 的具体不足]
- **改进点**：[具体改进措施，引用成功 prompts 的模式]
- **预期效果**：[改写后的优势]
EOF
```
