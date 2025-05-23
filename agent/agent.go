package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"agent/tools"
)

type Agent struct {
	client         *anthropic.Client
	getUserMessage func() (string, bool)
	tools          []tools.ToolDefinition
}

func NewAgent(
	client *anthropic.Client,
	getUserMessage func() (string, bool),
	tools []tools.ToolDefinition,
) *Agent {
	return &Agent{
		client:         client,
		getUserMessage: getUserMessage,
		tools:          tools,
	}
}

// ClaudeResponse represents a single Claude response, which may include text and tool-use blocks.
type ClaudeResponse struct {
	Texts    []string
	ToolUses []ToolUseBlock
}

type ToolUseBlock struct {
	ID   string
	Name string
	Args map[string]interface{}
}

// Expose RunInference for use in Bubble Tea model
func (a *Agent) RunInference(ctx context.Context, userInput string) (ClaudeResponse, error) {
	messageParam := anthropic.NewUserMessage(anthropic.NewTextBlock(userInput))
	conversation := []anthropic.MessageParam{messageParam}
	resp, err := a.runInference(ctx, conversation)
	if err != nil {
		return ClaudeResponse{}, err
	}
	var result ClaudeResponse
	for _, content := range resp.Content {
		if content.Type == "text" {
			result.Texts = append(result.Texts, content.Text)
		} else if content.Type == "tool_use" {
			var args map[string]interface{}
			_ = json.Unmarshal(content.Input, &args)
			result.ToolUses = append(result.ToolUses, ToolUseBlock{
				ID:   content.ID,
				Name: content.Name,
				Args: args,
			})
		}
	}
	return result, nil
}

func (a *Agent) Run(ctx context.Context) error {
	conversation := []anthropic.MessageParam{}

	fmt.Println("Chat with Claude (use 'ctrl-c' to quit)")

	readUserInput := true
	for {
		if readUserInput {
			fmt.Print("\u001b[94mYou\u001b[0m: ")
			userInput, ok := a.getUserMessage()
			if !ok {
				break
			}

			userMessage := anthropic.NewUserMessage(anthropic.NewTextBlock(userInput))
			conversation = append(conversation, userMessage)
		}

		message, err := a.runInference(ctx, conversation)
		if err != nil {
			return err
		}
		conversation = append(conversation, message.ToParam())

		toolResults := []anthropic.ContentBlockParamUnion{}
		for _, content := range message.Content {
			switch content.Type {
			case "text":
				fmt.Printf("\u001b[93mClaude\u001b[0m: %s\n", content.Text)
			case "tool_use":
				result := a.executeTool(content.ID, content.Name, content.Input)
				toolResults = append(toolResults, result)
			}
		}
		if len(toolResults) == 0 {
			readUserInput = true
			continue
		}
		readUserInput = false
		conversation = append(conversation, anthropic.NewUserMessage(toolResults...))
	}

	return nil
}

// ExecuteTool is a public wrapper for tool execution, allowing external packages to call tools and get (string, error).
func (a *Agent) ExecuteTool(name string, input json.RawMessage) (string, error) {
	for _, toolDef := range a.tools {
		if toolDef.Name == name {
			return toolDef.Function(input)
		}
	}
	return "", fmt.Errorf("tool not found: %s", name)
}

func (a *Agent) executeTool(id, name string, input json.RawMessage) anthropic.ContentBlockParamUnion {
	var toolDef tools.ToolDefinition
	var found bool
	for _, tool := range a.tools {
		if tool.Name == name {
			toolDef = tool
			found = true
			break
		}
	}
	if !found {
		return anthropic.NewToolResultBlock(id, "tool not found", true)
	}
	fmt.Printf("\u001b[92mtool\u001b[0m: %s(%s)\n", name, input)
	response, err := toolDef.Function(input)
	if err != nil {
		return anthropic.NewToolResultBlock(id, err.Error(), true)
	}
	return anthropic.NewToolResultBlock(id, response, false)
}

func (a *Agent) runInference(ctx context.Context, conversation []anthropic.MessageParam) (*anthropic.Message, error) {
	anthropicTools := []anthropic.ToolUnionParam{}
	for _, tool := range a.tools {
		anthropicTools = append(anthropicTools, anthropic.ToolUnionParam{
			OfTool: &anthropic.ToolParam{
				Name:        tool.Name,
				Description: anthropic.String(tool.Description),
				InputSchema: tool.InputSchema,
			},
		})
	}
	message, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaude3_7SonnetLatest,
		MaxTokens: int64(1024),
		Messages:  conversation,
		Tools:     anthropicTools,
	})

	return message, err
}
