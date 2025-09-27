package controllers

import (
	"electric_payment/config"
	"electric_payment/model"
	"fmt"
	"strings"
	"time"

	"go.mau.fi/whatsmeow/types/events"
)

// PermissionRule defines a set of rules for command permissions, including custom allow logic,
// denial messages, daily usage quota, and cooldown period between uses.
//
// Fields:
//   - AllowFunc: A function that determines if a user is allowed to execute the command.
//     Returns a boolean indicating permission, and custom denial messages in English and Indonesian.
//   - DenyMessageID: Custom denial message in Indonesian.
//   - DenyMessageEN: Custom denial message in English.
//   - MaxDailyQuota: Maximum number of times the command can be used per day.
//   - CooldownSeconds: Minimum number of seconds required between command uses.
type PermissionRule struct {
	AllowFunc       func(user *model.WAPhoneUser) (bool, string, string) // returns: (allowed, customMsgEN, customMsgID)
	DenyMessageID   string
	DenyMessageEN   string
	MaxDailyQuota   int // e.g. 5 times per day
	CooldownSeconds int // e.g. 10 sec cooldown
}

// PromptPermissionResult represents the result of a permission check for a command prompt,
// including whether the action is allowed, an optional message, remaining daily uses, and cooldown time.
//
// Fields:
//   - Allowed: Indicates if the action is permitted.
//   - Message: Denial message or empty if allowed.
//   - UsesLeft: Number of times the command can still be used today.
//   - CooldownLeft: Seconds remaining before the command can be used again.
type PromptPermissionResult struct {
	Allowed      bool
	Message      string // deny message or empty if allowed
	UsesLeft     int    // how many times can still use today
	CooldownLeft int    // seconds left before allowed again
}

// getUsageKey builds Redis key like: perm:usage:{cmd}:{userID}:{yyyy-mm-dd}
func getUsageKey(cmd string, userID uint) string {
	date := time.Now().Format("2006-01-02")
	return fmt.Sprintf("perm:usage:%s:%d:%s", cmd, userID, date)
}

// getCooldownKey builds Redis key like: perm:cooldown:{cmd}:{userID}
func getCooldownKey(cmd string, userID uint) string {
	return fmt.Sprintf("perm:cooldown:%s:%d", cmd, userID)
}

