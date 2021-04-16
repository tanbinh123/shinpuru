package controllers

import (
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/snowflake"
	"github.com/gofiber/fiber/v2"
	"github.com/makeworld-the-better-one/go-isemoji"
	"github.com/sarulabs/di/v2"
	"github.com/zekroTJA/shinpuru/internal/core/config"
	"github.com/zekroTJA/shinpuru/internal/core/database"
	"github.com/zekroTJA/shinpuru/internal/core/middleware"
	"github.com/zekroTJA/shinpuru/internal/core/webserver/v1/models"
	"github.com/zekroTJA/shinpuru/internal/core/webserver/wsutil"
	sharedmodels "github.com/zekroTJA/shinpuru/internal/shared/models"
	"github.com/zekroTJA/shinpuru/internal/util/report"
	"github.com/zekroTJA/shinpuru/internal/util/static"
	"github.com/zekroTJA/shinpuru/pkg/discordutil"
	"github.com/zekroTJA/shinpuru/pkg/fetch"
	"github.com/zekroTJA/shinpuru/pkg/permissions"
)

type GuildsController struct {
	session *discordgo.Session
	cfg     *config.Config
	db      database.Database
	pmw     *middleware.PermissionsMiddleware
}

func (c *GuildsController) Setup(container di.Container, router fiber.Router) {
	c.session = container.Get(static.DiDiscordSession).(*discordgo.Session)
	c.cfg = container.Get(static.DiConfig).(*config.Config)
	c.db = container.Get(static.DiDatabase).(database.Database)
	c.pmw = container.Get(static.DiPermissionMiddleware).(*middleware.PermissionsMiddleware)

	router.Get("", c.getGuilds)
	router.Get("/:guildid", c.getGuild)
	router.Get("/:guildid/scoreboard", c.getGuildScoreboard)
	router.Get("/:guildid/starboard", c.getGuildStarboard)
	router.Delete("/:guildid/antiraid/joinlog", c.pmw.HandleWs(c.session, "sp.guild.config.antiraid"), c.deleteGuildAntiraidJoinlog)
	router.Get("/:guildid/reports", c.getReports)
	router.Get("/:guildid/reports/count", c.getReportsCount)
	router.Get("/:guildid/permissions", c.getGuildPermissions)
	router.Post("/:guildid/permissions", c.pmw.HandleWs(c.session, "sp.guild.config.perms"), c.postGuildPermissions)
	router.Post("/:guildid/inviteblock", c.pmw.HandleWs(c.session, "sp.guild.mod.inviteblock"), c.postGuildToggleInviteblock)
	router.Get("/:guildid/unbanrequests", c.pmw.HandleWs(c.session, "sp.guild.mod.unbanrequests"), c.getGuildUnbanrequests)
	router.Get("/:guildid/unbanrequests/count", c.pmw.HandleWs(c.session, "sp.guild.mod.unbanrequests"), c.getGuildUnbanrequestsCount)
	router.Get("/:guildid/unbanrequests/:id", c.pmw.HandleWs(c.session, "sp.guild.mod.unbanrequests"), c.getGuildUnbanrequest)
	router.Post("/:guildid/unbanrequests/:id", c.pmw.HandleWs(c.session, "sp.guild.mod.unbanrequests"), c.postGuildUnbanrequest)
	router.Get("/:guildid/settings", c.getGuildSettings)
	router.Post("/:guildid/settings", c.postGuildSettings)
	router.Get("/:guildid/settings/karma", c.pmw.HandleWs(c.session, "sp.guild.config.karma"), c.getGuildSettingsKarma)
	router.Post("/:guildid/settings/karma", c.pmw.HandleWs(c.session, "sp.guild.config.karma"), c.postGuildSettingsKarma)
	router.Get("/:guildid/settings/karma/blocklist", c.pmw.HandleWs(c.session, "sp.guild.config.karma"), c.getGuildSettingsKarmaBlocklist)
	router.Put("/:guildid/settings/karma/blocklist/:memberid", c.pmw.HandleWs(c.session, "sp.guild.config.karma"), c.putGuildSettingsKarmaBlocklist)
	router.Delete("/:guildid/settings/karma/blocklist/:memberid", c.pmw.HandleWs(c.session, "sp.guild.config.karma"), c.deleteGuildSettingsKarmaBlocklist)
	router.Get("/:guildid/settings/antiraid", c.pmw.HandleWs(c.session, "sp.guild.config.antiraid"), c.getGuildSettingsAntiraid)
	router.Post("/:guildid/settings/antiraid", c.pmw.HandleWs(c.session, "sp.guild.config.antiraid"), c.postGuildSettingsAntiraid)
}

