# Implementation Plan

Below is a detailed, step-by-step plan to guide the code generation process for enhancing an existing Go CLI application that uses Anthropic Claude 3.5 Sonnet for code editing. The enhancement replaces the current text-based UI with an interactive, web-like terminal interface using Bubble Tea, Bubbles, and Lip Gloss, featuring a split-screen layout with a code view (including diffs and tabs), a chat interface, and a sidebar for file navigation. This plan breaks the development into small, manageable steps that can be executed sequentially by a code generation AI, ensuring each step is atomic and modifies no more than 20 files (typically 1-3 files per step).

## Setup and Basic Structure

- [x] **Step 1: Install dependencies**
  - **Task**: Install required Go libraries for Bubble Tea, Bubbles, Lip Gloss, diff computation, and syntax highlighting.
  - **Files**: None (user action)
  - **Step Dependencies**: None
  - **User Instructions**: Run the following commands in the terminal to install dependencies:
    ```bash
    go get github.com/charmbracelet/bubbletea
    go get github.com/charmbracelet/bubbles
    go get github.com/charmbracelet/lipgloss
    go get github.com/sergi/go-diff
    go get github.com/alecthomas/chroma
    ```

- [x] **Step 2: Set up Bubble Tea in main.go**
  - **Task**: Modify `main.go` to initialize and run a Bubble Tea program with a basic main model, replacing the existing text-based CLI loop.
  - **Files**:
    - `main.go`: Replace existing code with Bubble Tea initialization, e.g., `p := tea.NewProgram(&mainModel{}); if err := p.Start(); err != nil { log.Fatal(err) }`
  - **Step Dependencies**: Step 1
  - **User Instructions**: None

- [x] **Step 3: Create main model structure**
  - **Task**: Create `models/main.go` with the main Bubble Tea model struct, including basic `Init`, `Update`, and `View` methods, and fields for sub-models (chat, code view, sidebar) and agent state.
  - **Files**:
    - `models/main.go`: New file with `type mainModel struct { chat *chatModel, codeview *codeviewModel, sidebar *sidebarModel, agent *agent.Agent, conversation []anthropic.MessageParam, state string }` and basic method implementations
  - **Step Dependencies**: Step 2
  - **User Instructions**: None

## Chat Interface

- [ ] **Step 4: Implement basic chat sub-model**
  - **Task**: Create `models/chat.go` with a chat sub-model using `bubbles/textarea` for user input and `bubbles/viewport` for conversation history, including basic `Update` and `View` methods.
  - **Files**:
    - `models/chat.go`: New file with `type chatModel struct { textarea textarea.Model, viewport viewport.Model }` and method implementations
  - **Step Dependencies**: Step 3
  - **User Instructions**: None

- [ ] **Step 5: Integrate chat sub-model into main model**
  - **Task**: Add the chat sub-model to the main model and render it in the `View` method, displaying a simple "Chat" placeholder initially.
  - **Files**:
    - `models/main.go`: Modify to initialize `chat` field and include `m.chat.View()` in `View`
  - **Step Dependencies**: Step 4
  - **User Instructions**: None

- [ ] **Step 6: Style chat interface with Lip Gloss**
  - **Task**: Apply Lip Gloss styles to the chat input and viewport for a web-like aesthetic (e.g., borders, padding, colors).
  - **Files**:
    - `models/chat.go`: Add `lipgloss.Style` definitions and apply to `textarea` and `viewport` rendering
  - **Step Dependencies**: Step 5
  - **User Instructions**: None

- [ ] **Step 7: Implement conversation history**
  - **Task**: Modify the main model to maintain conversation history and update the chat sub-model to render it in the viewport.
  - **Files**:
    - `models/main.go`: Add logic to append messages to `conversation` field
    - `models/chat.go`: Update `View` to render `viewport` content from conversation history passed by main model
  - **Step Dependencies**: Step 5
  - **User Instructions**: None

