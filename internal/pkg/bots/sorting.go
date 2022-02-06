package bots

type sortByChannelLen []Bot

func (b sortByChannelLen) Len() int {
	return len(b)
}

func (b sortByChannelLen) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b sortByChannelLen) Less(i, j int) bool {
	return len(b[i].Channels()) < len(b[j].Channels())
}