func (c *GuildsController) getGuilds(ctx *fiber.Ctx) (err error) {
	uid := ctx.Locals("uid").(string)

	guilds := make([]*models.GuildReduced, len(c.session.State.Guilds))
	i := 0
	for _, g := range c.session.State.Guilds {
		if g.MemberCount < 10000 {
			for _, m := range g.Members {
				if m.User.ID == uid {
					guilds[i] = models.GuildReducedFromGuild(g)
					i++
					break
				}
			}
		} else {
			if gm, _ := c.session.GuildMember(g.ID, uid); gm != nil {
				guilds[i] = models.GuildReducedFromGuild(g)
				i++
			}
		}
	}
	guilds = guilds[:i]

	return ctx.JSON(&models.ListResponse{N: len(guilds), Data: guilds})
}

func (c *GuildsController) getGuild(ctx *fiber.Ctx) error {
	uid := ctx.Locals("uid").(string)

	guildID := ctx.Params("guildid")

	memb, _ := c.session.GuildMember(guildID, uid)
	if memb == nil {
		return fiber.ErrNotFound
	}

	guild, err := discordutil.GetGuild(c.session, guildID)
	if err != nil {
		return err
	}

	gRes := models.GuildFromGuild(guild, memb, c.db, c.cfg.Discord.OwnerID)

	return ctx.JSON(gRes)
}

func (c *GuildsController) getGuildScoreboard(ctx *fiber.Ctx) error {
	guildID := ctx.Params("guildid")
	limit, err := wsutil.GetQueryInt(ctx, "limit", 25, 1, 100)
	if err != nil {
		return err
	}

	karmaList, err := c.db.GetKarmaGuild(guildID, limit)

	if err == database.ErrDatabaseNotFound {
		return fiber.ErrNotFound
	} else if err != nil {
		return err
	}

	results := make([]*models.GuildKarmaEntry, len(karmaList))

	var i int
	for _, e := range karmaList {
		member, err := discordutil.GetMember(c.session, guildID, e.UserID)
		if err != nil {
			continue
		}
		results[i] = &models.GuildKarmaEntry{
			Member: models.MemberFromMember(member),
			Value:  e.Value,
		}
		i++
	}

	return ctx.JSON(&models.ListResponse{N: i, Data: results[:i]})
}

func (c *GuildsController) deleteGuildAntiraidJoinlog(ctx *fiber.Ctx) error {
	guildID := ctx.Params("guildid")

	if err := c.db.FlushAntiraidJoinList(guildID); err != nil && !database.IsErrDatabaseNotFound(err) {
		return err
	}

	return ctx.JSON(struct{}{})
}

func (c *GuildsController) getGuildStarboard(ctx *fiber.Ctx) error {
	guildID := ctx.Params("guildid")
	limit, err := wsutil.GetQueryInt(ctx, "limit", 20, 1, 100)
	if err != nil {
		return err
	}
	offset, err := wsutil.GetQueryInt(ctx, "offset", 0, 0, 0)
	if err != nil {
		return err
	}
	sortQ := ctx.Query("sort")

	var sort sharedmodels.StarboardSortBy
	switch string(sortQ) {
	case "latest":
		sort = sharedmodels.StarboardSortByLatest
	case "top":
		sort = sharedmodels.StarboardSortByMostRated
	default:
		return fiber.NewError(fiber.StatusBadRequest, "invalid sort property")
	}

	entries, err := c.db.GetStarboardEntries(guildID, sort, limit, offset)
	if err != nil && !database.IsErrDatabaseNotFound(err) {
		return err
	}

	results := make([]*models.StarboardEntryResponse, len(entries))

	var i int
	for _, e := range entries {
		if e.Deleted {
			continue
		}

		member, err := discordutil.GetMember(c.session, guildID, e.AuthorID)
		if err != nil {
			continue
		}

		results[i] = &models.StarboardEntryResponse{
			StarboardEntry: e,
			AuthorUsername: member.User.String(),
			AvatarURL:      member.User.AvatarURL(""),
			MessageURL: discordutil.GetMessageLink(&discordgo.Message{
				ChannelID: e.ChannelID,
				ID:        e.MessageID,
			}, guildID),
		}

		i++
	}

	return ctx.JSON(&models.ListResponse{N: i, Data: results[:i]})
}

