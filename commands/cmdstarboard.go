package commands

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/zekroTJA/shinpuru/util"
)

type CmdStarboard struct {
}

func (c *CmdStarboard) GetInvokes() []string {
	return []string{"starboard", "sb"}
}

func (c *CmdStarboard) GetDescription() string {
	return "set up and manage the guilds star board"
}

func (c *CmdStarboard) GetHelp() string {
	return ""
}

func (c *CmdStarboard) GetGroup() string {
	return GroupChat
}

func (c *CmdStarboard) GetPermission() int {
	return 5
}

func (c *CmdStarboard) Exec(args *CommandArgs) error {
	if len(args.Args) > 1 {
		return nil
	}

	switch strings.ToLower(args.Args[0]) {
	case "setup":
		c.setup(args)
	}

	return nil
}

func (c *CmdStarboard) setup(args *CommandArgs) error {
	msgIDs := make([]string, 0)

	starboard := &util.Starboard{
		GuildID: args.Guild.ID,
		Enabled: true,
	}

	sendSetupEmbed := func(step, maxSteps int, description string) (*discordgo.Message, error) {
		emb := &discordgo.MessageEmbed{
			Color:       util.ColorEmbedDefault,
			Title:       fmt.Sprintf("Starboard Setup Step %d/%d", step, maxSteps),
			Description: description,
			Footer: &discordgo.MessageEmbedFooter{
				Text: "You can exit the setup entering 'exit'.",
			},
		}
		msg, err := args.Session.ChannelMessageSendEmbed(args.Channel.ID, emb)
		msgIDs = append(msgIDs, msg.ID)
		return msg, err
	}

	finish := func(mc *util.MessageCollector) {
		fmt.Println(starboard)
		time.Sleep(5 * time.Second)
		args.Session.ChannelMessagesBulkDelete(args.Channel.ID, msgIDs)
		mc.Close("end")
	}

	_, err := sendSetupEmbed(1, 2, "Enter the channel where the starboard should be set up.\n"+
		"Enter `here` if you want to set up the starboard in thsi channel.")
	if err != nil {
		return err
	}

	filter := func(m *discordgo.Message) bool {
		return m.Author.ID == args.User.ID && strings.Trim(m.Content, " \t") != ""
	}
	options := &util.MessageCollectorOptions{
		MaxMessages:        500,
		DeleteMatchesAfter: true,
		Timeout:            5 * time.Minute,
	}
	mc, err := util.NewMessageCollector(args.Session, args.Channel.ID, filter, options)
	if err != nil {
		return err
	}

	currState := 0
	mc.OnMatched(func(msg *discordgo.Message, c *util.MessageCollector) {
		var hErr error

		if msg.Content == "exit" {
			currState = -1
			mAb, hErr := util.SendEmbedError(args.Session, args.Channel.ID, "Aborted.")
			if hErr != nil {
				msgIDs = append(msgIDs, mAb.ID)
			}
		}

		switch currState {
		case 0:
			var channel *discordgo.Channel
			if msg.Content == "here" {
				channel = args.Channel
			} else {
				channel, hErr = util.FetchChannel(args.Session, msg.GuildID, msg.Content, func(c *discordgo.Channel) bool {
					return c.Type == discordgo.ChannelTypeGuildText
				})
				if hErr != nil {
					errMsg, hErr := util.SendEmbedError(args.Session, msg.ChannelID,
						"Could not find anny text channel passing this resolvable. Please enter again.")
					if hErr != nil {
						msgIDs = append(msgIDs, errMsg.ID)
					}
					return
				}
			}
			starboard.ChannelID = channel.ID
			currState++
			sendSetupEmbed(2, 2, "Please enter the minimum ammount of :star: reactions a message needs to appear in the starboard.")

		case 1:
			if msg.Content == "exit" {
				currState = -1
				return
			}
			numb, hErr := strconv.Atoi(msg.Content)
			if hErr != nil {
				errMsg, hErr := util.SendEmbedError(args.Session, msg.ChannelID,
					"Invalid number. Please enter again.")
				if hErr != nil {

					msgIDs = append(msgIDs, errMsg.ID)
				}
				return
			}
			if numb < 1 {
				errMsg, hErr := util.SendEmbedError(args.Session, msg.ChannelID,
					"Number must be larger than 1. Please enter again.")
				if hErr != nil {
					msgIDs = append(msgIDs, errMsg.ID)
				}
				return
			}
			starboard.Minimum = numb
			currState++
			finish(mc)

		default:
			finish(mc)
		}
	})

	mc.OnClosed(func(reason string, c *util.MessageCollector) {
		fmt.Println(reason, len(c.CollectedMessages), len(c.CollectedMatches))
	})

	return nil
}
