# AI Code Reviewer 🤖

[![GitHub Action](https://img.shields.io/badge/GitHub_Action-v2.0-blue.svg)](https://github.com/marketplace/actions/ai-code-reviewer)
[![Docker Image](https://img.shields.io/docker/pulls/cqrect/ai-reviewer)](https://hub.docker.com/r/cqrect/ai-reviewer)

基于 AI 的智能代码审查 GitHub Action，自动为 Pull Request 提供专业代码审查建议。

## 功能特性 ✨

- ✅ AI 驱动的代码质量分析
- ✅ 自动生成改进建议
- ✅ 支持自定义审查规则
- ✅ 多语言项目兼容
- ✅ 敏感文件排除功能

## 快速开始

### 前置要求

- OpenAI API Key
- GitHub 仓库的 `write` 权限

### 基础配置

1. 在仓库根目录创建 `.github/workflows/review.yml`

```yaml
name: AI Code Review
on:
  pull_request_target:
    types: [opened, synchronize, reopened]

jobs:
  review:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      pull-requests: write
    steps:
      - name: Checkout Code
        uses: actions/checkout@v4

      - name: AI Review
        uses: cqrect/ai-reviewer@v2
        with:
          pr_number: ${{ github.event.pull_request.number }}
          github_token: ${{ secrets.GITHUB_TOKEN }}
          openai_api_key: ${{ secrets.OPENAI_API_KEY }}
```

**配置选项** ⚙️

| 参数名            | 必填 | 默认值          | 描述              |
| ----------------- | ---- | --------------- | ----------------- |
| `pr_number`       | 是   | -               | Pull Request 编号 |
| `github_token`    | 是   | -               | GitHub 访问令牌   |
| `openai_api_key`  | 是   | -               | OpenAI API 密钥   |
| `openai_base_url` | 否   | OpenAI 官方 API | 自定义 API 端点   |
| `model_name`      | 否   | gpt-3.5-turbo   | 使用的 AI 模型    |

**配置文件（.review.yml）**

```yaml
# 自定义审查规则
prompt: |
  请用中文进行代码审查，重点检查以下方面：
  1. 安全漏洞
  2. 性能优化
  3. 代码规范

# 总结模版
summary: |
  中文总结本次代码修改的主要内容，突出核心改进点

# 排除文件模式
exclude:
  - '**/test/**'
  - 'vendor/**'
  - '**/*.md'
```

## 提示 💡

1. 如果你希望某个文件跳过检查，可以在项目根目录创建 `.review.yml` 文件，并将其加入 `exclude` 列表。或者你也可以在文件开头的**5 行**内写上`Review: PASS`注释，示例如下：

```go
// Review: PASS
package main

import "fmt"

...
```

2. 如果遇到`401`权限问题，请注意在`action`中使用`pull_request_target`而不是`pull_request`。

3. `secrets.GITHUB_TOKEN`使用 `action`的默认值，不要手动设置。
