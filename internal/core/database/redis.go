package database

import (
	"fmt"
	"strings"
	"time"

	"github.com/zekroTJA/shinpuru/internal/core/config"

	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/snowflake"
	"github.com/go-redis/redis"
	"github.com/zekroTJA/shinpuru/internal/core/backup/backupmodels"
	"github.com/zekroTJA/shinpuru/internal/core/permissions"
	"github.com/zekroTJA/shinpuru/internal/core/twitchnotify"
	"github.com/zekroTJA/shinpuru/internal/util/imgstore"
	"github.com/zekroTJA/shinpuru/internal/util/report"
	"github.com/zekroTJA/shinpuru/internal/util/tag"
	"github.com/zekroTJA/shinpuru/internal/util/vote"
)

const (
	keySetting = "PROP"

	keyGuildPrefix        = "GUILD:PREFIX"
	keyGuildAutoRole      = "GUILD:AUTOROLE"
	keyGuildModLog        = "GUILD:MODLOG"
	keyGuildVoiceLog      = "GUILD:VOICELOG"
	keyGuildNotifyRole    = "GUILD:NOTROLE"
	keyGuildGhostPingMsg  = "GUILD:GPMSG"
	keyGuildJDoodleKey    = "GuUILD:JDOODLE"
	keyGuildInviteBlock   = "GUILD:INVBLOCK"
	keyGuildBackupEnabled = "GUILD:BACKUP"
	keyGuildJoinMsg       = "GUILD:JOINMSG"
	keyGuildLeaveMsg      = "GUILD:LEAVEMSG"
	keyGuildMuteRole      = "GUILD:MUTEROLE"

	keyAPISession = "API:SESSION"
)

type RedisMiddleware struct {
	client *redis.Client
	db     Database
}

func NewRedisMiddleware(config *config.DatabaseRedis, db Database) *RedisMiddleware {
	r := &RedisMiddleware{
		db: db,
	}

	r.client = redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.Type,
	})

	return r
}

func (r *RedisMiddleware) Connect(credentials ...interface{}) error {
	return r.db.Connect(credentials...)
}

func (r *RedisMiddleware) Close() {
	r.client.Close()
	r.db.Close()
}

func (r *RedisMiddleware) GetGuildPrefix(guildID string) (string, error) {
	var key = fmt.Sprintf("%s:%s", keyGuildPrefix, guildID)

	val, err := r.client.Get(key).Result()
	if err == redis.Nil {
		val, err = r.db.GetGuildPrefix(guildID)
		if err != nil {
			return "", err
		}

		err = r.client.Set(key, val, 0).Err()
		return val, err
	}
	if err != nil {
		return "", err
	}

	return val, nil
}

func (r *RedisMiddleware) SetGuildPrefix(guildID, newPrefix string) error {
	var key = fmt.Sprintf("%s:%s", keyGuildPrefix, guildID)

	if err := r.client.Set(key, newPrefix, 0).Err(); err != nil {
		return err
	}

	return r.db.SetGuildPrefix(guildID, newPrefix)
}

func (r *RedisMiddleware) GetGuildAutoRole(guildID string) (string, error) {
	var key = fmt.Sprintf("%s:%s", keyGuildAutoRole, guildID)

	val, err := r.client.Get(key).Result()
	if err == redis.Nil {
		val, err = r.db.GetGuildAutoRole(guildID)
		if err != nil {
			return "", err
		}

		err = r.client.Set(key, val, 0).Err()
		return val, err
	}
	if err != nil {
		return "", err
	}

	return val, nil
}

func (r *RedisMiddleware) SetGuildAutoRole(guildID, autoRoleID string) error {
	var key = fmt.Sprintf("%s:%s", keyGuildAutoRole, guildID)

	if err := r.client.Set(key, autoRoleID, 0).Err(); err != nil {
		return err
	}

	return r.db.SetGuildAutoRole(guildID, autoRoleID)
}

