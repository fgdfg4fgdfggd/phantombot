package webserver

import (
	"fmt"
	"sort"
	"strings"

	"github.com/zekroTJA/shinpuru/internal/core"

	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/snowflake"
	routing "github.com/qiangxue/fasthttp-routing"
	"github.com/valyala/fasthttp"
	"github.com/zekroTJA/shinpuru/internal/util"
)

func (ws *WebServer) handlerGetMe(ctx *routing.Context) error {
	userID := ctx.Get("uid").(string)

	user, err := ws.session.User(userID)
	if err != nil {
		return jsonError(ctx, err, fasthttp.StatusInternalServerError)
	}

	created, _ := util.GetDiscordSnowflakeCreationTime(user.ID)

	res := &User{
		User:      user,
		AvatarURL: user.AvatarURL(""),
		CreatedAt: created,
	}

	return jsonResponse(ctx, res, fasthttp.StatusOK)
}

func (ws *WebServer) handlerGuildsGet(ctx *routing.Context) error {
	userID := ctx.Get("uid").(string)

	guilds := make([]*GuildReduced, len(ws.session.State.Guilds))
	i := 0
	for _, g := range ws.session.State.Guilds {
		if g.MemberCount < 10000 {
			for _, m := range g.Members {
				if m.User.ID == userID {
					guilds[i] = GuildReducedFromGuild(g)
					i++
					break
				}
			}
		} else {
			if gm, _ := ws.session.GuildMember(g.ID, userID); gm != nil {
				guilds[i] = GuildReducedFromGuild(g)
				i++
			}
		}
	}
	guilds = guilds[:i]

	return jsonResponse(ctx, &ListResponse{
		N:    i,
		Data: guilds,
	}, fasthttp.StatusOK)
}

func (ws *WebServer) handlerGuildsGetGuild(ctx *routing.Context) error {
	userID := ctx.Get("uid").(string)

	guildID := ctx.Param("guildid")

	memb, _ := ws.session.GuildMember(guildID, userID)
	if memb == nil {
		return jsonError(ctx, errNotFound, fasthttp.StatusNotFound)
	}

	guild, err := ws.session.Guild(guildID)
	if err != nil {
		return jsonError(ctx, err, fasthttp.StatusInternalServerError)
	}

	return jsonResponse(ctx, GuildFromGuild(guild, memb), fasthttp.StatusOK)
}

func (ws *WebServer) handlerGuildsGetMember(ctx *routing.Context) error {
	userID := ctx.Get("uid").(string)

	guildID := ctx.Param("guildid")
	memberID := ctx.Param("memberid")

	var memb *discordgo.Member

	if memb, _ = ws.session.GuildMember(guildID, userID); memb == nil {
		return jsonError(ctx, errNotFound, fasthttp.StatusNotFound)
	}

	memb, _ = ws.session.GuildMember(guildID, memberID)
	if memb == nil {
		return jsonError(ctx, errNotFound, fasthttp.StatusNotFound)
	}

	return jsonResponse(ctx, MemberFromMember(memb), fasthttp.StatusOK)
}

func (ws *WebServer) handlerGetPermissions(ctx *routing.Context) error {
	userID := ctx.Get("uid").(string)

	guildID := ctx.Param("guildid")
	memberID := ctx.Param("memberid")

	if memb, _ := ws.session.GuildMember(guildID, userID); memb == nil {
		return jsonError(ctx, errNotFound, fasthttp.StatusNotFound)
	}

	perm, err := ws.cmdhandler.GetPermissions(ws.session, guildID, memberID)
	if err != nil {
		return jsonError(ctx, err, fasthttp.StatusBadRequest)
	}

	return jsonResponse(ctx, &PermissionsResponse{
		Permissions: perm,
	}, fasthttp.StatusOK)
}