func (c *GuildsController) getReports(ctx *fiber.Ctx) error {
	uid := ctx.Locals("uid").(string)

	guildID := ctx.Params("guildid")

	offset, err := wsutil.GetQueryInt(ctx, "offset", 0, 0, 0)
	if err != nil {
		return err
	}

	limit, err := wsutil.GetQueryInt(ctx, "limit", 0, 0, 0)
	if err != nil {
		return err
	}

	if memb, _ := c.session.GuildMember(guildID, uid); memb == nil {
		return fiber.ErrNotFound
	}

	var reps []*report.Report

	reps, err = c.db.GetReportsGuild(guildID, offset, limit)
	if err != nil {
		return err
	}

	resReps := make([]*models.Report, 0)
	if reps != nil {
		resReps = make([]*models.Report, len(reps))
		for i, r := range reps {
			resReps[i] = models.ReportFromReport(r, c.cfg.WebServer.PublicAddr)
		}
	}

	return ctx.JSON(&models.ListResponse{N: len(resReps), Data: resReps})
}

func (c *GuildsController) getReportsCount(ctx *fiber.Ctx) error {
	uid := ctx.Locals("uid").(string)

	guildID := ctx.Params("guildid")

	if memb, _ := c.session.GuildMember(guildID, uid); memb == nil {
		return fiber.ErrNotFound
	}

	count, err := c.db.GetReportsGuildCount(guildID)
	if err != nil {
		return err
	}

	return ctx.JSON(&models.Count{Count: count})
}

func (c *GuildsController) getGuildSettings(ctx *fiber.Ctx) error {
	guildID := ctx.Params("guildid")

	gs := new(models.GuildSettings)
	var err error

	if gs.Prefix, err = c.db.GetGuildPrefix(guildID); err != nil && !database.IsErrDatabaseNotFound(err) {
		return err
	}

	if gs.Perms, err = c.db.GetGuildPermissions(guildID); err != nil && !database.IsErrDatabaseNotFound(err) {
		return err
	}

	if gs.AutoRole, err = c.db.GetGuildAutoRole(guildID); err != nil && !database.IsErrDatabaseNotFound(err) {
		return err
	}

	if gs.ModLogChannel, err = c.db.GetGuildModLog(guildID); err != nil && !database.IsErrDatabaseNotFound(err) {
		return err
	}

	if gs.VoiceLogChannel, err = c.db.GetGuildVoiceLog(guildID); err != nil && !database.IsErrDatabaseNotFound(err) {
		return err
	}

	if gs.JoinMessageChannel, gs.JoinMessageText, err = c.db.GetGuildJoinMsg(guildID); err != nil && !database.IsErrDatabaseNotFound(err) {
		return err
	}

	if gs.LeaveMessageChannel, gs.LeaveMessageText, err = c.db.GetGuildLeaveMsg(guildID); err != nil && !database.IsErrDatabaseNotFound(err) {
		return err
	}

	return ctx.JSON(gs)
}

