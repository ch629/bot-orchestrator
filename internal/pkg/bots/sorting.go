package bots

type channelSort []*botState

func (b channelSort) Len() int {
	return len(b)
}

func (b channelSort) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b channelSort) Less(i, j int) bool {
	return len(b[i].channels) < len(b[j].channels)
}