func (ws *WebServer) handlerGetReports(ctx *routing.Context) error {
	userID := ctx.Get("uid").(string)

	guildID := ctx.Param("guildid")
	memberID := ctx.Param("memberid")

	sortBy := string(ctx.QueryArgs().Peek("sortBy"))

	if memb, _ := ws.session.GuildMember(guildID, userID); memb == nil {
		return jsonError(ctx, errNotFound, fasthttp.StatusNotFound)
	}

	var reps []*util.Report
	var err error

	if memberID != "" {
		reps, err = ws.db.GetReportsFiltered(guildID, memberID, -1)
	} else {
		reps, err = ws.db.GetReportsGuild(guildID)
	}

	if err != nil {
		return jsonError(ctx, err, fasthttp.StatusInternalServerError)
	}

	resReps := make([]*Report, 0)
	if reps != nil {
		resReps = make([]*Report, len(reps))
		for i, r := range reps {
			resReps[i] = ReportFromReport(r)
		}
	}

	if sortBy == "created" {
		sort.SliceStable(resReps, func(i, j int) bool {
			return resReps[i].Created.After(resReps[j].Created)
		})
	}

	return jsonResponse(ctx, &ListResponse{
		N:    len(resReps),
		Data: resReps,
	}, fasthttp.StatusOK)
}

func (ws *WebServer) handlerGetReport(ctx *routing.Context) error {
	// userID := ctx.Get("uid").(string)

	_id := ctx.Param("id")

	id, err := snowflake.ParseString(_id)
	if err != nil {
		return jsonError(ctx, err, fasthttp.StatusBadRequest)
	}

	rep, err := ws.db.GetReport(id)
	if err == core.ErrDatabaseNotFound {
		return jsonError(ctx, errNotFound, fasthttp.StatusNotFound)
	}
	if err != nil {
		return jsonError(ctx, err, fasthttp.StatusInternalServerError)
	}

	return jsonResponse(ctx, ReportFromReport(rep), fasthttp.StatusOK)
}

func (ws *WebServer) handlerGetPermissionsAllowed(ctx *routing.Context) error {
	// userID := ctx.Get("uid").(string)

	guildID := ctx.Param("guildid")
	memberID := ctx.Param("memberid")

	perms, err := ws.cmdhandler.GetPermissions(ws.session, guildID, memberID)
	if core.IsErrDatabaseNotFound(err) {
		return jsonError(ctx, errNotFound, fasthttp.StatusNotFound)
	}
	if err != nil {
		return jsonError(ctx, err, fasthttp.StatusInternalServerError)
	}

	cmds := ws.cmdhandler.GetCmdInstances()

	allowed := make([]string, len(cmds))
	i := 0
	for _, cmd := range cmds {
		if core.PermissionCheck(cmd.GetDomainName(), perms) {
			allowed[i] = cmd.GetDomainName()
			i++
		}
	}

	return jsonResponse(ctx, &ListResponse{
		N:    i,
		Data: allowed[:i],
	}, fasthttp.StatusOK)
}

func (ws *WebServer) handlerGetGuildSettings(ctx *routing.Context) error {
	gs := new(GuildSettings)

	guildID := ctx.Param("guildid")

	var err error

	if gs.Prefix, err = ws.db.GetGuildPrefix(guildID); err != nil && !core.IsErrDatabaseNotFound(err) {
		return errInternalOrNotFound(ctx, err)
	}

	if gs.Perms, err = ws.db.GetGuildPermissions(guildID); err != nil && !core.IsErrDatabaseNotFound(err) {
		return errInternalOrNotFound(ctx, err)
	}

	if gs.AutoRole, err = ws.db.GetGuildAutoRole(guildID); err != nil && !core.IsErrDatabaseNotFound(err) {
		return errInternalOrNotFound(ctx, err)
	}

	if gs.ModLogChannel, err = ws.db.GetGuildModLog(guildID); err != nil && !core.IsErrDatabaseNotFound(err) {
		return errInternalOrNotFound(ctx, err)
	}

	if gs.VoiceLogChannel, err = ws.db.GetGuildVoiceLog(guildID); err != nil && !core.IsErrDatabaseNotFound(err) {
		return errInternalOrNotFound(ctx, err)
	}

	if gs.JoinMessageText, gs.JoinMessageChannel, err = ws.db.GetGuildJoinMsg(guildID); err != nil && !core.IsErrDatabaseNotFound(err) {
		return errInternalOrNotFound(ctx, err)
	}

	if gs.LeaveMessageText, gs.LeaveMessageChannel, err = ws.db.GetGuildLeaveMsg(guildID); err != nil && !core.IsErrDatabaseNotFound(err) {
		return errInternalOrNotFound(ctx, err)
	}

	return jsonResponse(ctx, gs, fasthttp.StatusOK)
}

