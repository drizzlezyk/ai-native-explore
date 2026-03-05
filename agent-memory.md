


### Agent文档 Skill架构梳理

```
your-project/
├── .claude/                       # Claude Code 配置目录
│   ├── commands/                   # 自定义 Slash 命令（如 /new-feature）
│   ├── skills/                      # 自定义 Skills（针对特定任务的指令集）
│   │   ├── demand-analysis/         # 需求分析技能
│   │   │   └── SKILL.md             # 指导如何写 10_CONTEXT.md
│   │   ├── interface-validate/      # 接口验证技能
│   │   │   └── SKILL.md             # 指导如何执行测试并分析结果
│   │   ├── CI-validate/             # 本地CI检查技能
│   │   │   └── SKILL.md             # 指导如何执行测试并分析结果
│   │   └── deploy-to-test/          # 环境发布技能
│   │       ├── SKILL.md
│   │       └── scripts/              # 可能用到的部署脚本
│   └── settings.json                 # MCP 服务器配置
│
├── CLAUDE.md/Agent.md                # 项目全局记忆（最重要的文件！）
│
├── docs/                              # 所有文档
│   ├── _foundation/                   # 项目级文档
│   │   └── PROJECT_TRACKER.yaml       # 进度跟踪
│   └── features/                       # 按功能组织
│       └── user-login/                  # 示例功能
│           ├── 10_CONTEXT.md             # 需求
│           ├── 40_DESIGN.md               # 设计（含接口定义）
│           ├── TASK_PLAN.md                # 任务清单
│           └── TEST_PLAN.md                # 测试计划
│
└── src/                               # 代码目录
```