func (r *RedisMiddleware) GetGuildModLog(guildID string) (string, error) {
	var key = fmt.Sprintf("%s:%s", keyGuildModLog, guildID)

	val, err := r.client.Get(key).Result()
	if err == redis.Nil {
		val, err = r.db.GetGuildModLog(guildID)
		if err != nil {
			return "", err
		}

		err = r.client.Set(key, val, 0).Err()
		return val, err
	}
	if err != nil {
		return "", err
	}

	return val, nil
}

func (r *RedisMiddleware) SetGuildModLog(guildID, chanID string) error {
	var key = fmt.Sprintf("%s:%s", keyGuildModLog, guildID)

	if err := r.client.Set(key, chanID, 0).Err(); err != nil {
		return err
	}

	return r.db.SetGuildModLog(guildID, chanID)
}

func (r *RedisMiddleware) GetGuildVoiceLog(guildID string) (string, error) {
	var key = fmt.Sprintf("%s:%s", keyGuildVoiceLog, guildID)

	val, err := r.client.Get(key).Result()
	if err == redis.Nil {
		val, err = r.db.GetGuildVoiceLog(guildID)
		if err != nil {
			return "", err
		}

		err = r.client.Set(key, val, 0).Err()
		return val, err
	}
	if err != nil {
		return "", err
	}

	return val, nil
}

func (r *RedisMiddleware) SetGuildVoiceLog(guildID, chanID string) error {
	var key = fmt.Sprintf("%s:%s", keyGuildVoiceLog, guildID)

	if err := r.client.Set(key, chanID, 0).Err(); err != nil {
		return err
	}

	return r.db.SetGuildVoiceLog(guildID, chanID)
}

func (r *RedisMiddleware) GetGuildNotifyRole(guildID string) (string, error) {
	var key = fmt.Sprintf("%s:%s", keyGuildNotifyRole, guildID)

	val, err := r.client.Get(key).Result()
	if err == redis.Nil {
		val, err = r.db.GetGuildNotifyRole(guildID)
		if err != nil {
			return "", err
		}

		err = r.client.Set(key, val, 0).Err()
		return val, err
	}
	if err != nil {
		return "", err
	}

	return val, nil
}

func (r *RedisMiddleware) SetGuildNotifyRole(guildID, roleID string) error {
	var key = fmt.Sprintf("%s:%s", keyGuildNotifyRole, guildID)

	if err := r.client.Set(key, roleID, 0).Err(); err != nil {
		return err
	}

	return r.db.SetGuildNotifyRole(guildID, roleID)
}

func (r *RedisMiddleware) GetGuildGhostpingMsg(guildID string) (string, error) {
	var key = fmt.Sprintf("%s:%s", keyGuildGhostPingMsg, guildID)

	val, err := r.client.Get(key).Result()
	if err == redis.Nil {
		val, err = r.db.GetGuildGhostpingMsg(guildID)
		if err != nil {
			return "", err
		}

		err = r.client.Set(key, val, 0).Err()
		return val, err
	}
	if err != nil {
		return "", err
	}

	return val, nil
}

func (r *RedisMiddleware) SetGuildGhostpingMsg(guildID, msg string) error {
	var key = fmt.Sprintf("%s:%s", keyGuildGhostPingMsg, guildID)

	if err := r.client.Set(key, msg, 0).Err(); err != nil {
		return err
	}

	return r.db.SetGuildGhostpingMsg(guildID, msg)
}

func (r *RedisMiddleware) GetGuildPermissions(guildID string) (map[string]permissions.PermissionArray, error) {
	return r.db.GetGuildPermissions(guildID)
}

func (r *RedisMiddleware) SetGuildRolePermission(guildID, roleID string, p permissions.PermissionArray) error {
	return r.db.SetGuildRolePermission(guildID, roleID, p)
}

func (r *RedisMiddleware) GetMemberPermission(s *discordgo.Session, guildID string, memberID string) (permissions.PermissionArray, error) {
	return r.db.GetMemberPermission(s, guildID, memberID)
}

func (r *RedisMiddleware) GetGuildJdoodleKey(guildID string) (string, error) {
	var key = fmt.Sprintf("%s:%s", keyGuildJDoodleKey, guildID)

	val, err := r.client.Get(key).Result()
	if err == redis.Nil {
		val, err = r.db.GetGuildJdoodleKey(guildID)
		if err != nil {
			return "", err
		}

		err = r.client.Set(key, val, 0).Err()
		return val, err
	}
	if err != nil {
		return "", err
	}

	return val, nil
}

