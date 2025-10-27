# Prompt File Format Specification

**Version**: 1.0
**Date**: 2025-10-27
**Status**: Draft

## Overview

This document specifies the file format for saved prompts in the Prompt Learning System. All saved prompts follow this format to ensure consistency, searchability, and tooling compatibility.

## File Location

```
<project-root>/.meta-cc/prompts/library/{filename}.md
```

## File Naming Convention

**Pattern**: `{category}-{short-description}-{id}.md`

**Rules**:
- `category`: Lowercase, single word (e.g., `release`, `debug`, `refactor`)
- `short-description`: 2-4 words, kebab-case (e.g., `full-ci-monitoring`)
- `id`: Zero-padded 3-digit number (e.g., `001`, `002`, `123`)

**Examples**:
- `release-full-ci-monitoring-001.md`
- `debug-error-analysis-002.md`
- `refactor-extract-logic-003.md`
- `test-coverage-improvement-004.md`
- `docs-api-reference-005.md`

**ID Assignment**: Sequential, auto-generated based on existing files in category

## File Format Structure

### Overall Structure

```markdown
---
{YAML Frontmatter}
---

## Original Prompts
{Original user prompts that led to optimization}

## Optimized Prompt
{The final optimized prompt}

## Notes (Optional)
{Additional context, usage tips, variations}
```

### YAML Frontmatter Schema

**Required Fields**:
```yaml
id: string                    # Unique identifier (matches filename)
title: string                 # Human-readable title (max 80 chars)
category: string              # Category tag (lowercase, single word)
keywords: [string]            # Search keywords (3-10 items)
created: ISO8601              # Creation timestamp
updated: ISO8601              # Last update timestamp
usage_count: integer          # Number of times reused (starts at 0)
effectiveness: float          # Quality score 0.0-1.0 (default 1.0)
status: enum                  # active | archived | deprecated
```

**Optional Fields**:
```yaml
variables: [string]           # Detected variables like {VERSION}, {BRANCH}
tags: [string]                # Additional categorization tags
author: string                # Original author (for team sharing)
project: string               # Specific project name (for multi-project)
related: [string]             # Related prompt IDs
notes: string                 # Brief notes (max 200 chars)
```

### Field Specifications

#### id (Required)
- **Type**: String
- **Format**: `{category}-{description}-{id}` (matches filename without `.md`)
- **Example**: `release-full-ci-monitoring-001`
- **Validation**: Must match filename pattern

#### title (Required)
- **Type**: String
- **Length**: 10-80 characters
- **Format**: Title case, descriptive
- **Example**: `Full Release with CI Monitoring`
- **Purpose**: Human-readable display name

#### category (Required)
- **Type**: String (enum)
- **Allowed Values**:
  - `release` - Version releases, deployments
  - `debug` - Error investigation, troubleshooting
  - `refactor` - Code restructuring, cleanup
  - `test` - Test writing, coverage improvement
  - `docs` - Documentation updates
  - `feature` - New feature implementation
  - `hotfix` - Critical bug fixes
  - `optimization` - Performance improvements
  - `security` - Security-related changes
  - `other` - Miscellaneous
- **Example**: `release`
- **Purpose**: Primary categorization for filtering

#### keywords (Required)
- **Type**: Array of strings
- **Length**: 3-10 items
- **Format**: Lowercase, mix of English and native language
- **Example**: `[发布, release, 新版本, ci, 监控, automation]`
- **Purpose**: Similarity matching in search
- **Guidelines**:
  - Include common synonyms
  - Include English and native language terms
  - Include abbreviations (CI, CD, API, DB)
  - Include domain-specific terms

#### created (Required)
- **Type**: String (ISO8601 timestamp)
- **Format**: `YYYY-MM-DDTHH:MM:SSZ` (UTC)
- **Example**: `2025-10-27T09:15:30Z`
- **Purpose**: Track prompt age
- **Immutable**: Never changes after creation

#### updated (Required)
- **Type**: String (ISO8601 timestamp)
- **Format**: `YYYY-MM-DDTHH:MM:SSZ` (UTC)
- **Example**: `2025-10-27T14:22:45Z`
- **Purpose**: Track last modification or reuse
- **Updates**: Changed on metadata update or reuse