func (c *GuildsController) postGuildSettings(ctx *fiber.Ctx) error {
	uid := ctx.Locals("uid").(string)

	guildID := ctx.Params("guildid")

	var err error

	gs := new(models.GuildSettings)
	if err = ctx.BodyParser(gs); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	if gs.AutoRole != "" {
		if ok, _, err := c.pmw.CheckPermissions(c.session, guildID, uid, "sp.guild.config.autorole"); err != nil {
			return wsutil.ErrInternalOrNotFound(err)
		} else if !ok {
			return fiber.ErrUnauthorized
		}

		if gs.AutoRole == "__RESET__" {
			gs.AutoRole = ""
		}

		if err = c.db.SetGuildAutoRole(guildID, gs.AutoRole); err != nil {
			return wsutil.ErrInternalOrNotFound(err)
		}
	}

	if gs.ModLogChannel != "" {
		if ok, _, err := c.pmw.CheckPermissions(c.session, guildID, uid, "sp.guild.config.modlog"); err != nil {
			return wsutil.ErrInternalOrNotFound(err)
		} else if !ok {
			return fiber.ErrUnauthorized
		}

		if gs.ModLogChannel == "__RESET__" {
			gs.ModLogChannel = ""
		}

		if err = c.db.SetGuildModLog(guildID, gs.ModLogChannel); err != nil {
			return wsutil.ErrInternalOrNotFound(err)
		}
	}

	if gs.Prefix != "" {
		if ok, _, err := c.pmw.CheckPermissions(c.session, guildID, uid, "sp.guild.config.prefix"); err != nil {
			return wsutil.ErrInternalOrNotFound(err)
		} else if !ok {
			return fiber.ErrUnauthorized
		}

		if gs.Prefix == "__RESET__" {
			gs.Prefix = ""
		}

		if err = c.db.SetGuildPrefix(guildID, gs.Prefix); err != nil {
			return wsutil.ErrInternalOrNotFound(err)
		}
	}

	if gs.VoiceLogChannel != "" {
		if ok, _, err := c.pmw.CheckPermissions(c.session, guildID, uid, "sp.guild.config.voicelog"); err != nil {
			return wsutil.ErrInternalOrNotFound(err)
		} else if !ok {
			return fiber.ErrUnauthorized
		}

		if gs.VoiceLogChannel == "__RESET__" {
			gs.VoiceLogChannel = ""
		}

		if err = c.db.SetGuildVoiceLog(guildID, gs.VoiceLogChannel); err != nil {
			return wsutil.ErrInternalOrNotFound(err)
		}
	}

	if gs.JoinMessageChannel != "" && gs.JoinMessageText != "" {
		if ok, _, err := c.pmw.CheckPermissions(c.session, guildID, uid, "sp.guild.config.joinmsg"); err != nil {
			return wsutil.ErrInternalOrNotFound(err)
		} else if !ok {
			return fiber.ErrUnauthorized
		}

		if gs.JoinMessageChannel == "__RESET__" && gs.JoinMessageText == "__RESET__" {
			gs.JoinMessageChannel = ""
			gs.JoinMessageText = ""
		}

		if err = c.db.SetGuildJoinMsg(guildID, gs.JoinMessageChannel, gs.JoinMessageText); err != nil {
			return wsutil.ErrInternalOrNotFound(err)
		}
	}

	if gs.LeaveMessageChannel != "" && gs.LeaveMessageText != "" {
		if ok, _, err := c.pmw.CheckPermissions(c.session, guildID, uid, "sp.guild.config.leavemsg"); err != nil {
			return wsutil.ErrInternalOrNotFound(err)
		} else if !ok {
			return fiber.ErrUnauthorized
		}

		if gs.LeaveMessageChannel == "__RESET__" && gs.LeaveMessageText == "__RESET__" {
			gs.LeaveMessageChannel = ""
			gs.LeaveMessageText = ""
		}

		if err = c.db.SetGuildLeaveMsg(guildID, gs.LeaveMessageChannel, gs.LeaveMessageText); err != nil {
			return wsutil.ErrInternalOrNotFound(err)
		}
	}

	return ctx.JSON(struct{}{})
}

func (c *GuildsController) getGuildPermissions(ctx *fiber.Ctx) error {
	uid := ctx.Locals("uid").(string)

	guildID := ctx.Params("guildid")

	if memb, _ := c.session.GuildMember(guildID, uid); memb == nil {
		return fiber.ErrNotFound
	}

	var perms map[string]permissions.PermissionArray
	var err error

	if perms, err = c.db.GetGuildPermissions(guildID); err != nil && !database.IsErrDatabaseNotFound(err) {
		return err
	}

	return ctx.JSON(perms)
}

func (c *GuildsController) postGuildPermissions(ctx *fiber.Ctx) error {
	guildID := ctx.Params("guildid")

	update := new(models.PermissionsUpdate)
	if err := ctx.BodyParser(update); err != nil {
		return fiber.ErrBadRequest
	}

	sperm := update.Perm[1:]
	if !strings.HasPrefix(sperm, "sp.guild") && !strings.HasPrefix(sperm, "sp.etc") && !strings.HasPrefix(sperm, "sp.chat") {
		return fiber.NewError(fiber.StatusBadRequest, "you can only give permissions over the domains 'sp.guild', 'sp.etc' and 'sp.chat'")
	}

	perms, err := c.db.GetGuildPermissions(guildID)
	if err != nil {
		if database.IsErrDatabaseNotFound(err) {
			return fiber.ErrNotFound
		}
		return err
	}

	for _, roleID := range update.RoleIDs {
		rperms, ok := perms[roleID]
		if !ok {
			rperms = make(permissions.PermissionArray, 0)
		}

		rperms, changed := rperms.Update(update.Perm, false)

		if changed {
			if err = c.db.SetGuildRolePermission(guildID, roleID, rperms); err != nil {
				return err
			}
		}
	}

	return ctx.JSON(struct{}{})
}

