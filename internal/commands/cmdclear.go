package commands

import (
	"fmt"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/zekroTJA/shinpuru/internal/util"
	"github.com/zekroTJA/shinpuru/internal/util/static"
)

type CmdClear struct {
}

func (c *CmdClear) GetInvokes() []string {
	return []string{"clear", "c", "purge"}
}

func (c *CmdClear) GetDescription() string {
	return "clear messages in a channel"
}

func (c *CmdClear) GetHelp() string {
	return "`clear` - delete last message\n" +
		"`clear <n>` - clear an ammount of messages\n" +
		"`clear <n> <userResolvable>` - clear an ammount of messages by a specific user"
}

func (c *CmdClear) GetGroup() string {
	return GroupModeration
}

func (c *CmdClear) GetDomainName() string {
	return "sp.guild.mod.clear"
}

func (c *CmdClear) GetSubPermissionRules() []SubPermission {
	return nil
}

func (c *CmdClear) Exec(args *CommandArgs) error {
	var msgsStructs []*discordgo.Message
	var err error

	if len(args.Args) == 0 {
		msgsStructs, err = args.Session.ChannelMessages(args.Channel.ID, 1, "", "", "")
	} else {
		var memb *discordgo.Member
		n, err := strconv.Atoi(args.Args[0])
		if err != nil {
			return util.SendEmbedError(args.Session, args.Channel.ID,
				"Sorry, but the member can not be found on this guild. :cry:").
				DeleteAfter(8 * time.Second).Error()
		} else if n < 0 || n > 100 {
			return util.SendEmbedError(args.Session, args.Channel.ID,
				"Number of messages is invald and must be between *(including)* 0 and 100.").
				DeleteAfter(8 * time.Second).Error()
			return err
		}

		if len(args.Args) >= 2 {
			memb, err = util.FetchMember(args.Session, args.Guild.ID, args.Args[1])
			if err != nil {
				return util.SendEmbedError(args.Session, args.Channel.ID,
					"Sorry, but the member can not be found on this guild. :cry:").
					DeleteAfter(8 * time.Second).Error()
			}
		}
		msgsStructsUnsorted, err := args.Session.ChannelMessages(args.Channel.ID, n, "", "", "")
		if err != nil {
			return err
		}

		if memb != nil {
			for _, m := range msgsStructsUnsorted {
				if m.Author.ID == memb.User.ID {
					msgsStructs = append(msgsStructs, m)
				}
			}
		} else {
			msgsStructs = msgsStructsUnsorted
		}
	}

	if err != nil {
		return err
	}

	msgs := make([]string, len(msgsStructs))
	for i, m := range msgsStructs {
		msgs[i] = m.ID
	}

	err = args.Session.ChannelMessagesBulkDelete(args.Channel.ID, msgs)
	if err != nil {
		return err
	}

	multipleMsgs := ""
	if len(msgs) > 1 {
		multipleMsgs = "s"
	}

	return util.SendEmbed(args.Session, args.Channel.ID,
		fmt.Sprintf("Deleted %d message%s.", len(msgs), multipleMsgs), "", static.ColorEmbedUpdated).
		DeleteAfter(6 * time.Second).Error()
}
