package bots

import (
	"sort"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func Test_channelSort(t *testing.T) {
	botOne := &botState{
		id: uuid.New(),
		channels: map[string]struct{}{
			"one": {},
		},
	}
	botTwo := &botState{
		id: uuid.New(),
		channels: map[string]struct{}{
			"one": {},
			"two": {},
		},
	}
	originalBots := []*botState{botOne, botTwo}
	sortedBots := []*botState{botOne, botTwo}
	sort.Sort(channelSort(sortedBots))
	require.Equal(t, []*botState{
		originalBots[0],
		originalBots[1],
	}, []*botState(sortedBots))
}