func (c *GuildsController) postGuildToggleInviteblock(ctx *fiber.Ctx) error {
	guildID := ctx.Params("guildid")

	var data struct {
		Enabled bool `json:"enabled"`
	}

	if err := ctx.BodyParser(&data); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	val := ""
	if data.Enabled {
		val = "1"
	}

	if err := c.db.SetGuildInviteBlock(guildID, val); err != nil {
		return err
	}

	return ctx.JSON(struct{}{})
}

func (c *GuildsController) getGuildSettingsKarma(ctx *fiber.Ctx) error {
	guildID := ctx.Params("guildid")

	settings := new(models.KarmaSettings)

	var err error

	if settings.State, err = c.db.GetKarmaState(guildID); err != nil && !database.IsErrDatabaseNotFound(err) {
		return err
	}

	if settings.Tokens, err = c.db.GetKarmaTokens(guildID); err != nil && !database.IsErrDatabaseNotFound(err) {
		return err
	}

	emotesInc, emotesDec, err := c.db.GetKarmaEmotes(guildID)
	if err != nil && !database.IsErrDatabaseNotFound(err) {
		return err
	}
	settings.EmotesIncrease = strings.Split(emotesInc, "")
	settings.EmotesDecrease = strings.Split(emotesDec, "")

	return ctx.JSON(settings)
}

func (c *GuildsController) postGuildSettingsKarma(ctx *fiber.Ctx) error {
	guildID := ctx.Params("guildid")

	settings := new(models.KarmaSettings)
	var err error

	if err = ctx.BodyParser(settings); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	if err = c.db.SetKarmaState(guildID, settings.State); err != nil {
		return err
	}

	if !checkEmojis(settings.EmotesIncrease) || !checkEmojis(settings.EmotesDecrease) {
		return fiber.NewError(fiber.StatusBadRequest, "invalid emoji")
	}

	emotesInc := strings.Join(settings.EmotesIncrease, "")
	emotesDec := strings.Join(settings.EmotesDecrease, "")
	if err = c.db.SetKarmaEmotes(guildID, emotesInc, emotesDec); err != nil {
		return err
	}

	if err = c.db.SetKarmaTokens(guildID, settings.Tokens); err != nil {
		return err
	}

	return ctx.JSON(struct{}{})
}

func (c *GuildsController) getGuildSettingsKarmaBlocklist(ctx *fiber.Ctx) error {
	guildID := ctx.Params("guildid")

	idList, err := c.db.GetKarmaBlockList(guildID)
	if err != nil && !database.IsErrDatabaseNotFound(err) {
		return err
	}

	memberList := make([]*models.Member, len(idList))
	var m *discordgo.Member
	var i int
	for _, id := range idList {
		if m, err = discordutil.GetMember(c.session, guildID, id); err != nil {
			continue
		}
		memberList[i] = models.MemberFromMember(m)
		i++
	}

	memberList = memberList[:i]

	return ctx.JSON(&models.ListResponse{N: len(memberList), Data: memberList})
}

func (c *GuildsController) putGuildSettingsKarmaBlocklist(ctx *fiber.Ctx) error {
	guildID := ctx.Params("guildid")
	memberID := ctx.Params("memberid")

	memb, err := fetch.FetchMember(c.session, guildID, memberID)
	if err == fetch.ErrNotFound {
		return fiber.ErrNotFound
	}
	if err != nil {
		return err
	}

	ok, err := c.db.IsKarmaBlockListed(guildID, memb.User.ID)
	if err != nil {
		return err
	}
	if ok {
		return fiber.NewError(fiber.StatusBadRequest, "member is already blocklisted")
	}

	if err = c.db.AddKarmaBlockList(guildID, memb.User.ID); err != nil {
		return err
	}

	return ctx.JSON(struct{}{})
}

func (c *GuildsController) deleteGuildSettingsKarmaBlocklist(ctx *fiber.Ctx) error {
	guildID := ctx.Params("guildid")
	memberID := ctx.Params("memberid")

	ok, err := c.db.IsKarmaBlockListed(guildID, memberID)
	if err != nil {
		return err
	}
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, "member is not blocklisted")
	}

	if err = c.db.RemoveKarmaBlockList(guildID, memberID); err != nil {
		return err
	}

	return ctx.JSON(struct{}{})
}

func (c *GuildsController) getGuildSettingsAntiraid(ctx *fiber.Ctx) error {
	guildID := ctx.Params("guildid")

	settings := new(models.AntiraidSettings)

	var err error
	if settings.State, err = c.db.GetAntiraidState(guildID); err != nil && !database.IsErrDatabaseNotFound(err) {
		return err
	}

	if settings.RegenerationPeriod, err = c.db.GetAntiraidRegeneration(guildID); err != nil && !database.IsErrDatabaseNotFound(err) {
		return err
	}

	if settings.Burst, err = c.db.GetAntiraidBurst(guildID); err != nil && !database.IsErrDatabaseNotFound(err) {
		return err
	}

	return ctx.JSON(settings)
}

