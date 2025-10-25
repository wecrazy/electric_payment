package controllers

import (
	"electric_payment/config"
	"electric_payment/fun"
	"electric_payment/model"
	"electric_payment/webguibuilder"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func ComponentPage(db *gorm.DB, redisDB *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookies := c.Request.Cookies()

		// Parse JWT token from cookie
		tokenString, err := c.Cookie("token")
		if err != nil {
			fun.ClearCookiesAndRedirect(c, cookies)
			return
		}
		tokenString = strings.ReplaceAll(tokenString, " ", "+")

		decrypted, err := fun.GetAESDecrypted(tokenString)
		if err != nil {
			logrus.Error("Error during decryption:", err)
			fun.ClearCookiesAndRedirect(c, cookies)
			return
		}
		var claims map[string]interface{}
		err = json.Unmarshal(decrypted, &claims)
		if err != nil {
			logrus.Error("Error converting JSON to map:", err)
			fun.ClearCookiesAndRedirect(c, cookies)
			return
		}
		componentID := c.Param("component")
		componentID = strings.ReplaceAll(componentID, "/", "")
		componentID = strings.ReplaceAll(componentID, "..", "")
		componentPrv, ok := claims[componentID]
		if !ok {
			fun.ClearCookiesAndRedirect(c, cookies)
			return
		}
		componentPrvStr, ok := componentPrv.(string)
		if !ok || componentPrvStr == "" {
			fun.ClearCookiesAndRedirect(c, cookies)
			return
		}
		if string(componentPrvStr[1:2]) != "1" {
			fun.ClearCookiesAndRedirect(c, cookies)
			return
		}
		var admin model.Admin
		db.Where("id = ?", uint(claims["id"].(float64))).Find(&admin)

		imageMaps := map[string]interface{}{
			"t":  fun.GenerateRandomString(3),
			"id": admin.ID,
		}
		pathString, err := fun.GetAESEcryptedURLfromJSON(imageMaps)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not encripting image " + err.Error()})
			return
		}
		profile_image := "/profile/default.jpg?f=" + pathString

		var adminStatusData model.AdminStatus
		if err := db.First(&adminStatusData, admin.Status).Error; err != nil {
			logrus.Errorf("failed to parse data status for admin: %v", err)
		}
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("<div class='%s'>", adminStatusData.ClassName))
		sb.WriteString(adminStatusData.Title)
		sb.WriteString("</div>")
		adminStatusHTML := sb.String()

		// Check if admin role has specific app configuration
		var appConfig model.AppConfig
		var appName, appLogo, appVersion, appVersionNo, appVersionCode, appVersionName string

		if err := db.Where("role_id = ? AND is_active = ?", admin.Role, true).First(&appConfig).Error; err == nil {
			// Use role-specific app configuration
			appName = appConfig.AppName
			appLogo = appConfig.AppLogo
			appVersion = appConfig.AppVersion
			appVersionNo = appConfig.VersionNo
			appVersionCode = appConfig.VersionCode
			appVersionName = appConfig.VersionName
			// logrus.Infof("Using role-specific app config for role %d: %s", admin.Role, appConfig.AppName)
		} else {
			// Fallback to default config
			appName = config.GetConfig().App.Name
			appLogo = config.GetConfig().App.Logo
			appVersion = config.GetConfig().App.Version
			appVersionNo = strconv.Itoa(config.GetConfig().App.VersionNo)
			appVersionCode = config.GetConfig().App.VersionCode
			appVersionName = config.GetConfig().App.VersionName
			// logrus.Infof("Using default app config for role %d (no specific config found)", admin.Role)
		}

		replacements := map[string]any{
			"APP_NAME":         appName,
			"APP_LOGO":         appLogo,
			"APP_VERSION":      appVersion,
			"APP_VERSION_NO":   appVersionNo,
			"APP_VERSION_CODE": appVersionCode,
			"APP_VERSION_NAME": appVersionName,
			"APP_WEB_URL":      config.GetConfig().App.WebPublicURL,
			"fullname":         admin.Fullname,
			"username":         admin.Username,
			"userid":           admin.ID,
			"phone":            admin.Phone,
			"email":            admin.Email,
			"role_name":        claims["role_name"].(string),
			// "status_name":      claims["status_name"].(string),
			"status_name":    template.HTML(adminStatusHTML),
			"last_login":     claims["last_login"].(string),
			"created_at_str": claims["created_at_str"].(string),
			"profile_image":  profile_image,
			"ip":             admin.IP,
			"GLOBAL_URL":     fun.GLOBAL_URL,
			/* Whatsmeow */
			"REFRESH_WHATSAPP_QRCODE":                      fun.GLOBAL_URL + "web/" + fun.GetRedis("web:"+admin.Session, redisDB) + "/tab-whatsapp/refresh-qrcode",
			"QR_CODE":                                      "log-data/qrcode.txt",
			"PING_BOT":                                     fun.GLOBAL_URL + "web/" + fun.GetRedis("web:"+admin.Session, redisDB) + "/tab-whatsapp/ping",
			"SEND_TEXT_BOT":                                fun.GLOBAL_URL + "web/" + fun.GetRedis("web:"+admin.Session, redisDB) + "/tab-whatsapp/send_text",
			"SEND_IMAGE_BOT":                               fun.GLOBAL_URL + "web/" + fun.GetRedis("web:"+admin.Session, redisDB) + "/tab-whatsapp/send_image",
			"SEND_DOCUMENT_BOT":                            fun.GLOBAL_URL + "web/" + fun.GetRedis("web:"+admin.Session, redisDB) + "/tab-whatsapp/send_document",
			"SEND_LOCATION_BOT":                            fun.GLOBAL_URL + "web/" + fun.GetRedis("web:"+admin.Session, redisDB) + "/tab-whatsapp/send_location",
			"SEND_POLLING_BOT":                             fun.GLOBAL_URL + "web/" + fun.GetRedis("web:"+admin.Session, redisDB) + "/tab-whatsapp/send_polling",
			"WAG_JSON":                                     fun.GLOBAL_URL + "web/" + fun.GetRedis("web:"+admin.Session, redisDB) + "/tab-whatsapp/groups",
			"TABLE_WHATSAPP_BOT_LANGUAGE":                  webguibuilder.TABLE_WHATSAPP_BOT_LANGUAGE(admin.Session, redisDB, db),
			"TABLE_WHATSAPP_BOT_MESSAGE_REPLY":             webguibuilder.TABLE_WHATSAPP_BOT_MESSAGE_REPLY(admin.Session, redisDB, db),
			"TABLE_WHATSAPP_USER_MANAGEMENT":               webguibuilder.TABLE_WHATSAPP_USER_MANAGEMENT(admin.Session, redisDB, db),
			"END_SESSION_WHATSAPP":                         fun.GLOBAL_URL + "web/" + fun.GetRedis("web:"+admin.Session, redisDB) + "/tab-whatsapp/end-session",
			"TABLE_WHATSAPP_BOT_LOG_MSG_RECEIVED":          webguibuilder.TABLE_WHATSAPP_BOT_LOG_MSG_RECEIVED(admin.Session, redisDB, db),
			"ENDPOINT_TABLE_WHATSAPP_BOT_LOG_MSG_RECEIVED": fun.GLOBAL_URL + "web/" + fun.GetRedis("web:"+admin.Session, redisDB) + "/tab-whatsapp/wa_log_msg_received",
			"RESET_QUOTA_PROMPT":                           fun.GLOBAL_URL + "web/" + fun.GetRedis("web:"+admin.Session, redisDB) + "/tab-whatsapp-user-management/reset_quota_prompt",
			"UNBAN_USER_ENDPOINT":                          fun.GLOBAL_URL + "web/" + fun.GetRedis("web:"+admin.Session, redisDB) + "/tab-whatsapp-user-management/unban_user",

			/* App Config */
			"TABLE_APP_CONFIGURATION": webguibuilder.TABLE_APP_CONFIGURATION(admin.Session, redisDB),

			/* Whatsapp Conversation */
			"IsUserWhatsappLoggedInEndpoint":  fun.GLOBAL_URL + "web/" + fun.GetRedis("web:"+admin.Session, redisDB) + "/tab-whatsapp-conversation/is_user_logged_in/" + strconv.Itoa(int(admin.ID)),
			"CheckUserWAStatusEndpoint":       fun.GLOBAL_URL + "web/" + fun.GetRedis("web:"+admin.Session, redisDB) + "/tab-whatsapp-conversation/" + strconv.Itoa(int(admin.ID)) + "/status",
			"GetDetailedUserWAStatusEndpoint": fun.GLOBAL_URL + "web/" + fun.GetRedis("web:"+admin.Session, redisDB) + "/tab-whatsapp-conversation/" + strconv.Itoa(int(admin.ID)) + "/status/detailed",
			"ConnectUserWAEndpoint":           fun.GLOBAL_URL + "web/" + fun.GetRedis("web:"+admin.Session, redisDB) + "/tab-whatsapp-conversation/" + strconv.Itoa(int(admin.ID)) + "/connect",
			"DisconnectUserWAEndpoint":        fun.GLOBAL_URL + "web/" + fun.GetRedis("web:"+admin.Session, redisDB) + "/tab-whatsapp-conversation/" + strconv.Itoa(int(admin.ID)) + "/disconnect",
			"GetUserWAQREndpoint":             fun.GLOBAL_URL + "web/" + fun.GetRedis("web:"+admin.Session, redisDB) + "/tab-whatsapp-conversation/" + strconv.Itoa(int(admin.ID)) + "/qr",
			"RefreshUserWAQREndpoint":         fun.GLOBAL_URL + "web/" + fun.GetRedis("web:"+admin.Session, redisDB) + "/tab-whatsapp-conversation/" + strconv.Itoa(int(admin.ID)) + "/qr/refresh",
			"CommonWAEndpoint":                fun.GLOBAL_URL + "web/" + fun.GetRedis("web:"+admin.Session, redisDB) + "/tab-whatsapp-conversation/" + strconv.Itoa(int(admin.ID)) + "/",
		}
		c.HTML(200, componentID+".html", replacements)
	}
}
