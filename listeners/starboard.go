package listeners

import (
	"github.com/bwmarrin/discordgo"
	"github.com/zekroTJA/shinpuru/util"
)

type ListenerStarboard struct {
}

func NewListenerStarboard() *ListenerStarboard {
	return &ListenerStarboard{}
}

func (l *ListenerStarboard) Handler(s *discordgo.Session, e *discordgo.MessageReactionAdd) {
	if e.Emoji.Name != util.StarboardReaction {
		return
	}
	for _, sb := range util.SetupStarboards {
		if e.GuildID == sb.GuildID && e.ChannelID == sb.ChannelID {
			user, err := s.User(e.UserID)
			if err != nil || user == nil || user.Bot {
				return
			}
			msg, err := s.ChannelMessage(e.ChannelID, e.MessageID)
			if err != nil {
				return
			}
			reactionCount := 0
			for _, r := range msg.Reactions {
				if r.Emoji.Name == util.StarboardReaction {
					reactionCount++
				}
			}
			if reactionCount >= sb.Minimum {

			}
		}
	}
}