func (ws *WebServer) handlerPostGuildSettings(ctx *routing.Context) error {
	userID := ctx.Get("uid").(string)

	guildID := ctx.Param("guildid")

	var err error

	gs := new(GuildSettings)
	if err = parseJSONBody(ctx, gs); err != nil {
		return jsonError(ctx, err, fasthttp.StatusBadRequest)
	}

	if gs.AutoRole != "" {
		if ok, err := ws.cmdhandler.CheckPermissions(ws.session, guildID, userID, "sp.guild.config.autorole"); err != nil {
			return errInternalOrNotFound(ctx, err)
		} else if !ok {
			return jsonError(ctx, errUnauthorized, fasthttp.StatusUnauthorized)
		}

		if gs.AutoRole == "__RESET__" {
			gs.AutoRole = ""
		}

		if err = ws.db.SetGuildAutoRole(guildID, gs.AutoRole); err != nil {
			return errInternalOrNotFound(ctx, err)
		}
	}

	if gs.ModLogChannel != "" {
		if ok, err := ws.cmdhandler.CheckPermissions(ws.session, guildID, userID, "sp.guild.config.modlog"); err != nil {
			return errInternalOrNotFound(ctx, err)
		} else if !ok {
			return jsonError(ctx, errUnauthorized, fasthttp.StatusUnauthorized)
		}

		if gs.ModLogChannel == "__RESET__" {
			gs.ModLogChannel = ""
		}

		if err = ws.db.SetGuildModLog(guildID, gs.ModLogChannel); err != nil {
			return errInternalOrNotFound(ctx, err)
		}
	}

	if gs.Prefix != "" {
		if ok, err := ws.cmdhandler.CheckPermissions(ws.session, guildID, userID, "sp.guild.config.prefix"); err != nil {
			return errInternalOrNotFound(ctx, err)
		} else if !ok {
			return jsonError(ctx, errUnauthorized, fasthttp.StatusUnauthorized)
		}

		if gs.Prefix == "__RESET__" {
			gs.Prefix = ""
		}

		if err = ws.db.SetGuildPrefix(guildID, gs.Prefix); err != nil {
			return errInternalOrNotFound(ctx, err)
		}
	}

	if gs.VoiceLogChannel != "" {
		if ok, err := ws.cmdhandler.CheckPermissions(ws.session, guildID, userID, "sp.guild.config.voicelog"); err != nil {
			return errInternalOrNotFound(ctx, err)
		} else if !ok {
			return jsonError(ctx, errUnauthorized, fasthttp.StatusUnauthorized)
		}

		if gs.VoiceLogChannel == "__RESET__" {
			gs.VoiceLogChannel = ""
		}

		if err = ws.db.SetGuildVoiceLog(guildID, gs.VoiceLogChannel); err != nil {
			return errInternalOrNotFound(ctx, err)
		}
	}

	if gs.JoinMessageChannel != "" && gs.JoinMessageText != "" {
		if ok, err := ws.cmdhandler.CheckPermissions(ws.session, guildID, userID, "sp.guild.config.joinmsg"); err != nil {
			return errInternalOrNotFound(ctx, err)
		} else if !ok {
			return jsonError(ctx, errUnauthorized, fasthttp.StatusUnauthorized)
		}

		if gs.JoinMessageChannel == "__RESET__" && gs.JoinMessageText == "__RESET__" {
			gs.JoinMessageChannel = ""
			gs.JoinMessageText = ""
		}

		if err = ws.db.SetGuildJoinMsg(guildID, gs.JoinMessageText, gs.JoinMessageChannel); err != nil {
			return errInternalOrNotFound(ctx, err)
		}
	}

	if gs.LeaveMessageChannel != "" && gs.LeaveMessageText != "" {
		if ok, err := ws.cmdhandler.CheckPermissions(ws.session, guildID, userID, "sp.guild.config.leavemsg"); err != nil {
			return errInternalOrNotFound(ctx, err)
		} else if !ok {
			return jsonError(ctx, errUnauthorized, fasthttp.StatusUnauthorized)
		}

		if gs.LeaveMessageChannel == "__RESET__" && gs.LeaveMessageText == "__RESET__" {
			gs.LeaveMessageChannel = ""
			gs.LeaveMessageText = ""
		}

		if err = ws.db.SetGuildLeaveMsg(guildID, gs.LeaveMessageText, gs.LeaveMessageChannel); err != nil {
			return errInternalOrNotFound(ctx, err)
		}
	}

	return jsonResponse(ctx, nil, fasthttp.StatusOK)
}