func (r *RedisMiddleware) SetGuildJdoodleKey(guildID, jdkey string) error {
	var key = fmt.Sprintf("%s:%s", keyGuildJDoodleKey, guildID)

	if err := r.client.Set(key, jdkey, 0).Err(); err != nil {
		return err
	}

	return r.db.SetGuildJdoodleKey(guildID, jdkey)
}

func (r *RedisMiddleware) GetGuildBackup(guildID string) (bool, error) {
	var key = fmt.Sprintf("%s:%s", keyGuildBackupEnabled, guildID)

	var val bool
	err := r.client.Get(key).Scan(&val)
	if err == redis.Nil {
		val, err = r.db.GetGuildBackup(guildID)
		if err != nil {
			return false, err
		}

		err = r.client.Set(key, val, 0).Err()
		return val, err
	}
	if err != nil {
		return false, err
	}

	return val, nil
}

func (r *RedisMiddleware) SetGuildBackup(guildID string, enabled bool) error {
	var key = fmt.Sprintf("%s:%s", keyGuildBackupEnabled, guildID)

	if err := r.client.Set(key, enabled, 0).Err(); err != nil {
		return err
	}

	return r.db.SetGuildBackup(guildID, enabled)
}

func (r *RedisMiddleware) GetGuildInviteBlock(guildID string) (string, error) {
	var key = fmt.Sprintf("%s:%s", keyGuildInviteBlock, guildID)

	val, err := r.client.Get(key).Result()
	if err == redis.Nil {
		val, err = r.db.GetGuildInviteBlock(guildID)
		if err != nil {
			return "", err
		}

		err = r.client.Set(key, val, 0).Err()
		return val, err
	}
	if err != nil {
		return "", err
	}

	return val, nil
}

func (r *RedisMiddleware) SetGuildInviteBlock(guildID string, data string) error {
	var key = fmt.Sprintf("%s:%s", keyGuildInviteBlock, guildID)

	if err := r.client.Set(key, data, 0).Err(); err != nil {
		return err
	}

	return r.db.SetGuildInviteBlock(guildID, data)
}

func (r *RedisMiddleware) GetGuildJoinMsg(guildID string) (string, string, error) {
	var key = fmt.Sprintf("%s:%s", keyGuildJoinMsg, guildID)

	var val1, val2 string

	raw, err := r.client.Get(key).Result()
	if err == redis.Nil {
		val1, val2, err = r.db.GetGuildJoinMsg(guildID)
		if err != nil {
			return "", "", err
		}

		err = r.client.Set(key, fmt.Sprintf("%s|%s", val1, val2), 0).Err()
		return val1, val2, err
	}
	if err != nil {
		return "", "", err
	}

	rawSplit := strings.Split(raw, "|")
	val1, val2 = rawSplit[0], rawSplit[1]

	return val1, val2, nil
}

func (r *RedisMiddleware) SetGuildJoinMsg(guildID string, channelID string, msg string) error {
	var key = fmt.Sprintf("%s:%s", keyGuildJoinMsg, guildID)

	if err := r.client.Set(key, fmt.Sprintf("%s|%s", channelID, msg), 0).Err(); err != nil {
		return err
	}

	return r.db.SetGuildJoinMsg(guildID, channelID, msg)
}

func (r *RedisMiddleware) GetGuildLeaveMsg(guildID string) (string, string, error) {
	var key = fmt.Sprintf("%s:%s", keyGuildLeaveMsg, guildID)

	var val1, val2 string

	raw, err := r.client.Get(key).Result()
	if err == redis.Nil {
		val1, val2, err = r.db.GetGuildLeaveMsg(guildID)
		if err != nil {
			return "", "", err
		}

		err = r.client.Set(key, fmt.Sprintf("%s|%s", val1, val2), 0).Err()
		return val1, val2, err
	}
	if err != nil {
		return "", "", err
	}

	rawSplit := strings.Split(raw, "|")
	val1, val2 = rawSplit[0], rawSplit[1]

	return val1, val2, nil
}

