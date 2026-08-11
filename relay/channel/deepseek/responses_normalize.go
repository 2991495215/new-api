package deepseek

import (
	"encoding/json"
	"fmt"
	"strings"

	kitutil "github.com/QuantumNous/new-api/relaykit/relayconvert/kitutil"
)

const (
	deepseekResponsesInputTypeReasoning            = "reasoning"
	deepseekResponsesInputTypeFunctionCall         = "function_call"
	deepseekResponsesInputTypeFunctionCallOutput   = "function_call_output"
	deepseekResponsesInputTypeCustomToolCall       = "custom_tool_call"
	deepseekResponsesInputTypeCustomToolCallOutput = "custom_tool_call_output"
	deepseekResponsesInputTypeLocalShellCall       = "local_shell_call"
	deepseekResponsesInputTypeLocalShellCallOutput = "local_shell_call_output"
)

var deepseekResponsesCallOutputType = map[string]string{
	deepseekResponsesInputTypeFunctionCall:   deepseekResponsesInputTypeFunctionCallOutput,
	deepseekResponsesInputTypeCustomToolCall: deepseekResponsesInputTypeCustomToolCallOutput,
	deepseekResponsesInputTypeLocalShellCall: deepseekResponsesInputTypeLocalShellCallOutput,
}

// normalizeDeepSeekResponsesInput reorders Responses input so DeepSeek can
// replay Codex turns that use parallel tool calls.
func normalizeDeepSeekResponsesInput(raw json.RawMessage) (json.RawMessage, error) {
	if len(raw) == 0 || kitutil.GetJsonType(raw) != "array" {
		return raw, nil
	}

	var items []map[string]any
	if err := kitutil.Unmarshal(raw, &items); err != nil {
		return nil, fmt.Errorf("invalid deepseek responses input: %w", err)
	}

	out, err := kitutil.Marshal(normalizeDeepSeekResponsesItems(items))
	if err != nil {
		return nil, fmt.Errorf("failed to marshal deepseek responses input: %w", err)
	}
	return out, nil
}

func normalizeDeepSeekResponsesItems(items []map[string]any) []map[string]any {
	result := make([]map[string]any, 0, len(items))
	deferred := make([]map[string]any, 0)
	pendingCalls := make([]map[string]any, 0)
	var turnReasoning map[string]any
	emittedCallsInTurn := 0

	for _, item := range items {
		if item == nil {
			result = append(result, item)
			continue
		}

		switch strings.TrimSpace(kitutil.Interface2String(item["type"])) {
		case deepseekResponsesInputTypeReasoning:
			turnReasoning = item
			emittedCallsInTurn = 0
			if len(pendingCalls) > 0 {
				deferred = append(deferred, item)
			} else {
				result = append(result, item)
			}
		case "message":
			role := strings.TrimSpace(kitutil.Interface2String(item["role"]))
			if role == "assistant" || role == "user" {
				emittedCallsInTurn = 0
			}
			if len(pendingCalls) > 0 {
				deferred = append(deferred, item)
			} else {
				result = append(result, item)
			}
		case deepseekResponsesInputTypeFunctionCall, deepseekResponsesInputTypeCustomToolCall, deepseekResponsesInputTypeLocalShellCall:
			pendingCalls = append(pendingCalls, item)
		case deepseekResponsesInputTypeFunctionCallOutput, deepseekResponsesInputTypeCustomToolCallOutput, deepseekResponsesInputTypeLocalShellCallOutput:
			callID := strings.TrimSpace(kitutil.Interface2String(item["call_id"]))
			if callID == "" {
				callID = strings.TrimSpace(kitutil.Interface2String(item["id"]))
			}
			outputType := strings.TrimSpace(kitutil.Interface2String(item["type"]))
			callIndex := -1
			for i, call := range pendingCalls {
				callType := strings.TrimSpace(kitutil.Interface2String(call["type"]))
				if deepseekResponsesCallOutputType[callType] != outputType {
					continue
				}
				candidate := strings.TrimSpace(kitutil.Interface2String(call["call_id"]))
				if candidate == "" {
					candidate = strings.TrimSpace(kitutil.Interface2String(call["id"]))
				}
				if candidate == callID {
					callIndex = i
					break
				}
			}
			if callIndex >= 0 {
				call := pendingCalls[callIndex]
				pendingCalls = append(pendingCalls[:callIndex], pendingCalls[callIndex+1:]...)
				if turnReasoning != nil && emittedCallsInTurn > 0 {
					result = append(result, cloneResponsesItem(turnReasoning))
				}
				result = append(result, call)
				emittedCallsInTurn++
			}
			result = append(result, item)
			result = append(result, deferred...)
			deferred = deferred[:0]
		default:
			if len(pendingCalls) > 0 {
				deferred = append(deferred, item)
			} else {
				result = append(result, item)
			}
		}
	}

	result = append(result, pendingCalls...)
	result = append(result, deferred...)
	return result
}

func cloneResponsesItem(item map[string]any) map[string]any {
	raw, err := kitutil.Marshal(item)
	if err != nil {
		return item
	}
	var out map[string]any
	if err := kitutil.Unmarshal(raw, &out); err != nil {
		return item
	}
	return out
}
