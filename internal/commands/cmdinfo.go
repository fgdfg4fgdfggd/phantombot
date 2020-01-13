package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/zekroTJA/shinpuru/internal/util"
)

type CmdInfo struct {
}

func (c *CmdInfo) GetInvokes() []string {
	return []string{"info", "information", "description", "credits", "version", "invite"}
}

func (c *CmdInfo) GetDescription() string {
	return "display some information about this bot"
}

func (c *CmdInfo) GetHelp() string {
	return "`info`"
}

func (c *CmdInfo) GetGroup() string {
	return GroupGeneral
}

func (c *CmdInfo) GetDomainName() string {
	return "sp.etc.info"
}

func (c *CmdInfo) Exec(args *CommandArgs) error {
	invLink := fmt.Sprintf("https://rcdforum.com",
		args.Session.State.User.ID, util.InvitePermission)
	emb := &discordgo.MessageEmbed{
		Color: util.ColorEmbedDefault,
		Title: "Info",
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: args.Session.State.User.AvatarURL(""),
		},
		Description: "Phanton Bot a excuslive bot to RCD and RCDForums, " +
			"Created and maintained by RCDForum Staff.",
		Fields: []*discordgo.MessageEmbedField{
			&discordgo.MessageEmbedField{
				Name:  "Robert Miller",
				Value: "[https://rcdforum.com/u/robet_miller)",
			},
			&discordgo.MessageEmbedField{
				Name: "Version",
				Value: fmt.Sprintf("This instance is running on version **%s** (commit hash `%s`)",
					util.AppVersion, util.AppCommit),
			},
			&discordgo.MessageEmbedField{
				Name:  "Licence",
				Value: "Closed Source.",
			},
			&discordgo.MessageEmbedField{
				Name: "Invite",
				Value: fmt.Sprintf("[Site Link](%s).\n```\n%s\n```",
					invLink, invLink),
			},
			&discordgo.MessageEmbedField{
				Name:  "Bug Hunters",
				Value: "Much :heart: to all [**bug hunters**](https://github.com/zekroTJA/shinpuru/blob/dev/bughunters.md).",
			},
			&discordgo.MessageEmbedField{
				Name:  "Development state",
				Value: "You can see current tasks [here](https://rcdforum.com/c/news/5).",
			},
			&discordgo.MessageEmbedField{
				Name: "Credits",
				Value: "[Here](https://rcdforum.com/) RCDForum Site.\n" +
					"Welcome to RCD.",
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "© 2019-2020 RCDForum (Dis_chat)",
		},
	}
	_, err := args.Session.ChannelMessageSendEmbed(args.Channel.ID, emb)
	return err
}