func (r *RedisMiddleware) SetGuildLeaveMsg(guildID string, channelID string, msg string) error {
	var key = fmt.Sprintf("%s:%s", keyGuildLeaveMsg, guildID)

	if err := r.client.Set(key, fmt.Sprintf("%s|%s", channelID, msg), 0).Err(); err != nil {
		return err
	}

	return r.db.SetGuildLeaveMsg(guildID, channelID, msg)
}

func (r *RedisMiddleware) AddReport(rep *report.Report) error {
	return r.db.AddReport(rep)
}

func (r *RedisMiddleware) DeleteReport(id snowflake.ID) error {
	return r.db.DeleteReport(id)
}

func (r *RedisMiddleware) GetReport(id snowflake.ID) (*report.Report, error) {
	return r.db.GetReport(id)
}

func (r *RedisMiddleware) GetReportsGuild(guildID string, offset, limit int) ([]*report.Report, error) {
	return r.db.GetReportsGuild(guildID, offset, limit)
}

func (r *RedisMiddleware) GetReportsFiltered(guildID, memberID string, repType int) ([]*report.Report, error) {
	return r.db.GetReportsFiltered(guildID, memberID, repType)
}

func (r *RedisMiddleware) GetReportsGuildCount(guildID string) (int, error) {
	return r.db.GetReportsGuildCount(guildID)
}

func (r *RedisMiddleware) GetReportsFilteredCount(guildID, memberID string, repType int) (int, error) {
	return r.db.GetReportsFilteredCount(guildID, memberID, repType)
}

func (r *RedisMiddleware) GetSetting(setting string) (string, error) {
	var key = fmt.Sprintf("%s:%s", keySetting, setting)

	val, err := r.client.Get(key).Result()
	if err == redis.Nil {
		val, err = r.db.GetSetting(setting)
		if err != nil {
			return "", err
		}

		err = r.client.Set(key, val, 0).Err()
		return val, err
	}
	if err != nil {
		return "", err
	}

	return val, nil
}

func (r *RedisMiddleware) SetSetting(setting, value string) error {
	var key = fmt.Sprintf("%s:%s", keySetting, setting)

	if err := r.client.Set(key, value, 0).Err(); err != nil {
		return err
	}

	return r.db.SetSetting(setting, value)
}

func (r *RedisMiddleware) GetVotes() (map[string]*vote.Vote, error) {
	return r.db.GetVotes()
}

func (r *RedisMiddleware) AddUpdateVote(votes *vote.Vote) error {
	return r.db.AddUpdateVote(votes)
}

func (r *RedisMiddleware) DeleteVote(voteID string) error {
	return r.db.DeleteVote(voteID)
}

func (r *RedisMiddleware) GetMuteRoles() (map[string]string, error) {
	return r.db.GetMuteRoles()
}

func (r *RedisMiddleware) GetMuteRoleGuild(guildID string) (string, error) {
	var key = fmt.Sprintf("%s:%s", keyGuildMuteRole, guildID)

	val, err := r.client.Get(key).Result()
	if err == redis.Nil {
		val, err = r.db.GetMuteRoleGuild(guildID)
		if err != nil {
			return "", err
		}

		err = r.client.Set(key, val, 0).Err()
		return val, err
	}
	if err != nil {
		return "", err
	}

	return val, nil
}

func (r *RedisMiddleware) SetMuteRole(guildID, roleID string) error {
	var key = fmt.Sprintf("%s:%s", keyGuildMuteRole, guildID)

	if err := r.client.Set(key, roleID, 0).Err(); err != nil {
		return err
	}

	return r.db.SetMuteRole(guildID, roleID)
}

func (r *RedisMiddleware) GetAllTwitchNotifies(twitchUserID string) ([]*twitchnotify.TwitchNotifyDBEntry, error) {
	return r.db.GetAllTwitchNotifies(twitchUserID)
}

func (r *RedisMiddleware) GetTwitchNotify(twitchUserID, guildID string) (*twitchnotify.TwitchNotifyDBEntry, error) {
	return r.db.GetTwitchNotify(twitchUserID, guildID)
}

