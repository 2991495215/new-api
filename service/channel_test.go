package service

import (
	"errors"
	"net/http"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/stretchr/testify/assert"
)

func TestShouldDisableChannelWith429Threshold(t *testing.T) {
	orig := common.AutomaticDisableChannelEnabled
	common.AutomaticDisableChannelEnabled = true
	defer func() { common.AutomaticDisableChannelEnabled = orig }()

	ResetChannelAutoDisable429Counter(101)
	err := &types.NewAPIError{Err: errors.New("rate limit"), StatusCode: http.StatusTooManyRequests}

	assert.False(t, ShouldDisableChannelWith429Threshold(101, err))
	assert.False(t, ShouldDisableChannelWith429Threshold(101, err))
	assert.True(t, ShouldDisableChannelWith429Threshold(101, err))
	assert.False(t, ShouldDisableChannelWith429Threshold(101, err))
}

func TestResetChannelAutoDisable429Counter(t *testing.T) {
	orig := common.AutomaticDisableChannelEnabled
	common.AutomaticDisableChannelEnabled = true
	defer func() { common.AutomaticDisableChannelEnabled = orig }()

	ResetChannelAutoDisable429Counter(202)
	err := &types.NewAPIError{Err: errors.New("rate limit"), StatusCode: http.StatusTooManyRequests}
	assert.False(t, ShouldDisableChannelWith429Threshold(202, err))
	assert.False(t, ShouldDisableChannelWith429Threshold(202, err))

	ResetChannelAutoDisable429Counter(202)
	assert.False(t, ShouldDisableChannelWith429Threshold(202, err))
	assert.False(t, ShouldDisableChannelWith429Threshold(202, err))
	assert.True(t, ShouldDisableChannelWith429Threshold(202, err))
}

func TestShouldDisableChannelNon429StaysImmediate(t *testing.T) {
	orig := common.AutomaticDisableChannelEnabled
	common.AutomaticDisableChannelEnabled = true
	defer func() { common.AutomaticDisableChannelEnabled = orig }()

	err := &types.NewAPIError{Err: errors.New("unauthorized"), StatusCode: http.StatusUnauthorized}
	assert.True(t, ShouldDisableChannelWith429Threshold(303, err))
}
