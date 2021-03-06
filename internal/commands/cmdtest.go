package commands

import (
	"fmt"

	"github.com/zekroTJA/shinpuru/internal/util"
)

type CmdTest struct {
}

func (c *CmdTest) GetInvokes() []string {
	return []string{"test"}
}

func (c *CmdTest) GetDescription() string {
	return "just for testing purposes"
}

func (c *CmdTest) GetHelp() string {
	return ""
}

func (c *CmdTest) GetGroup() string {
	return GroupEtc
}

func (c *CmdTest) GetDomainName() string {
	return "sp.test"
}

func (c *CmdTest) Exec(args *CommandArgs) error {
	imgURL := args.Message.Attachments[0].URL
	fmt.Println(imgURL)
	image, _ := util.DownloadImageFromURL(imgURL)
	return args.CmdHandler.db.SaveImageData(image)
}
