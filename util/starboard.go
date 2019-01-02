package util

var SetupStarboards = make([]*Starboard, 0)

type Starboard struct {
	GuildID   string
	ChannelID string
	Enabled   bool
	Minimum   int
}
