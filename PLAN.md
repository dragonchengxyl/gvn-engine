没问题，这里是完整且连续的 `plan_expansion.md`，已经全部放在一个代码框内，你可以直接点击右上角的“复制”按钮一键提取。

把它和 `SKILL.md` 一起发给你的 AI 助手，你们就可以开始构建这个次世代的视觉小说引擎了。

```markdown
# Project Nexus: Next-Gen Go Visual Narrative Framework
**Version:** 2.0 (Architectural Evolution)
**Codename:** GVN-Nexus

## 1. 架构升级愿景 (Paradigm Shift)
GVN-Nexus 从传统的“状态突变型”引擎升级为**“云原生友好、事件溯源、编译驱动”**的下一代视觉小说框架。
1. **Doc-as-Code (MD-VNDL)**: 使用类 Markdown 语法彻底替代 JSON 剧本，降低创作心智负担。
2. **Event Sourcing (事件溯源)**: 所有状态变更抽象为可逆的 Delta-Action，实现毫秒级剧情回退（Time Machine）与极小体积存档。
3. **ECS & Shader Pipeline**: 引入实体组件系统与 GLSL 着色器，实现电影级视效。

---

## 2. 目录结构扩展规范 (Directory Upgrades)
AI 在执行 Phase 5-8 时，必须在原有架构上扩充以下目录结构：

```text
gvn-engine/
├── internal/
│   ├── compiler/         # [新增] Phase 5: MD-VNDL 剧本编译器
│   │   ├── lexer.go      # 词法分析 (Tokenization)
│   │   ├── parser.go     # 语法树构建 (AST Generation)
│   │   └── ast.go        # 语法树节点定义
│   ├── history/          # [新增] Phase 6: 时间机器与事件溯源
│   │   ├── action.go     # Delta Action 接口与实现
│   │   └── stack.go      # 历史记录栈 (Undo/Redo)
│   ├── ecs/              # [新增] Phase 7: 轻量级实体组件系统
│   │   ├── entity.go
│   │   └── components.go

```

---

## 3. 核心数据契约 (Core Interfaces & Contracts)

**指令：AI 必须严格遵守以下 Go 接口定义，严禁擅自修改核心契约。**

### 3.1 The Action Interface (Phase 6 核心)

所有的状态修改（如改变变量、显示立绘、更改背景）必须实现 `Action` 接口。

```go
type Action interface {
    // Apply 执行动作，使状态向前推进
    Apply(state *GameState) error
    // Undo 撤销动作，使状态精确回退到 Apply 之前
    Undo(state *GameState) error
}

```

### 3.2 The AST Node (Phase 5 核心)

解析后的剧本是一棵语法树，所有的节点必须实现 `Node` 接口。

```go
type NodeType int

type Node interface {
    Type() NodeType
    Execute(engine *Engine) // 触发对应的 Action 生成并压入历史栈
}

```

---

## 4. MD-VNDL 语法规范 (Syntax Grammar)

AI 在编写 `lexer.go` 和 `parser.go` 时，必须严格支持以下正则匹配规则：

* **系统指令 (System Directives)**: 以 `@` 开头，支持 `--` 传参。
* `@bg: school_day.png --fade 1.5`
* `@bgm: theme.ogg --loop true`


* **对话指令 (Dialogue)**: `[角色名] (内联参数) "台词内容"`
* `[凯尔] (enter: left, expr: sad) "我不想这样做。"`
* *解析要求*: 必须能正确提取角色名、动作表 (map) 和台词字符串。


* **分支指令 (Choices)**: 以 `>>` 开头，子项以 `-` 开头。
* `>> 面对敌人的攻击：`
* `  - 拔剑反击 -> Label_Fight`
* `  - 转身逃跑 -> Label_Flee`


* **标记指令 (Labels)**: 以 `#` 开头，用于跳转定位。
* `# Label_Fight`



---

## 5. 详细开发路线图 (Execution Phases)

### Phase 5: MD-VNDL 编译器 (The Compiler)

* **原则**: TDD (测试驱动开发)。必须先写测试再写实现。
* **Task 5.1**: 实现 `compiler/lexer.go`。将字符串按行扫描，输出 `Token` 序列 (如 `TokenDialogue`, `TokenSystem`)。
* **Task 5.2**: 实现 `compiler/parser.go`。将 Token 转换为 `Node` 对象数组 (如 `DialogNode`, `ChoiceNode`)。
* **Task 5.3**: 编写 `compiler_test.go`。输入一段完整的 `.nvn` 文本，断言其生成的 AST 结构 100% 正确。

### Phase 6: 时间机器与事件溯源 (The Time Machine)

* **原则**: 绝对的数据确定性 (Deterministic)。
* **Task 6.1**: 实现 `Action` 接口的具体结构，如 `ActionShowText`, `ActionChangeBG`。每一个 Action 都必须包含 `Undo` 逻辑。
* **Task 6.2**: 实现 `history.Stack`。当玩家向上滚动鼠标滚轮时，调用当前动作的 `Undo()`，并将指针回退。
* **Task 6.3**: 改造现有的 `Engine.Update`，使其不再直接修改状态，而是生成 `Action` 并通过 `Stack.Apply()` 执行。

### Phase 7: ECS 与电影级视效 (Cinematic Vfx)

* **原则**: 渲染与逻辑深度解耦。
* **Task 7.1**: 实现轻量级 `ecs` 模块。定义 `Transform` (坐标), `Sprite` (贴图), `Shader` (特效) 组件。
* **Task 7.2**: 集成 Ebiten `DrawRectShader`。编写通用的 `glsl` 着色器加载器。
* **Task 7.3**: 实现非阻塞转场。例如 `fade 1.5` 必须在后台起一个协程或 Tween 动画，不阻塞主线程的同时改变组件的 Alpha 值。

### Phase 8: 自动化构建与 Headless (CI/CD Ready)

* **原则**: 支持无 GUI 运行。
* **Task 8.1**: 在 `main.go` 引入 `-headless` flag。如果为 true，跳过 `ebiten.RunGame`，直接在内存中 `for` 循环遍历执行 AST。
* **Task 8.2**: 编写 `Dockerfile`，用于打包构建环境。
* **Task 8.3**: 提供 GitLab/GitHub Action 脚本，实现在推送到仓库时，自动运行剧情逻辑测试，并交叉编译输出 Windows `.exe` 与 Android `.apk` 产物。

---

## 6. 给 AI 的特定执行指令 (AI Execution Directives)

> **Context**: 你是我的高级系统架构师。我们将严格执行此《GVN-Nexus 扩展计划》。
> **AI 行动守则 (Rules of Engagement)**:
> 1. **严禁破坏性重构**: 我们在现有 Phase 1-4 的基础上扩展，不要删除原有的 Ebitengine 主循环逻辑，而是用新模块去平滑替换旧逻辑。
> 2. **TDD 强制要求**: 在执行 Phase 5 时，**必须**先提供完整的 `parser_test.go`，证明你的正则表达式和 AST 构建逻辑是坚不可摧的。
> 3. **无缝执行**: 当我进入下一 Phase 的指令时，直接输出代码。文件如果较长，请按模块拆分回复，不要等待我的允许。
> 
> 
> **初始任务**: 请准备好，我们将从 **Phase 5 (MD-VNDL 编译器)** 开始。