func (r *RedisMiddleware) SetTwitchNotify(twitchNotify *twitchnotify.TwitchNotifyDBEntry) error {
	return r.db.SetTwitchNotify(twitchNotify)
}

func (r *RedisMiddleware) DeleteTwitchNotify(twitchUserID, guildID string) error {
	return r.db.DeleteTwitchNotify(twitchUserID, guildID)
}

func (r *RedisMiddleware) AddBackup(guildID, fileID string) error {
	return r.db.AddBackup(guildID, fileID)
}

func (r *RedisMiddleware) DeleteBackup(guildID, fileID string) error {
	return r.db.DeleteBackup(guildID, fileID)
}

func (r *RedisMiddleware) GetBackups(guildID string) ([]*backupmodels.Entry, error) {
	return r.db.GetBackups(guildID)
}

func (r *RedisMiddleware) GetGuilds() ([]string, error) {
	return r.db.GetGuilds()
}

func (r *RedisMiddleware) AddTag(tag *tag.Tag) error {
	return r.db.AddTag(tag)
}

func (r *RedisMiddleware) EditTag(tag *tag.Tag) error {
	return r.db.EditTag(tag)
}

func (r *RedisMiddleware) GetTagByID(id snowflake.ID) (*tag.Tag, error) {
	return r.db.GetTagByID(id)
}

func (r *RedisMiddleware) GetTagByIdent(ident string, guildID string) (*tag.Tag, error) {
	return r.db.GetTagByIdent(ident, guildID)
}

func (r *RedisMiddleware) GetGuildTags(guildID string) ([]*tag.Tag, error) {
	return r.db.GetGuildTags(guildID)
}

func (r *RedisMiddleware) DeleteTag(id snowflake.ID) error {
	return r.db.DeleteTag(id)
}

func (r *RedisMiddleware) SetSession(sessionKey, userID string, expires time.Time) error {
	var keyByKey = fmt.Sprintf("%s:K:%s", keyAPISession, sessionKey)
	var keyByUser = fmt.Sprintf("%s:U:%s", keyAPISession, userID)

	if err := r.client.Set(keyByKey, userID, time.Until(expires)).Err(); err != nil {
		return err
	}

	if err := r.client.Set(keyByUser, sessionKey, time.Until(expires)).Err(); err != nil {
		return err
	}

	return r.db.SetSession(sessionKey, userID, expires)
}

func (r *RedisMiddleware) GetSession(sessionKey string) (string, error) {
	var keyByKey = fmt.Sprintf("%s:K:%s", keyAPISession, sessionKey)

	val, err := r.client.Get(keyByKey).Result()
	if err == redis.Nil {
		val, err = r.db.GetSession(sessionKey)
		if err != nil {
			return "", err
		}

		keyByUser := fmt.Sprintf("%s:U:%s", keyAPISession, val)
		if err = r.client.Set(keyByKey, val, 0).Err(); err != nil {
			return "", err
		}
		err = r.client.Set(keyByUser, sessionKey, 0).Err()
		return val, err
	}
	if err != nil {
		return "", err
	}

	return val, nil
}

func (r *RedisMiddleware) DeleteSession(userID string) error {
	var keyByUser = fmt.Sprintf("%s:U:%s", keyAPISession, userID)

	sessionKey, err := r.client.Get(keyByUser).Result()
	if err == redis.Nil {
		return r.db.DeleteSession(userID)
	}
	if err != nil {
		return err
	}

	keyByKey := fmt.Sprintf("%s:K:%s", keyAPISession, sessionKey)
	if err := r.client.Del(keyByKey).Err(); err != nil && err != redis.Nil {
		return err
	}

	return r.db.DeleteSession(userID)
}

func (r *RedisMiddleware) GetImageData(id snowflake.ID) (*imgstore.Image, error) {
	return r.db.GetImageData(id)
}

func (r *RedisMiddleware) SaveImageData(image *imgstore.Image) error {
	return r.db.SaveImageData(image)
}

func (r *RedisMiddleware) RemoveImageData(id snowflake.ID) error {
	return r.db.RemoveImageData(id)
}