- [ ] **Step 8: Integrate agent logic for sending messages**
  - **Task**: Implement logic in the main model to send user messages to Claude using `tea.Cmd` for asynchronous API calls, update conversation history with responses, and manage state transitions (e.g., `waitingForInput`, `waitingForClaude`).
  - **Files**:
    - `models/main.go`: Add `tea.Cmd` for `agent.RunInference`, handle custom `claudeResponseMsg`, and update `conversation`
    - `agent/agent.go`: Ensure `RunInference` is accessible and functional
  - **Step Dependencies**: Step 7
  - **User Instructions**: None

## Layout and Additional UI Components

- [ ] **Step 9: Implement split-screen layout**
  - **Task**: Modify the main model's `View` method to render a split-screen layout with left (code view/sidebar) and right (chat) panels using Lip Gloss for layout styling.
  - **Files**:
    - `models/main.go`: Update `View` to use `lipgloss.JoinHorizontal` and `lipgloss.JoinVertical` for layout
  - **Step Dependencies**: Step 5
  - **User Instructions**: None

- [ ] **Step 10: Implement sidebar sub-model**
  - **Task**: Create `models/sidebar.go` with a sidebar sub-model using `bubbles/list` to display files from the current directory via direct file system access.
  - **Files**:
    - `models/sidebar.go`: New file with `type sidebarModel struct { list list.Model }` and `os.ReadDir` integration
  - **Step Dependencies**: Step 9
  - **User Instructions**: None

- [ ] **Step 11: Integrate sidebar into main model**
  - **Task**: Add the sidebar sub-model to the main model and render it in the left panel of the split-screen layout.
  - **Files**:
    - `models/main.go`: Modify to initialize `sidebar` field and include `m.sidebar.View()` in left panel
  - **Step Dependencies**: Step 10
  - **User Instructions**: None

- [ ] **Step 12: Implement code view sub-model**
  - **Task**: Create `models/codeview.go` with a code view sub-model using `bubbles/viewport` for file contents and a slice for managing tabs of open files.
  - **Files**:
    - `models/codeview.go`: New file with `type codeviewModel struct { viewport viewport.Model, tabs []string, activeTab int }`
  - **Step Dependencies**: Step 9
  - **User Instructions**: None

- [ ] **Step 13: Integrate code view into main model**
  - **Task**: Add the code view sub-model to the main model and render it in the left panel below the sidebar.
  - **Files**:
    - `models/main.go`: Modify to initialize `codeview` field and include `m.codeview.View()` in left panel
  - **Step Dependencies**: Step 12
  - **User Instructions**: None

- [ ] **Step 14: Enable opening files from sidebar**
  - **Task**: Implement logic to open files selected in the sidebar into code view tabs, reading file contents and updating the viewport.
  - **Files**:
    - `models/sidebar.go`: Add keybinding to trigger file opening
    - `models/codeview.go`: Add logic to add tabs and set `viewport` content
    - `models/main.go`: Handle message passing between sidebar and code view
  - **Step Dependencies**: Step 11, Step 13
  - **User Instructions**: None

## Tool Integration and Diffs

- [ ] **Step 15: Implement tool execution logic**
  - **Task**: Modify the main model to handle tool requests from Claude (e.g., `read_file`, `list_files`, `edit_file`), execute them asynchronously using `tea.Cmd`, and update the UI accordingly.
  - **Files**:
    - `models/main.go`: Add `tea.Cmd` for tool execution and handle results in `Update`
    - `agent/agent.go`: Ensure tool execution methods are callable
  - **Step Dependencies**: Step 8
  - **User Instructions**: None

- [ ] **Step 16: Implement diff computation for edit_file**
  - **Task**: Use `go-diff` to compute diffs when `edit_file` is executed, storing them in the main model for display.
  - **Files**:
    - `models/main.go`: Add diff computation logic in tool execution handler
  - **Step Dependencies**: Step 15
  - **User Instructions**: None