func (c *GuildsController) postGuildSettingsAntiraid(ctx *fiber.Ctx) error {
	guildID := ctx.Params("guildid")

	settings := new(models.AntiraidSettings)
	if err := ctx.BodyParser(settings); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	if settings.RegenerationPeriod < 1 {
		return fiber.NewError(fiber.StatusBadRequest, "regeneration period must be larger than 0")
	}
	if settings.Burst < 1 {
		return fiber.NewError(fiber.StatusBadRequest, "burst must be larger than 0")
	}

	var err error

	if err = c.db.SetAntiraidState(guildID, settings.State); err != nil {
		return err
	}

	if err = c.db.SetAntiraidRegeneration(guildID, settings.RegenerationPeriod); err != nil {
		return err
	}

	if err = c.db.SetAntiraidBurst(guildID, settings.Burst); err != nil {
		return err
	}

	return ctx.JSON(struct{}{})
}

func (c *GuildsController) getGuildUnbanrequests(ctx *fiber.Ctx) error {
	guildID := ctx.Params("guildid")

	requests, err := c.db.GetGuildUnbanRequests(guildID)
	if err != nil && !database.IsErrDatabaseNotFound(err) {
		return err
	}
	if requests == nil {
		requests = make([]*report.UnbanRequest, 0)
	}

	for _, r := range requests {
		r.Hydrate()
	}

	return ctx.JSON(&models.ListResponse{N: len(requests), Data: requests})
}

func (c *GuildsController) getGuildUnbanrequestsCount(ctx *fiber.Ctx) error {
	guildID := ctx.Params("guildid")

	stateFilter, err := wsutil.GetQueryInt(ctx, "state", -1, 0, 0)
	if err != nil {
		return err
	}

	requests, err := c.db.GetGuildUnbanRequests(guildID)
	if err != nil && !database.IsErrDatabaseNotFound(err) {
		return err
	}
	if requests == nil {
		requests = make([]*report.UnbanRequest, 0)
	}

	count := len(requests)
	if stateFilter > -1 {
		count = 0
		for _, r := range requests {
			if int(r.Status) == stateFilter {
				count++
			}
		}
	}

	return ctx.JSON(&models.Count{Count: count})
}

func (c *GuildsController) getGuildUnbanrequest(ctx *fiber.Ctx) error {
	guildID := ctx.Params("guildid")
	id := ctx.Params("id")

	request, err := c.db.GetUnbanRequest(id)
	if err != nil && !database.IsErrDatabaseNotFound(err) {
		return err
	}
	if request == nil || request.GuildID != guildID {
		return fiber.ErrNotFound
	}

	return ctx.JSON(request.Hydrate())
}

func (c *GuildsController) postGuildUnbanrequest(ctx *fiber.Ctx) error {
	uid := ctx.Locals("uid").(string)

	guildID := ctx.Params("guildid")
	id := ctx.Params("id")

	rUpdate := new(report.UnbanRequest)
	if err := ctx.BodyParser(rUpdate); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	request, err := c.db.GetUnbanRequest(id)
	if err != nil && !database.IsErrDatabaseNotFound(err) {
		return err
	}
	if request == nil || request.GuildID != guildID {
		return fiber.ErrNotFound
	}

	if rUpdate.ProcessedMessage == "" {
		return fiber.NewError(fiber.StatusBadRequest, "process reason message must be provided")
	}

	if request.ID, err = snowflake.ParseString(id); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	request.ProcessedBy = uid
	request.Status = rUpdate.Status
	request.Processed = time.Now()
	request.ProcessedMessage = rUpdate.ProcessedMessage

	if err = c.db.UpdateUnbanRequest(request); err != nil {
		return err
	}

	if request.Status == report.UnbanRequestStateAccepted {
		if err = c.session.GuildBanDelete(request.GuildID, request.UserID); err != nil {
			return err
		}
	}

	return ctx.JSON(request.Hydrate())
}

// ---------------------------------------------------------------------------
// - HELPERS

func checkEmojis(emojis []string) bool {
	for _, e := range emojis {
		if !isemoji.IsEmojiNonStrict(e) {
			return false
		}
	}
	return true
}