#### usage_count (Required)
- **Type**: Integer
- **Range**: 0 to unlimited
- **Initial**: 0 (increments on first reuse)
- **Example**: `5`
- **Purpose**: Track popularity for ranking
- **Updates**: Incremented each time prompt is reused

#### effectiveness (Required)
- **Type**: Float
- **Range**: 0.0 to 1.0
- **Initial**: 1.0 (default assumption)
- **Example**: `0.85`
- **Purpose**: Quality scoring (future enhancement)
- **Calculation** (future): Based on reuse patterns and user feedback

#### status (Required)
- **Type**: String (enum)
- **Allowed Values**:
  - `active` - Currently in use
  - `archived` - No longer actively used but kept for reference
  - `deprecated` - Obsolete, should not be used
- **Initial**: `active`
- **Example**: `active`
- **Purpose**: Lifecycle management

#### variables (Optional)
- **Type**: Array of strings
- **Format**: Variable names with braces (e.g., `{VERSION}`)
- **Example**: `["{VERSION}", "{BRANCH}", "{DATE}"]`
- **Purpose**: Indicate prompt template variables
- **Auto-detected**: Extracted from optimized prompt content

#### tags (Optional)
- **Type**: Array of strings
- **Length**: 0-5 items
- **Example**: `["automation", "ci-cd", "github-actions"]`
- **Purpose**: Additional categorization beyond primary category

#### author (Optional)
- **Type**: String
- **Format**: Name or username
- **Example**: `"John Doe"` or `"@johndoe"`
- **Purpose**: Attribution for team sharing

#### project (Optional)
- **Type**: String
- **Example**: `"meta-cc"` or `"frontend-app"`
- **Purpose**: Multi-project differentiation in shared libraries

#### related (Optional)
- **Type**: Array of strings (prompt IDs)
- **Example**: `["release-simple-002", "release-hotfix-003"]`
- **Purpose**: Link related prompts

#### notes (Optional)
- **Type**: String
- **Length**: Max 200 characters
- **Example**: `"Use this for full releases with CI monitoring. For simple releases, see release-simple-002."`
- **Purpose**: Brief contextual notes

## Content Sections

### Original Prompts (Required)

**Purpose**: Record the original user prompts that led to this optimization

**Format**:
```markdown
## Original Prompts
- 提交和发布新版本
- 发布新版本
- release new version with monitoring
```