func (ws *WebServer) handlerGetGuildPermissions(ctx *routing.Context) error {
	userID := ctx.Get("uid").(string)

	guildID := ctx.Param("guildid")

	if memb, _ := ws.session.GuildMember(guildID, userID); memb == nil {
		return jsonError(ctx, errNotFound, fasthttp.StatusNotFound)
	}

	var perms map[string]core.PermissionArray
	var err error

	if perms, err = ws.db.GetGuildPermissions(guildID); err != nil && !core.IsErrDatabaseNotFound(err) {
		return errInternalOrNotFound(ctx, err)
	}

	return jsonResponse(ctx, perms, fasthttp.StatusOK)
}

func (ws *WebServer) handlerPostGuildPermissions(ctx *routing.Context) error {
	userID := ctx.Get("uid").(string)

	guildID := ctx.Param("guildid")

	update := new(PermissionsUpdate)
	if err := parseJSONBody(ctx, update); err != nil {
		return jsonError(ctx, errInvalidArguments, fasthttp.StatusBadRequest)
	}

	sperm := update.Perm[1:]
	if !strings.HasPrefix(sperm, "sp.guild") && !strings.HasPrefix(sperm, "sp.etc") && !strings.HasPrefix(sperm, "sp.chat") {
		return jsonError(ctx, fmt.Errorf("you can only give permissions over the domains 'sp.guild', 'sp.etc' and 'sp.chat'"), fasthttp.StatusBadRequest)
	}

	if ok, err := ws.cmdhandler.CheckPermissions(ws.session, guildID, userID, "sp.guild.config.perms"); err != nil {
		return jsonError(ctx, err, fasthttp.StatusInternalServerError)
	} else if !ok {
		return jsonError(ctx, errUnauthorized, fasthttp.StatusUnauthorized)
	}

	perms, err := ws.db.GetGuildPermissions(guildID)
	if err != nil {
		if core.IsErrDatabaseNotFound(err) {
			return jsonError(ctx, errNotFound, fasthttp.StatusNotFound)
		}
		return jsonError(ctx, err, fasthttp.StatusInternalServerError)
	}

	for _, roleID := range update.RoleIDs {
		rperms, ok := perms[roleID]
		if !ok {
			rperms = make(core.PermissionArray, 0)
		}

		rperms = rperms.Update(update.Perm)

		if err = ws.db.SetGuildRolePermission(guildID, roleID, rperms); err != nil {
			return jsonError(ctx, err, fasthttp.StatusInternalServerError)
		}
	}

	return jsonResponse(ctx, nil, fasthttp.StatusOK)
}

