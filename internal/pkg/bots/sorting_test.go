package bots

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_channelSort(t *testing.T) {
	botOne := &bot{
		channels: map[string]struct{}{
			"one": {},
		},
	}
	botTwo := &bot{
		channels: map[string]struct{}{
			"one": {},
			"two": {},
		},
	}
	sortedBots := []Bot{botTwo, botOne}
	sort.Sort(channelSort(sortedBots))
	require.Equal(t, []Bot{
		botOne,
		botTwo,
	}, sortedBots)
}
