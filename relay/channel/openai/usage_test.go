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
