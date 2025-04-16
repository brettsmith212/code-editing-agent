# Project Name
Enhanced CLI UI for Code Editing with Claude

## Project Description
This project enhances an existing Go CLI application that uses Anthropic Claude 3.5 Sonnet for agentic code editing with tools (`read_file`, `list_files`, `edit_file`). The current text-based UI will be replaced with an interactive, web-like terminal interface using Bubble Tea, Bubbles, and Lip Gloss. The UI will feature a split-screen layout: a code view with diffs, tabs, and a sidebar on the left, and a chat interface on the right. Users can review, approve, or reject Claude’s changes (individually or all at once) via interactive controls. The design will be minimal, elegant, and terminal-friendly, using dark blues, white text, and green/red accents.

## Target Audience
- Developers using the CLI for code editing with Claude’s agentic capabilities.
- Users of all skill levels who prefer a terminal-based workflow.

## Existing Functionality
- **Chat Loop**: CLI-based interaction with Claude, with user input (blue), Claude responses (yellow), and tool logs (green), using `bufio.Scanner` and `fmt.Print`.
- **Tools**:
  - `read_file`: Reads file contents.
  - `list_files`: Lists files in a directory.
  - `edit_file`: Edits or creates files by replacing text.
- **Agent**: Manages conversation history and tool execution via Anthropic API, with a blocking loop in `Agent.Run`.

## Desired Features
### Chat Interface
- [ ] Replace CLI chat with Bubble Tea components:
    - [ ] `textarea` for multi-line user input, styled with Lip Gloss (blue background).
    - [ ] `viewport` for scrollable conversation history (user in blue, Claude in yellow, tools in green).
    - [ ] “Keep All” and “Reject All” buttons for batch approving/rejecting changes, styled with hover effects.
    - [ ] Status bar for real-time feedback (e.g., “Waiting for Claude…”).
- [ ] Integrate agent logic into Bubble Tea’s `Update` function:
    - [ ] Refactor `Agent.Run` to handle input, API calls, and tools as state transitions.
    - [ ] Use `tea.Cmd` for asynchronous Claude API calls and tool executions.
    - [ ] Implement a state machine (e.g., `waitingForInput`, `waitingForClaude`, `showingToolResult`) for clarity.

### Code View
- [ ] Split-screen panel for file content and diffs:
    - [ ] Display file contents with basic syntax highlighting using `github.com/alecthomas/chroma`.
    - [ ] Show diffs for `edit_file` changes (red for removed, green for added) using `github.com/sergi/go-diff`.
    - [ ] Support line-by-line approval/rejection of changes with buttons or shortcuts.
    - [ ] Scrollable tab bar for open files, with indicators for pending changes.
    - [ ] Buttons or shortcuts for approving, rejecting, editing, or reverting changes.

### Sidebar
- [ ] File navigation panel:
    - [ ] List files using direct file system access (`os.ReadDir` or `filepath.Walk`) for speed.
    - [ ] Bubbles’ `list` component for file selection, opening files in tabs.
    - [ ] Support pagination or filtering for large repos.

### Tool Integration
- [ ] Display Claude-driven tool actions and results:
    - [ ] Inline chat messages (e.g., “tool: edit_file(…)”) with Bubbles’ `spinner` for async operations.
    - [ ] Extensible for future tools (e.g., `git`, `curl`) via a plugin-like `ToolDefinition` system.
- [ ] Collapsible terminal panel for future tool outputs:
    - [ ] Show results (e.g., `git status`) in a `viewport`.
    - [ ] Fallback to chat-based output if panel is complex.

### User Experience
- [ ] Intuitive controls:
    - [ ] Keyboard shortcuts (e.g., `Ctrl+S` to send message, `Ctrl+T` to switch tabs, `Ctrl+A` to approve).
    - [ ] Mouse support for tabs, buttons, and sidebar.
- [ ] Real-time UI updates for messages, diffs, and file lists.
- [ ] Error handling:
    - [ ] Inline chat messages or status bar for errors (e.g., “File not found”, “API error”).
    - [ ] Red-colored error messages with icons (e.g., ✗).

### Performance
- [ ] Idiomatic Go practices for reasonable performance:
    - [ ] Asynchronous handling of API calls and file operations using goroutines and `tea.Cmd`.
    - [ ] Cache file lists and diffs to reduce I/O.
    - [ ] Optimize for large files and long chat histories.

### Accessibility
- [ ] Support for all skill levels:
    - [ ] Mouse-driven interactions for beginners (e.g., clicking tabs, buttons).
    - [ ] Keyboard-driven navigation for advanced users.
    - [ ] Optional help panel with keybinding documentation.

## Design Requests
- [ ] Minimal, web-like aesthetic:
    - [ ] Dark blue background, white text, green/red diff accents.
    - [ ] Compact, readable typography with Lip Gloss.
- [ ] Web-inspired UI elements:
    - [ ] Buttons for approve/reject/edit/revert with hover effects (e.g., bold or color change).
    - [ ] Scrollable tabs and sidebar with active/inactive states.
    - [ ] Icons via Nerd Fonts (e.g., file, tool, and error indicators).
    - [ ] Subtle animations (e.g., fade-in for new messages or diffs).

## Technical Notes
- **Refactoring `Agent.Run` for Bubble Tea**:
  - Move agent logic into Bubble Tea’s `Update` function, using a state machine for states like `waitingForInput`, `waitingForClaude`, `showingToolResult`.
  - Mitigate refactoring effort:
    - Refactor in steps: input handling, conversation history, API calls, rendering.
    - Document mappings from old `Agent.Run` logic to new Bubble Tea components.
  - Mitigate async complexity:
    - Wrap Claude API calls and tool executions in goroutines, returning custom `tea.Msg` types (e.g., `claudeResponseMsg`, `toolResultMsg`).
    - Use `tea.Cmd` for async tasks, leveraging Anthropic SDK’s timeout support.
  - Mitigate blocking risks:
    - Run file operations (e.g., `os.ReadFile`, `os.WriteFile`) in goroutines.
    - Show loading states with Bubbles’ `spinner`.
    - Use buffered channels or timeouts for API/tool results.
- **Libraries**:
  - `github.com/sergi/go-diff`: For diff computation and display.
  - `github.com/alecthomas/chroma`: For syntax highlighting.
  - `os`, `filepath`: For sidebar file navigation.
  - `encoding/json`: For tool inputs/outputs (consider `github.com/tidwall/gjson` for lightweight parsing).
  - `github.com/stretchr/testify`: For future unit tests.
- **Charm Tools**:
  - Bubble Tea: Core UI framework.
  - Bubbles: `textarea`, `viewport`, `list`, `spinner` for UI components.
  - Lip Gloss: Styling for colors, borders, and animations.
  - Huh: Optional for confirmation forms (e.g., “Approve all changes?”).
  - Gum: Fallback for simple prompts.
- Ensure cross-platform compatibility (Windows, macOS, Linux).
- Focus on manual testing, with unit tests deferred.

## Other Notes
- Preserve existing functionality in `agent/` and `tools/` packages.
- Design the UI to be extensible for future tools (e.g., `git`, `curl`) via dynamic `ToolDefinition` registration.
- Future ideas:
  - Configuration file (`.agentrc` or `config.yaml`) for UI settings (colors, keybindings).
  - Undo/redo stack for approved/rejected changes, stored in memory or a file.
  - Session persistence (`session.json`) for conversation history and tabs.
  - Help panel with keybinding documentation for accessibility.
- Recommend Nerd Fonts in the README for optimal icon display.
- Test UI performance with large repos and long chat histories.
- Consider `io/fs` for modern file system APIs if needed.