// CheckPromptPermission evaluates whether a user has permission to execute a specific command prompt.
// It checks against a set of predefined permission rules, including user type, phone number, ban status,
// daily quota limits, and cooldown periods. If the command is not found in the rules, it performs additional
// checks for bad words, work order numbers, and technician IDs for CSNA users. The function returns a
// PromptPermissionResult indicating whether the command is allowed, along with appropriate messages and
// quota/cooldown information.
//
// Parameters:
//
//	v        - The incoming WhatsApp message event.
//	cmd      - The command string to check permission for.
//	user     - The user attempting to execute the command.
//	userLang - The user's language preference ("en" or "id").
//
// Returns:
//
//	PromptPermissionResult - Struct containing permission status, messages, quota, and cooldown info.
func CheckPromptPermission(v *events.Message, cmd string, user *model.WAPhoneUser, userLang string) PromptPermissionResult {
	rules := map[string]PermissionRule{
		"ping": {
			AllowFunc: func(u *model.WAPhoneUser) (bool, string, string) {
				if u.UserType == model.WaBotSuperUser ||
					u.PhoneNumber == config.GetConfig().Whatsmeow.WaSuperUser ||
					u.PhoneNumber == config.GetConfig().Whatsmeow.WaSupport ||
					u.PhoneNumber == config.GetConfig().Whatsmeow.WaTechnicalSupport {
					return true, "", ""
				}
				return false, "❌ You are not allowed to use ping.", "❌ Anda tidak punya izin untuk ping."
			},
			MaxDailyQuota:   25,
			CooldownSeconds: 5,
			DenyMessageEN:   "❌ You’re not allowed, quota used up, or too fast.",
			DenyMessageID:   "❌ Anda tidak punya izin untuk ping, quota habis, atau terlalu cepat.",
		},

		// /* Special */
		// "/all-cmd": {
		// 	AllowFunc: func(u *model.WAPhoneUser) (bool, string, string) {
		// 		if u.UserType == model.WaBotSuperUser || u.PhoneNumber == config.GetConfig().Whatsmeow.WaSuperUser {
		// 			return true, "", ""
		// 		}
		// 		return false, "❌ You are not allowed to see all commands.", "❌ Anda tidak berhak melihat semua perintah."
		// 	},
		// 	MaxDailyQuota:   20,
		// 	CooldownSeconds: 10,
		// 	DenyMessageEN:   "❌ You are not allowed to see all commands (quota exceeded or too frequent).",
		// 	DenyMessageID:   "❌ Anda tidak berhak melihat semua perintah (quota habis atau terlalu cepat).",
		// },
	}

	originalSenderJID := NormalizeSenderJID(v.Info.Sender.String())
	stanzaID := v.Info.ID
	_ = originalSenderJID
	_ = stanzaID

	rule, ok := rules[cmd]
	if !ok {
		// Check bad words
		found, banned, warn := CheckAndTrackBadWords(user.ID, cmd, userLang)
		if found {
			if banned {
				return PromptPermissionResult{
					Allowed: false,
					Message: warn,
				}
			}
			return PromptPermissionResult{
				Allowed: false,
				Message: warn,
			}
		}

		if user.UserOf == model.UserOfPLTMH {
			// TODO: add other special checks for PLTMH users if needed
		}

		// No rule or static command prompt for wa bot found
		msg := config.GetConfig().Whatsmeow.WaErrorMessage.EN.UnknownPrompt
		if userLang == "id" {
			msg = config.GetConfig().Whatsmeow.WaErrorMessage.ID.UnknownPrompt
		}
		_ = msg
		// return PromptPermissionResult{
		// 	Allowed: false,
		// 	Message: msg,
		// }
		// Uncomment this if you want to return unknown prompt message

		return PromptPermissionResult{}

	}

	// check AllowFunc
	allowed, customEN, customID := rule.AllowFunc(user)
	if !allowed {
		msg := customEN
		if userLang == "id" {
			msg = customID
		}
		if msg == "" { // fallback
			if userLang == "id" {
				msg = rule.DenyMessageID
			} else {
				msg = rule.DenyMessageEN
			}
		}
		return PromptPermissionResult{
			Allowed: false,
			Message: msg,
		}
	}

	userID := user.ID
	usesLeft := rule.MaxDailyQuota
	cooldownLeft := 0

	// cooldown check
	if rule.CooldownSeconds > 0 {
		cooldownKey := getCooldownKey(cmd, userID)
		ttl, _ := rdb.TTL(contx, cooldownKey).Result()
		if ttl > 0 {
			cooldownLeft = int(ttl.Seconds())
			msg := fmt.Sprintf("⏱ Please wait %d seconds before using *%s* command again.", cooldownLeft, cmd)
			if userLang == "id" {
				msg = fmt.Sprintf("⏱ Tunggu %d detik sebelum memakai perintah *%s* lagi.", cooldownLeft, cmd)
			}
			return PromptPermissionResult{
				Allowed:      false,
				Message:      msg,
				CooldownLeft: cooldownLeft,
			}
		}
	}

	// quota check
	if rule.MaxDailyQuota > 0 {
		usageKey := getUsageKey(cmd, userID)
		count, _ := rdb.Get(contx, usageKey).Int()
		usesLeft = rule.MaxDailyQuota - count
		if count >= rule.MaxDailyQuota {
			ttl, _ := rdb.TTL(contx, usageKey).Result()
			hours := int(ttl.Hours())
			minutes := int(ttl.Minutes()) % 60

			var msg string
			if userLang == "id" {
				msg = fmt.Sprintf("🚫 Anda telah mencapai batas harian untuk perintah *%s*. Coba lagi dalam %dj %dm.", cmd, hours, minutes)
			} else {
				msg = fmt.Sprintf("🚫 You have reached your daily limit for command *%s*. Try again in %dh %dm.", cmd, hours, minutes)
			}

			return PromptPermissionResult{
				Allowed:  false,
				Message:  msg,
				UsesLeft: 0,
			}
		}
	}

	// passed: increment usage & set cooldown
	pipe := rdb.TxPipeline()
	if rule.MaxDailyQuota > 0 {
		usageKey := getUsageKey(cmd, userID)
		pipe.Incr(contx, usageKey)
		pipe.Expire(contx, usageKey, time.Duration(config.GetConfig().Whatsmeow.RedisExpiry)*time.Hour)
	}
	if rule.CooldownSeconds > 0 {
		cooldownKey := getCooldownKey(cmd, userID)
		pipe.Set(contx, cooldownKey, "1", time.Duration(rule.CooldownSeconds)*time.Second)
	}
	_, _ = pipe.Exec(contx)

	return PromptPermissionResult{
		Allowed:      true,
		UsesLeft:     usesLeft - 1, // after this use
		CooldownLeft: 0,
	}
}