**Guidelines**:
- List all variations that led to this optimization
- Keep original language (don't translate)
- Order by chronological usage if known
- Useful for similarity matching

### Optimized Prompt (Required)

**Purpose**: The final, optimized prompt ready for reuse

**Format**:
```markdown
## Optimized Prompt

使用预发布自动化工作流完成完整的版本发布：

1. 验证所有测试通过：`make test`
2. 更新版本号和 CHANGELOG
3. 创建 git tag 并推送
4. 触发 CI/CD 流水线
5. 监控发布状态直到完成
6. 验证新版本可用性

重要约束：
- 使用 `make commit` 确保质量门
- CI 必须全部通过才能发布
- 保持 CHANGELOG 格式一致性
- 发布后验证关键功能
```

**Guidelines**:
- Clear, actionable steps
- Include constraints and best practices
- Use variable placeholders: `{VERSION}`, `{BRANCH}`
- Maintain formatting (bullets, numbers, code blocks)
- Keep concise but complete

### Notes (Optional)

**Purpose**: Additional context, variations, tips

**Format**:
```markdown
## Notes

**Usage Tips**:
- For hotfix releases, use `release-hotfix-003` instead
- Add `--dry-run` flag for testing
- Monitor Slack #releases channel during deployment

**Variations**:
- Simple release (no CI monitoring): `release-simple-002`
- Emergency hotfix: `release-hotfix-003`

**History**:
- Created during Phase 27 release
- Updated 2025-10-27 to add monitoring step
```

**Guidelines**:
- Optional section, omit if not needed
- Use for context that doesn't fit elsewhere
- Keep concise
- Use subsections for organization

## Validation Rules

### File-Level Validation
1. File extension must be `.md`
2. Filename must match pattern `{category}-{description}-{id}.md`
3. File must exist in `.meta-cc/prompts/library/`

### YAML Validation
1. YAML frontmatter must be valid (parseable by `yq` or similar)
2. All required fields must be present
3. Field types must match specification
4. Enum fields must use allowed values
5. Timestamps must be valid ISO8601 format

### Content Validation
1. Must include "## Original Prompts" section
2. Must include "## Optimized Prompt" section
3. Original prompts section must have at least one item
4. Optimized prompt must be non-empty

### Consistency Validation
1. `id` in YAML must match filename (without `.md`)
2. `category` in YAML must match category in filename
3. `updated` timestamp must be >= `created` timestamp
4. `usage_count` must be >= 0
5. `effectiveness` must be between 0.0 and 1.0

## Example Files

### Example 1: Simple Release Prompt

**Filename**: `release-simple-001.md`

```markdown
---
id: release-simple-001
title: Simple Version Release
category: release
keywords: [release, 发布, version, simple]
created: 2025-10-27T08:00:00Z
updated: 2025-10-27T08:00:00Z
usage_count: 0
effectiveness: 1.0
variables: ["{VERSION}"]
status: active
---

## Original Prompts
- 发布版本
- release version

## Optimized Prompt

执行简单版本发布：

1. 运行 `./scripts/release/release.sh {VERSION}`
2. 验证发布成功
3. 更新相关文档

约束：
- 仅用于小版本发布
- 不包含 CI 监控
```

### Example 2: Debug Error Analysis

**Filename**: `debug-error-analysis-001.md`

```markdown
---
id: debug-error-analysis-001
title: Comprehensive Error Analysis
category: debug
keywords: [debug, error, analysis, troubleshooting, 错误分析, 调试]
created: 2025-10-27T10:30:00Z
updated: 2025-10-27T15:45:00Z
usage_count: 3
effectiveness: 1.0
tags: ["systematic", "root-cause"]
status: active
---

## Original Prompts
- 分析错误
- debug the error
- troubleshoot failure

## Optimized Prompt

系统化分析错误：

1. **收集信息**：
   - 错误消息和堆栈跟踪
   - 复现步骤
   - 环境信息 (OS, 版本, 配置)

2. **定位根因**：
   - 使用二分法缩小范围
   - 检查最近的变更
   - 查阅相关日志

3. **验证修复**：
   - 编写复现测试
   - 确认修复有效
   - 检查无回归

4. **预防措施**：
   - 添加错误处理
   - 改进日志
   - 更新文档

约束：
- 遵循 TDD 原则
- 测试覆盖率 ≥80%
- 文档化解决方案

## Notes

**Usage Tips**:
- For quick fixes, use `debug-quick-fix-002`
- For performance issues, use `debug-performance-003`

**Related**:
- `test-coverage-improvement-004` - For adding tests
- `docs-troubleshooting-005` - For documentation
```

### Example 3: Refactoring Prompt

**Filename**: `refactor-extract-logic-001.md`

```markdown
---
id: refactor-extract-logic-001
title: Extract Logic Refactoring
category: refactor
keywords: [refactor, extract, 重构, 提取, logic, function]
created: 2025-10-26T14:20:00Z
updated: 2025-10-27T09:10:00Z
usage_count: 1
effectiveness: 1.0
variables: ["{FILE}", "{FUNCTION}"]
tags: ["code-quality", "maintainability"]
status: active
author: "Claude Code User"
---

## Original Prompts
- 重构代码提取逻辑
- extract logic from {FILE}
- refactor to improve readability

## Optimized Prompt

提取逻辑到独立函数：

目标文件：`{FILE}`
目标函数：`{FUNCTION}`

步骤：
1. **分析**：
   - 识别可提取的逻辑块
   - 评估依赖关系
   - 确定函数签名

2. **提取**：
   - 创建新函数
   - 保持单一职责
   - 使用清晰的命名

3. **测试**：
   - 保持原有测试通过
   - 为新函数添加单元测试
   - 验证无副作用

4. **文档**：
   - 添加函数注释
   - 更新相关文档
   - 记录重构原因

约束：
- 遵循项目代码规范
- 测试覆盖率不降低
- 保持向后兼容
- 代码行数 ≤200 (如超过需分阶段)

## Notes

**Best Practices**:
- Keep extracted functions pure when possible
- Use descriptive names over comments
- Consider using interfaces for flexibility

**Related Patterns**:
- Extract method
- Extract class
- Replace conditional with polymorphism
```

## Tooling Support

### Parsing YAML Frontmatter

**Using `yq`**:
```bash
# Extract all fields
yq -f extract '.' file.md

# Extract specific field
yq -f extract '.keywords' file.md

# Validate YAML
yq -f extract '.' file.md > /dev/null && echo "Valid" || echo "Invalid"
```

**Using Python**:
```python
import yaml

def parse_frontmatter(filepath):
    with open(filepath, 'r') as f:
        content = f.read()

    # Extract frontmatter between --- delimiters
    parts = content.split('---', 2)
    if len(parts) < 3:
        return None

    frontmatter = yaml.safe_load(parts[1])
    markdown_content = parts[2].strip()

    return {
        'metadata': frontmatter,
        'content': markdown_content
    }
```

### Validation Script

**Basic validation**:
```bash
#!/bin/bash
# validate-prompt.sh

FILE="$1"

# Check file extension
[[ "$FILE" =~ \.md$ ]] || { echo "Error: Not a .md file"; exit 1; }

# Check filename pattern
BASENAME=$(basename "$FILE")
[[ "$BASENAME" =~ ^[a-z]+-[a-z-]+-[0-9]{3}\.md$ ]] || {
    echo "Error: Filename doesn't match pattern"
    exit 1
}

# Validate YAML
yq -f extract '.' "$FILE" > /dev/null || {
    echo "Error: Invalid YAML frontmatter"
    exit 1
}

# Check required fields
REQUIRED=(id title category keywords created updated usage_count effectiveness status)
for field in "${REQUIRED[@]}"; do
    yq -f extract ".$field" "$FILE" > /dev/null || {
        echo "Error: Missing required field: $field"
        exit 1
    }
done

# Check required sections
grep -q "## Original Prompts" "$FILE" || {
    echo "Error: Missing 'Original Prompts' section"
    exit 1
}

grep -q "## Optimized Prompt" "$FILE" || {
    echo "Error: Missing 'Optimized Prompt' section"
    exit 1
}

echo "✅ Validation passed: $FILE"
```

### Generation Template

**Template for new prompts**:
```markdown
---
id: {CATEGORY}-{DESCRIPTION}-{ID}
title: {TITLE}
category: {CATEGORY}
keywords: [{KEYWORDS}]
created: {TIMESTAMP}
updated: {TIMESTAMP}
usage_count: 0
effectiveness: 1.0
variables: []
status: active
---

## Original Prompts
- {ORIGINAL_1}
- {ORIGINAL_2}

## Optimized Prompt

{OPTIMIZED_CONTENT}

## Notes

{OPTIONAL_NOTES}
```

## Migration and Versioning

### Format Version

**Current**: 1.0
**Future versions**: Add `format_version` field to YAML

### Migration Strategy

When format changes:
1. Add `format_version: "1.0"` to existing files
2. Provide migration script for bulk updates
3. Support backward compatibility for 2 major versions
4. Document breaking changes in CHANGELOG

### Backward Compatibility

- New optional fields: Always backward compatible
- New required fields: Requires migration script
- Renamed fields: Deprecated for 2 versions before removal
- Removed fields: Deprecated for 2 versions before removal

## Best Practices

### For Users

1. **Consistent Categories**: Use predefined categories when possible
2. **Rich Keywords**: Include English, native language, and abbreviations
3. **Clear Titles**: Make them descriptive and searchable
4. **Regular Maintenance**: Archive unused prompts, update effectiveness scores
5. **Team Sharing**: Commit useful prompts to git for team collaboration

### For Tooling

1. **Validation**: Always validate before reading/writing
2. **Error Handling**: Gracefully handle malformed files
3. **Atomic Writes**: Use temp files and rename to avoid corruption
4. **Backup**: Keep backups before bulk modifications
5. **Logging**: Log validation failures for debugging

## References

- [Phase 28 Implementation Plan](./PHASE-28-IMPLEMENTATION-PLAN.md)
- [YAML 1.2 Specification](https://yaml.org/spec/1.2/spec.html)
- [ISO 8601 Date Format](https://en.wikipedia.org/wiki/ISO_8601)
- [Markdown Specification](https://spec.commonmark.org/)

## Changelog

- **2025-10-27**: Initial specification (v1.0)
