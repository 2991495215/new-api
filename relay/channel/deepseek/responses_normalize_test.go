package deepseek

import (
	"testing"

	kitutil "github.com/QuantumNous/new-api/relaykit/relayconvert/kitutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeDeepSeekResponsesParallelCallsCopiesReasoning(t *testing.T) {
	raw := []byte(`[
		{"type":"reasoning","content":[{"type":"reasoning_text","text":"first"}]},
		{"type":"function_call","call_id":"call_1","name":"a","arguments":"{}"},
		{"type":"function_call","call_id":"call_2","name":"b","arguments":"{}"},
		{"type":"function_call_output","call_id":"call_1","output":"{}"},
		{"type":"function_call_output","call_id":"call_2","output":"{}"}
	]`)

	got, err := normalizeDeepSeekResponsesInput(raw)
	require.NoError(t, err)

	var items []map[string]any
	require.NoError(t, kitutil.Unmarshal(got, &items))
	require.Len(t, items, 6)
	assert.Equal(t, "reasoning", items[0]["type"])
	assert.Equal(t, "function_call", items[1]["type"])
	assert.Equal(t, "call_1", items[1]["call_id"])
	assert.Equal(t, "function_call_output", items[2]["type"])
	assert.Equal(t, "reasoning", items[3]["type"])
	assert.Equal(t, "function_call", items[4]["type"])
	assert.Equal(t, "call_2", items[4]["call_id"])
	assert.Equal(t, "function_call_output", items[5]["type"])
}

func TestNormalizeDeepSeekResponsesFlushesInterleavedItemsAfterOutput(t *testing.T) {
	raw := []byte(`[
		{"type":"reasoning","content":[{"type":"reasoning_text","text":"think"}]},
		{"type":"function_call","call_id":"call_1","name":"a","arguments":"{}"},
		{"type":"message","role":"developer","content":[{"type":"input_text","text":"context"}]},
		{"type":"function_call","call_id":"call_2","name":"b","arguments":"{}"},
		{"type":"function_call_output","call_id":"call_1","output":"{}"},
		{"type":"function_call_output","call_id":"call_2","output":"{}"}
	]`)

	got, err := normalizeDeepSeekResponsesInput(raw)
	require.NoError(t, err)

	var items []map[string]any
	require.NoError(t, kitutil.Unmarshal(got, &items))
	require.Len(t, items, 7)
	assert.Equal(t, "reasoning", items[0]["type"])
	assert.Equal(t, "function_call", items[1]["type"])
	assert.Equal(t, "function_call_output", items[2]["type"])
	assert.Equal(t, "message", items[3]["type"])
	assert.Equal(t, "developer", items[3]["role"])
	assert.Equal(t, "reasoning", items[4]["type"])
	assert.Equal(t, "function_call", items[5]["type"])
	assert.Equal(t, "call_2", items[5]["call_id"])
	assert.Equal(t, "function_call_output", items[6]["type"])
}

func TestNormalizeDeepSeekResponsesRepairsCustomToolCalls(t *testing.T) {
	raw := []byte(`[
		{"type":"custom_tool_call","call_id":"patch_1","name":"apply_patch","input":"*** Begin Patch"},
		{"type":"message","role":"developer","content":[{"type":"input_text","text":"hook context"}]},
		{"type":"custom_tool_call_output","call_id":"patch_1","output":"done"}
	]`)

	got, err := normalizeDeepSeekResponsesInput(raw)
	require.NoError(t, err)

	var items []map[string]any
	require.NoError(t, kitutil.Unmarshal(got, &items))
	require.Len(t, items, 3)
	assert.Equal(t, "custom_tool_call", items[0]["type"])
	assert.Equal(t, "custom_tool_call_output", items[1]["type"])
	assert.Equal(t, "message", items[2]["type"])
	assert.Equal(t, "developer", items[2]["role"])
}
