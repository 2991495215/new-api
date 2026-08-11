package openai

import (
	"testing"

	"github.com/QuantumNous/new-api/constant"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relaykit/dto"
	relaytypes "github.com/QuantumNous/new-api/relaykit/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFillMissingImageTokenDetails(t *testing.T) {
	info := &relaycommon.RelayInfo{
		OriginModelName: "qwen3.7-plus",
		RelayFormat:     relaytypes.RelayFormatOpenAI,
		ChannelMeta: &relaycommon.ChannelMeta{
			ChannelType:       constant.ChannelTypeAdvancedCustom,
			UpstreamModelName: "qwen3.7-plus",
		},
	}
	meta := &relaytypes.TokenCountMeta{
		CombineText:   "ok",
		MessagesCount: 1,
		Files: []*relaytypes.FileMeta{
			{FileType: relaytypes.FileTypeImage},
		},
	}
	usage := &dto.Usage{PromptTokens: 1788}

	fillMissingImageTokenDetailsWithMeta(info, usage, meta)

	require.Greater(t, usage.PromptTokensDetails.ImageTokens, 0)
	assert.Equal(t, usage.PromptTokens, usage.PromptTokensDetails.TextTokens+usage.PromptTokensDetails.ImageTokens)
}

func TestFillMissingImageTokenDetailsSkipsTextOnlyRequest(t *testing.T) {
	info := &relaycommon.RelayInfo{
		OriginModelName: "qwen3.7-plus",
		RelayFormat:     relaytypes.RelayFormatOpenAI,
		ChannelMeta: &relaycommon.ChannelMeta{
			ChannelType:       constant.ChannelTypeAdvancedCustom,
			UpstreamModelName: "qwen3.7-plus",
		},
	}
	meta := &relaytypes.TokenCountMeta{
		CombineText:   "ok",
		MessagesCount: 1,
	}
	usage := &dto.Usage{PromptTokens: 11}

	fillMissingImageTokenDetailsWithMeta(info, usage, meta)

	assert.Zero(t, usage.PromptTokensDetails.ImageTokens)
	assert.Zero(t, usage.PromptTokensDetails.TextTokens)
}

func TestRewriteStreamUsageInjectsImageTokenDetails(t *testing.T) {
	usage := &dto.Usage{
		PromptTokens:     1788,
		CompletionTokens: 10,
		TotalTokens:      1798,
		PromptTokensDetails: dto.InputTokenDetails{
			TextTokens:  11,
			ImageTokens: 1777,
		},
	}

	data := `{"id":"chatcmpl-test","object":"chat.completion.chunk","created":1,"model":"qwen3.7-plus","choices":[],"usage":{"prompt_tokens":1788,"completion_tokens":10,"total_tokens":1798,"prompt_tokens_details":{"cached_tokens":0}}}`
	rewritten := rewriteStreamUsage(data, usage)

	require.Contains(t, rewritten, `"text_tokens":11`)
	require.Contains(t, rewritten, `"image_tokens":1777`)
}

func TestRewriteStreamUsageKeepsUnrelatedDataWhenNoImageDetails(t *testing.T) {
	data := `{"id":"chatcmpl-test","object":"chat.completion.chunk","created":1,"model":"qwen3.7-plus","choices":[],"usage":{"prompt_tokens":11,"completion_tokens":1,"total_tokens":12}}`
	rewritten := rewriteStreamUsage(data, &dto.Usage{PromptTokens: 11, CompletionTokens: 1, TotalTokens: 12})

	assert.Equal(t, data, rewritten)
}