func (ws *WebServer) handlerPostGuildMemberReport(ctx *routing.Context) error {
	userID := ctx.Get("uid").(string)

	guildID := ctx.Param("guildid")

	memberID := ctx.Param("memberid")

	repReq := new(ReportRequest)
	if err := parseJSONBody(ctx, repReq); err != nil {
		return jsonError(ctx, err, fasthttp.StatusBadRequest)
	}

	if ok, err := ws.cmdhandler.CheckPermissions(ws.session, guildID, userID, "sp.guild.mod.report"); err != nil {
		return jsonError(ctx, err, fasthttp.StatusInternalServerError)
	} else if !ok {
		return jsonError(ctx, errUnauthorized, fasthttp.StatusUnauthorized)
	}

	if memberID == userID {
		return jsonError(ctx, fmt.Errorf("you can not report yourself"), fasthttp.StatusBadRequest)
	}

	if repReq.Type >= len(util.ReportTypes) || repReq.Type < util.ReportTypesReserved || len(repReq.Reason) < 3 {
		return jsonError(ctx, errInvalidArguments, fasthttp.StatusBadRequest)
	}

	if repReq.Attachment != "" && util.ImgUrlSRx.MatchString(repReq.Attachment) {
		return jsonError(ctx,
			fmt.Errorf("attachment must be a valid url to a file with type of png, jpg, jpeg, gif, ico, tiff, img, bmp or mp4."),
			fasthttp.StatusBadRequest)
	}

	repID := util.NodesReport[repReq.Type].Generate()

	rep := &util.Report{
		AttachmehtURL: repReq.Attachment,
		ExecutorID:    userID,
		GuildID:       guildID,
		Msg:           repReq.Reason,
		Type:          repReq.Type,
		VictimID:      memberID,
		ID:            repID,
	}

	if err := ws.db.AddReport(rep); err != nil {
		return jsonError(ctx, err, fasthttp.StatusInternalServerError)
	}

	return jsonResponse(ctx, ReportFromReport(rep), fasthttp.StatusCreated)
}

func (ws *WebServer) handlerPostGuildMemberKick(ctx *routing.Context) error {
	userID := ctx.Get("uid").(string)

	guildID := ctx.Param("guildid")

	memberID := ctx.Param("memberid")

	req := new(ReasonRequest)
	if err := parseJSONBody(ctx, req); err != nil {
		return jsonError(ctx, err, fasthttp.StatusBadRequest)
	}

	if ok, err := ws.cmdhandler.CheckPermissions(ws.session, guildID, userID, "sp.guild.mod.kick"); err != nil {
		return jsonError(ctx, err, fasthttp.StatusInternalServerError)
	} else if !ok {
		return jsonError(ctx, errUnauthorized, fasthttp.StatusUnauthorized)
	}

	if memberID == userID {
		return jsonError(ctx, fmt.Errorf("you can not kick yourself"), fasthttp.StatusBadRequest)
	}

	guild, err := ws.session.Guild(guildID)
	if err != nil {
		return jsonError(ctx, err, fasthttp.StatusInternalServerError)
	}

	executor, err := ws.session.GuildMember(guildID, userID)
	if err != nil {
		return jsonError(ctx, err, fasthttp.StatusInternalServerError)
	}

	victim, err := ws.session.GuildMember(guildID, memberID)
	if err != nil {
		return jsonError(ctx, err, fasthttp.StatusInternalServerError)
	}

	if util.RolePosDiff(victim, executor, guild) >= 0 {
		return jsonError(ctx, fmt.Errorf("you can not kick members with higher or same permissions than/as yours"), fasthttp.StatusBadRequest)
	}

	if len(req.Reason) < 3 {
		return jsonError(ctx, errInvalidArguments, fasthttp.StatusBadRequest)
	}

	if req.Attachment != "" && util.ImgUrlSRx.MatchString(req.Attachment) {
		return jsonError(ctx,
			fmt.Errorf("attachment must be a valid url to a file with type of png, jpg, jpeg, gif, ico, tiff, img, bmp or mp4."),
			fasthttp.StatusBadRequest)
	}

	repID := util.NodesReport[0].Generate()

	rep := &util.Report{
		AttachmehtURL: req.Attachment,
		ExecutorID:    userID,
		GuildID:       guildID,
		Msg:           req.Reason,
		Type:          0,
		VictimID:      memberID,
		ID:            repID,
	}

	if err = ws.db.AddReport(rep); err != nil {
		return jsonError(ctx, err, fasthttp.StatusInternalServerError)
	}

	if err = ws.session.GuildMemberDelete(guildID, memberID); err != nil {
		return jsonError(ctx, err, fasthttp.StatusInternalServerError)
	}

	return jsonResponse(ctx, ReportFromReport(rep), fasthttp.StatusCreated)
}