// CheckAndTrackBadWords checks if cmd has bad words, increments counter in Redis,
// and returns: isBadWordFound, isBannedNow, warnMessage (based on userLang)
func CheckAndTrackBadWords(userID uint, cmd, userLang string) (bool, bool, string) {
	// Load enabled bad words from DB (simple version)
	var badWords []model.BadWord
	if err := dbWeb.Where("is_enabled = ?", true).Find(&badWords).Error; err != nil {
		return false, false, ""
	}

	// check if cmd contains bad words
	loweredCmd := strings.ToLower(cmd)
	found := false
	for _, bw := range badWords {
		if bw.Language == userLang && strings.Contains(loweredCmd, bw.Word) {
			found = true
			break
		}
	}

	if !found {
		return false, false, ""
	}

	// build redis key
	key := fmt.Sprintf("badword:user:%d", userID)

	// increment strike
	strike, err := rdb.Incr(contx, key).Result()
	if err != nil {
		return true, false, warnMessage(userLang, 0) // fallback
	}

	// set expiry if first time
	if strike == 1 {
		rdb.Expire(contx, key, 24*time.Hour)
	}

	// check if reached max
	if strike >= int64(config.GetConfig().Whatsmeow.WhatsappMaxBadWordStrike) {
		// optionally ban user in DB
		dbWeb.Model(&model.WAPhoneUser{}).
			Where("id = ?", userID).
			Updates(map[string]interface{}{
				"is_banned":       true,
				"allowed_to_call": false,
			})
		var banMsg string
		if userLang == "id" {
			banMsg = config.GetConfig().Whatsmeow.WaErrorMessage.ID.AccountBannedCozBadWord
		} else {
			banMsg = config.GetConfig().Whatsmeow.WaErrorMessage.EN.AccountBannedCozBadWord
		}
		return true, true, banMsg
	}

	return true, false, warnMessage(userLang, int(strike))
}

func warnMessage(lang string, strike int) string {
	if lang == "id" {
		return fmt.Sprintf("⚠️ Peringatan: kata kasar terdeteksi! (%d/%d). Jika mencapai %d, akun Anda akan diblokir.", strike, config.GetConfig().Whatsmeow.WhatsappMaxBadWordStrike, config.GetConfig().Whatsmeow.WhatsappMaxBadWordStrike)
	}
	return fmt.Sprintf("⚠️ Warning: bad words detected! (%d/%d). If it reaches %d, your account will be banned.", strike, config.GetConfig().Whatsmeow.WhatsappMaxBadWordStrike, config.GetConfig().Whatsmeow.WhatsappMaxBadWordStrike)
}
