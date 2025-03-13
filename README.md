# AI Reviewer 🤖

![GitHub Action](https://img.shields.io/badge/GitHub%20Action-Ready-blue?logo=github)
![OpenAI Powered](https://img.shields.io/badge/OpenAI-Powered-green?logo=openai)

智能代码审查机器人，自动为 Pull Request 提供专业代码审查建议。

## 快速开始 🚀

### 基本使用

1. 在仓库 `.github/workflows/ai-review.yml` 创建新工作流：

   ```yaml
   name: AI Code Review
   on:
     pull_request_target:
       types: [opened, reopened, synchronize]

   jobs:
     review:
       runs-on: ubuntu-latest
       permissions:
         contents: read
         pull-requests: write
       steps:
         - uses: cqrect/ai-reviewer@v1
           with:
             pr_number: ${{ github.event.pull_request.number }}
             github_token: ${{ secrets.GITHUB_TOKEN }}
             openai_api_key: ${{ secrets.OPENAI_API_KEY }}
   ```

2. 在仓库设置中添加`Secrets`:
   - `OPENAI_API_KEY`: 你的 API 密钥
   - `GITHUB_TOKEN`: GitHub 访问令牌

**配置选项 ⚙️**
输入参数
|参数名称|必需|默认值|描述
---|---|---|---
`pr_number`|是|-|要审查的 Pull Request 编号，通过 `github.event.pull_request.number` 获取
`github_token`|是|-|GitHub 访问令牌，用以评论和审查
`openai_api_key`|是|-|AI API_KEY
`openai_base_url`|否|-|自定义 OpenAI 兼容 API 地址
`model_name`|否|gpt-3.5-turbo|使用的 AI 模型名称
`lang`|否|中文|AI 回答所使用的语言

权限要求

```yaml
permissions:
  contents: read #需要读取代码
  pull-requests: write #需要添加审查评论
```

## 常见问题 ❓

**Q: 遇到 401 权限错误怎么办？**

A: 请使用`pull_request_target`而不是`pull_request`。

**Q: 我可以审核自己的 PR 吗？**

A: 不可以审核`GITHUB_TOKEN`提供者的 PR，即**不能自己审核自己**。

**Q: 支持哪些模型？支持本地模型吗？**

A: 只要兼容`OpenAPI`即可。