- [ ] **Step 17: Display diffs in code view**
  - **Task**: Modify the code view to display diffs with color highlighting using `chroma` for syntax highlighting when changes are proposed.
  - **Files**:
    - `models/codeview.go`: Update `View` to render diffs from main model
  - **Step Dependencies**: Step 16
  - **User Instructions**: None

- [ ] **Step 18: Implement diff approval mechanisms**
  - **Task**: Add UI elements for approving/rejecting individual changes in the code view (e.g., keybindings) and batch options ("Keep All", "Reject All") in the chat interface.
  - **Files**:
    - `models/codeview.go`: Add approval keybindings and logic
    - `models/chat.go`: Add batch approval buttons
    - `models/main.go`: Coordinate approval logic and apply changes
  - **Step Dependencies**: Step 17
  - **User Instructions**: None

## Final Touches

- [ ] **Step 19: Enhance styling and animations**
  - **Task**: Refine Lip Gloss styles across all components and add subtle animations (e.g., fade-in for new messages) for a polished, web-like look.
  - **Files**:
    - `models/chat.go`: Update styles and add animations
    - `models/codeview.go`: Update styles
    - `models/sidebar.go`: Update styles
  - **Step Dependencies**: Step 18
  - **User Instructions**: None

- [ ] **Step 20: Implement error handling**
  - **Task**: Add error display in a status bar or inline in the chat for common errors (e.g., file not found, API errors), ensuring graceful recovery.
  - **Files**:
    - `models/main.go`: Add error field and handling in `Update`
    - `models/chat.go`: Update `View` to display errors
  - **Step Dependencies**: Step 18
  - **User Instructions**: None

- [ ] **Step 21: Optimize performance**
  - **Task**: Implement caching for file lists in the sidebar and diffs in the code view to reduce redundant computations, ensuring async operations don't block the UI.
  - **Files**:
    - `models/sidebar.go`: Add caching for file list
    - `models/codeview.go`: Add caching for diffs
  - **Step Dependencies**: Step 18
  - **User Instructions**: None

- [ ] **Step 22: Manual testing and refinement**
  - **Task**: Manually test the application with various scenarios (e.g., file editing, API failures) and make necessary adjustments based on feedback.
  - **Files**: Various (as needed)
  - **Step Dependencies**: Step 21
  - **User Instructions**: Test the application by:
    - Opening files from the sidebar.
    - Sending messages to Claude and approving/rejecting changes.
    - Simulating errors (e.g., invalid file paths). Provide feedback for refinements.

## Summary of Approach and Key Considerations

This implementation plan transforms the existing text-based Go CLI into an interactive terminal application over 22 steps, organized into five sections: **Setup and Basic Structure**, **Chat Interface**, **Layout and Additional UI Components**, **Tool Integration and Diffs**, and **Final Touches**. The approach starts with establishing the Bubble Tea framework, then builds the chat interface with agent integration, adds the split-screen layout with sidebar and code view, integrates tool execution with diff display and approval, and concludes with styling, error handling, and optimization.

### Key Considerations:
- **Logical Progression**: Steps build incrementally, starting with core structure and progressing to complex features like diffs and approvals, ensuring dependencies (e.g., chat interface before agent integration) are met.
- **Atomic Steps**: Each step modifies 1-3 files, keeping changes manageable for a single code generation iteration.
- **Asynchronous Handling**: `tea.Cmd` is used for Claude API calls and tool executions to maintain UI responsiveness.
- **Reuse of Existing Code**: The plan leverages `agent/agent.go` and `tools/` packages, integrating them into the Bubble Tea framework with minimal refactoring.
- **Testing**: Incremental testing is encouraged, with a final comprehensive manual test to ensure all features work cohesively.
- **Error Management**: Error handling is deferred to the final stages to focus on core functionality first, then address edge cases systematically.

This plan ensures a smooth transition to a fully functional, interactive CLI application that meets the specification while remaining practical for code generation and user interaction.