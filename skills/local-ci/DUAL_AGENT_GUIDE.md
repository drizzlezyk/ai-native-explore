# Dual-Agent Security Fix System

## 概述

本系统使用 **双 Agent + 证据核查** 模式来自动修复和验证安全问题。

## 工作流程

```
┌─────────────────┐
│ Run gosec scan  │
│ Find N issues   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Fixer Agent    │
│ • Read scans    │
│ • Fix issues    │
│ • Commit fixes  │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Verifier Agent  │
│ • Read evidence │
│ • Verify fixes  │
│ • Run tests     │
│ • Give verdict  │
└────────┬────────┘
         │
         ▼
    PASS/PARTIAL/FAIL
```

## 快速开始

### 1. 运行安全扫描

```bash
bash .claude/skills/local-ci/scripts/run_gosec.sh
```

如果发现安全问题，会生成 `.ci-temp/gosec-report.json`

### 2. 启动双 Agent 修复流程

```bash
bash .claude/skills/local-ci/scripts/fix_security_issues.sh
```

### 3. 分发 Agent 任务

脚本会提示你分发两个 Agent：

#### Phase 1: Fixer Agent

使用 Claude Code 的 Task tool 分发：

```python
Task(
    subagent_type="general-purpose",
    description="Fix security issues",
    prompt="""
    Read and follow: .claude/skills/local-ci/prompts/fixer-agent-prompt.md

    Evidence: .ci-temp/gosec-report.json

    Fix all security issues found in the report.
    """,
    model="sonnet"  # 或 "opus" 用于复杂修复
)
```

#### Phase 2: Verifier Agent

等待 Fixer 完成后，分发 Verifier：

```python
Task(
    subagent_type="general-purpose",
    description="Verify security fixes",
    prompt="""
    Read and follow: .claude/skills/local-ci/prompts/verifier-agent-prompt.md

    Evidence:
    - Baseline: .ci-temp/gosec-baseline.json
    - After-fix: .ci-temp/gosec-report-after-fix.json
    - Commits: {BEFORE_COMMIT}..{AFTER_COMMIT}

    Verify all fixes independently. Do NOT trust Fixer's claims.
    """,
    model="opus"  # 使用最强模型进行验证
)
```

### 4. 查看结果

#### Fixer 的输出

查看修复摘要：
```bash
cat .ci-temp/security-fixes.md
```

#### Verifier 的输出

查看验证报告：
```bash
cat .ci-temp/verification-report.md
```

根据 Verifier 的判决：
- **PASS** ✅: 所有问题已修复，可以合并
- **PARTIAL** ⚠️: 部分修复，需要继续工作
- **FAIL** ❌: 修复无效，需要重新处理

## Agent 职责

### Fixer Agent

**读取**:
- `.ci-temp/gosec-report.json` - 安全扫描结果

**输出**:
- 修复后的代码（通过 git commits）
- `.ci-temp/security-fixes.md` - 修复摘要

**原则**:
- 每个问题一个 commit
- 描述性的 commit 消息
- 修复后验证代码编译

### Verifier Agent

**读取**:
- `.ci-temp/gosec-baseline.json` - 修复前扫描
- `.ci-temp/gosec-report-after-fix.json` - 修复后扫描
- `git diff <before>..<after>` - 实际代码变更
- `.ci-temp/security-fixes.md` - 仅参考，不信任

**输出**:
- `.ci-temp/verification-report.md` - 验证报告
- 判决: PASS / PARTIAL / FAIL

**原则**:
- 完全独立验证
- 仅信任原始证据
- 对比修复前后的扫描结果
- 验证实际代码变更
- 运行编译和测试

## 常见安全问题修复

| 规则 | 问题 | 修复方法 |
|------|------|----------|
| G101 | 硬编码凭证 | 移到环境变量 |
| G104 | 未处理错误 | 添加 `if err != nil` |
| G201 | SQL 注入 | 使用参数化查询 |
| G304 | 文件路径遍历 | 使用 `filepath.Clean()` |
| G401 | 弱加密 (MD5) | 使用 SHA256 |
| G402 | 不安全的 TLS | 设置 `MinVersion: TLS1.2` |
| G404 | 弱随机数 | 使用 `crypto/rand` |

详细修复示例见 Prompt 文件。

## 文件结构

```
.claude/skills/local-ci/
├── prompts/
│   ├── fixer-agent-prompt.md        # Fixer Agent 指令
│   └── verifier-agent-prompt.md     # Verifier Agent 指令
├── scripts/
│   ├── run_gosec.sh                 # 运行安全扫描
│   └── fix_security_issues.sh       # 双 Agent 协调脚本
└── SKILL.md                         # 总文档

.ci-temp/                            # 临时文件 (不提交)
├── gosec-baseline.json              # 基线扫描
├── gosec-report.json                # 当前扫描
├── gosec-report-after-fix.json      # 修复后扫描
├── security-fixes.md                # Fixer 摘要
└── verification-report.md           # Verifier 报告
```

## 优势

### vs. 单 Agent 修复

| 特性 | 单 Agent | 双 Agent |
|------|----------|----------|
| 验证可靠性 | 自我验证 | 独立验证 |
| 证据来源 | 主 Agent 报告 | 原始扫描结果 |
| 错误检测 | 容易遗漏 | 双重检查 |
| 可信度 | 中等 | 高 |

### vs. 人工修复

| 特性 | 人工 | 双 Agent |
|------|------|----------|
| 速度 | 慢 | 快 |
| 一致性 | 因人而异 | 标准化 |
| 审计跟踪 | 手动记录 | 自动生成 |
| 可扩展性 | 有限 | 高 |

## 最佳实践

1. **始终审查 Verifier 的报告** - 即使是 PASS 判决
2. **对于 PARTIAL 判决** - 手动修复剩余问题后重新运行流程
3. **对于 FAIL 判决** - 查看验证报告了解失败原因
4. **保留证据** - 在清理前备份 `.ci-temp/` 目录以备审计
5. **逐步修复** - 对于大量问题，可以分批修复和验证

## 故障排除

### 问题: Fixer Agent 未生成修复

**检查**:
```bash
# 确认有安全问题
cat .ci-temp/gosec-report.json | jq '.Issues | length'

# 查看 Agent 日志
```

### 问题: Verifier 报告 FAIL

**原因**:
- 修复未实际解决问题
- 引入了新问题
- 代码无法编译

**解决**:
1. 阅读 `.ci-temp/verification-report.md`
2. 查看失败的具体原因
3. 手动修复或重新运行 Fixer

### 问题: 脚本找不到 gosec 报告

**解决**:
```bash
# 先运行扫描
bash .claude/skills/local-ci/scripts/run_gosec.sh

# 然后运行修复
bash .claude/skills/local-ci/scripts/fix_security_issues.sh
```

## 贡献

如需添加新的安全规则修复模板，编辑：
- `prompts/fixer-agent-prompt.md` - 添加修复示例
- `prompts/verifier-agent-prompt.md` - 添加验证标准
