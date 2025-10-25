package routes

import (
	"electric_payment/config"
	"electric_payment/controllers"
	"electric_payment/fun"
	"electric_payment/internal/gormdb"
	"electric_payment/middleware"
	"electric_payment/model"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	analytics "github.com/tom-draper/api-analytics/analytics/go/gin"
)

func StaticFile(router *gin.Engine) {
	staticPath := config.GetConfig().App.StaticDir
	publishedDir := config.GetConfig().App.PublishedDir

	// Resolve static path to absolute
	staticPath, err := filepath.Abs(staticPath)
	if err != nil {
		fmt.Println("Error getting absolute path:", err)
		return
	}

	// Load global HTML templates
	router.LoadHTMLGlob(filepath.Join(staticPath, "**", "*.html"))

	// Serve static directories
	if publishedDir != "" {
		var directories []string

		// Support multiple directories
		if strings.Contains(publishedDir, "|") {
			directories = strings.Split(publishedDir, "|")
		} else {
			directories = append(directories, publishedDir)
		}

		for _, dir := range directories {
			// Skip entries with '#' (optional: handle as comment/ignore marker)
			if strings.Contains(dir, "#") {
				continue
			}

			// Clean relative path
			cleanDir := filepath.Clean(dir)

			// Combine with static root
			staticDirPath := filepath.Join(staticPath, cleanDir)

			// Check if it exists
			if _, err := os.Stat(staticDirPath); os.IsNotExist(err) {
				fmt.Println("Directory does not exist:", staticDirPath)
				continue
			}

			// Serve static files under constructed URL
			urlPath := path.Join(fun.GLOBAL_URL, cleanDir)
			router.Static(urlPath, staticDirPath)

			fmt.Println("📂 Published static dir:", staticDirPath, "at", urlPath)
		}
	}

	router.Static("./uploads", "uploads")
	router.Static("./log-data", "log")
	router.Static("./wa", "whatsmeow")
	router.Static("./wa_reply", "web/file/wa_reply")
	router.Static("/media", "./web/assets/whatsapp_media") // Serve WhatsApp media files
}

func HtmlRoutes(router *gin.Engine, redisDB *redis.Client) {
	db := gormdb.Databases.Web

	// To view the dashboard API analytics go to: https://www.apianalytics.dev/dashboard and enter your API key
	router.Use(analytics.Analytics(config.GetConfig().Default.APIKeyApiAnalyticsDev)) // Add middleware

	router.GET(fun.GLOBAL_URL+"hello", func(ctx *gin.Context) {
		data := map[string]string{
			"message": "Hello, World!",
		}
		ctx.JSON(http.StatusOK, data)
	})

	router.GET(fun.GLOBAL_URL+"ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong!"})
	})

	router.GET(fun.GLOBAL_URL+"ws", controllers.WebSocketVerify(db))

	// router.GET(fun.GLOBAL_URL+"", controllers.GetWebLandingPage(db)) // LANDING PAGE
	// TODO: add a proper landing page
	router.GET(fun.GLOBAL_URL, func(c *gin.Context) { c.Redirect(http.StatusPermanentRedirect, fun.GLOBAL_URL+"login") })

	router.GET(fun.GLOBAL_URL+"login", controllers.GetWebLogin(db))            // WEB LOGIN
	router.POST(fun.GLOBAL_URL+"login", controllers.PostWebLogin(db, redisDB)) // SEND LOGIN CREDENTIALS
	router.GET(fun.GLOBAL_URL+"captcha", controllers.GetCaptchaImage())

	router.GET(fun.GLOBAL_URL+"forgot-password", controllers.GetWebForgotPassword(db))
	router.POST(fun.GLOBAL_URL+"forgot-password", controllers.PostForgotPassword(db, redisDB))
	router.GET(fun.GLOBAL_URL+"reset-password/:email/:token_data", controllers.GetWebResetPassword(db, redisDB))
	router.POST(fun.GLOBAL_URL+"reset-password/:email/:token_data", controllers.PostResetPassword(db, redisDB))

	//MAIN PAGE
	router.GET(fun.GLOBAL_URL+"page", controllers.MainPage(db, redisDB))

	// LOGOUT BY BUTTON
	router.GET(fun.GLOBAL_URL+"logout", controllers.GetWebLogout(db))
	// router.GET(fun.GLOBAL_URL+"register", controllers.getRegister(db))

	router.GET(fun.GLOBAL_URL+"profile/default.jpg", controllers.GetUserProfile(db))

	// Check WhatsApp number registration
	router.GET(fun.GLOBAL_URL+"check_wa", controllers.CheckWAPhoneNumberIsRegistered())

	// Landing Page
	router.GET(fun.GLOBAL_URL+"welcome", controllers.GetWebLandingPage()) // LANDING PAGE
	pltmhLembangPalesanPayment := router.Group("/payment")
	{
		pltmhLembangPalesanPayment.GET("", controllers.GetPaymentPage())
	}

	// API routes for payment system
	apiRoutes := router.Group(fun.GLOBAL_URL + "api")
	{
		apiRoutes.GET("/ping", func(c *gin.Context) {
			i := c.Query("i")
			if i != "" {
				c.JSON(http.StatusOK, gin.H{"message": "pong", "i": i})
			} else {
				c.JSON(http.StatusOK, gin.H{"message": "pong"})
			}
		})
		apiRoutes.POST("/check-customer", controllers.CheckCustomerData())
		apiRoutes.POST("/process-payment", controllers.ProcessPayment())
	}

	// Endpoint Web routes group
	web := router.Group(fun.GLOBAL_URL+"web/:access", middleware.AuthMiddleware(db, redisDB))
	{

		//GUI PAGE COMPONENT
		web.GET("/components/:component", controllers.ComponentPage(db, redisDB))

		// Handle dynamic folder structure
		web.GET("/uploads/:year/:month/:day/:filename", func(c *gin.Context) {
			// Extract parameters from the route
			year := c.Param("year")
			month := c.Param("month")
			day := c.Param("day")
			filename := c.Param("filename")

			// Construct the file path
			filePath := filepath.Join("./uploads", year, month, day, filename)

			// Clean the file path to prevent directory traversal
			safePath := filepath.Clean(filePath)

			// Ensure the safePath is within the uploads directory
			if !filepath.HasPrefix(safePath, filepath.Clean("./uploads")) {
				c.JSON(http.StatusForbidden, gin.H{"error": "invalid file path"})
				return
			}

			// Serve the file
			c.File(safePath)
		})

		/* Dashboard */
		tabDashboard := web.Group("/tab-dashboard")
		{
			tabDashboard.GET("/pie_chart_total", func(c *gin.Context) {
				// Example: just return a 204 No Content with no body
				// ADD some graphics
				c.Status(http.StatusNoContent)
			})
		}

		/* Tab App Config */
		tabAppConfig := web.Group("/tab-app-config")
		{
			tabAppConfig.POST("/table", controllers.TableAppConfig())
		}

		/*
			Tab Whatsapp
		*/
		tabWhatsapp := web.Group("/tab-whatsapp")
		{
			tabWhatsapp.GET("/refresh-qrcode", controllers.RefreshWhatsappQrcode())
			tabWhatsapp.GET("/end-session", controllers.EndSessionWhatsapp())
			tabWhatsapp.POST("/ping", controllers.PingWhatsapp())
			tabWhatsapp.POST("/send_text", controllers.SendTextWhatsapp())
			tabWhatsapp.POST("/send_image", controllers.SendImageWhatsapp())
			tabWhatsapp.POST("/send_document", controllers.SendDocumentWhatsapp())
			tabWhatsapp.POST("/send_location", controllers.SendLocationWhatsapp())
			tabWhatsapp.POST("/send_polling", controllers.SendPollingWhatsapp())
			tabWhatsapp.GET("/groups", controllers.GetWhatsappGroups())
			tabWhatsapp.POST("/wa_log_msg_received", controllers.GetTbLogMsgReceived())
			// Language
			tabWhatsapp.POST("/table_language", controllers.TableWhatsappBotLanguage())
			tabWhatsapp.PUT("/table_language", controllers.PutDataWhatsappBotLanguage())
			tabWhatsapp.DELETE("/table_language/:id", controllers.DeleteDataWhatsappBotLanguage())
			tabWhatsapp.GET("/table_language.csv", controllers.ExportTable[model.Language](db, "File di unggah"))
			tabWhatsapp.POST("/table_language/create", controllers.PostNewWhatsappBotLanguage())
			tabWhatsapp.GET("/last_update_table_language", controllers.LastUpdateTableWhatsappBotLanguage())
			// Message Reply
			tabWhatsapp.POST("/table_message_reply", controllers.TableWhatsappBotMessageReply())
			tabWhatsapp.PUT("/table_message_reply", controllers.PutDataWhatsappBotMessageReply())
			tabWhatsapp.DELETE("/table_message_reply/:id", controllers.DeleteDataWhatsappBotMessageReply())
			tabWhatsapp.GET("/table_message_reply.csv", controllers.ExportTable[model.WAMessageReply](db, "File di unggah"))
			tabWhatsapp.POST("/table_message_reply/create", controllers.PostNewWhatsappBotMessageReply())
			tabWhatsapp.GET("/table_message_reply/batch/template", controllers.GetBatchTemplateWhatsappBotMessageReply[model.WAMessageReply]())
			tabWhatsapp.POST("/table_message_reply/batch/create", controllers.PostBatchUploadDataWhatsappBotMessageReply[model.WAMessageReply]())
			tabWhatsapp.GET("/last_update_table_message_reply", controllers.LastUpdateTableWhatsappBotMessageReply())
		}

		/*
			Tab Whatsapp User Management
		*/
		tabWaUserManagement := web.Group("/tab-whatsapp-user-management")
		{
			tabWaUserManagement.POST("/table", controllers.TableWhatsappUserManagement())
			tabWaUserManagement.PUT("/table", controllers.PutUpdatedWhatsappUserManagement())
			tabWaUserManagement.DELETE("/table/:id", controllers.DeleteDataFromTableWhatsappUserManagement())
			tabWaUserManagement.POST("/table/create", controllers.CreateNewDataTableWhatsappUserManagement())
			tabWaUserManagement.GET("/table/batch/template", controllers.GetBatchTemplateWhatsappUserManagement[model.WAPhoneUser]())
			tabWaUserManagement.POST("/table/batch/create", controllers.PostBatchUploadDataWhatsappUserManagement[model.WAPhoneUser]())
			tabWaUserManagement.POST("/reset_quota_prompt", controllers.ResetQuotaWhatsappPrompt())
			tabWaUserManagement.POST("/unban_user", controllers.UnbanUser())
		}

		/*
			Tab Whatsapp Conversation
		*/
		tabWhatsappConversation := web.Group("/tab-whatsapp-conversation")
		{
			// Check if user's WhatsApp is logged in (legacy - returns only logged_in boolean)
			tabWhatsappConversation.GET("/:userid/status", controllers.IsUserWhatsappLoggedIn(db, redisDB))
			// Get detailed status information for user's WhatsApp client
			tabWhatsappConversation.GET("/:userid/status/detailed", controllers.GetUserWhatsappStatus(db, redisDB))
			// Connect user's WhatsApp client
			tabWhatsappConversation.POST("/:userid/connect", controllers.ConnectUserWhatsapp(db, redisDB))
			// Disconnect user's WhatsApp client
			tabWhatsappConversation.POST("/:userid/disconnect", controllers.DisconnectUserWhatsapp(db, redisDB))
			// Get QR code for user's WhatsApp client
			tabWhatsappConversation.GET("/:userid/qr", controllers.GetUserWhatsappQR(db, redisDB))
			// Refresh QR code for user's WhatsApp client
			tabWhatsappConversation.POST("/:userid/qr/refresh", controllers.RefreshUserWhatsappQR(db, redisDB))
			// Send message using user's WhatsApp client
			tabWhatsappConversation.POST("/:userid/send", controllers.SendUserWhatsappMessage(db, redisDB))
			// WhatsApp Interface Components
			tabWhatsappConversation.GET("/:userid/sidebar-left", controllers.GetWhatsappSidebarLeft(db, redisDB))
			tabWhatsappConversation.GET("/:userid/chat-area", controllers.GetWhatsappChatArea(db, redisDB))
			tabWhatsappConversation.GET("/:userid/contact-list", controllers.GetWhatsappContactList(db, redisDB))
			tabWhatsappConversation.GET("/:userid/conversation-history", controllers.GetWhatsappConversationHistory(db, redisDB))
			// Search Functions
			tabWhatsappConversation.POST("/:userid/search/contacts", controllers.SearchWhatsappContacts(db, redisDB))
			tabWhatsappConversation.POST("/:userid/search/messages", controllers.SearchWhatsappMessages(db, redisDB))
			tabWhatsappConversation.POST("/:userid/search/conversations", controllers.SearchWhatsappConversations(db, redisDB))
			// List all active clients (admin endpoint)
			tabWhatsappConversation.GET("/active", controllers.ListActiveWhatsappClients())

		}

		tabRoles := web.Group("/tab-roles")
		{
			// /web/tab-roles/admin/status
			tabRoles.GET("/roles/gui", controllers.GetRolesGui(db))

			tabRoles.GET("/roles/modal", controllers.ModalTabRoles(db))

			tabRoles.POST("/roles/create", controllers.PostRole(db))
			tabRoles.PATCH("/roles", controllers.PatchRole(db))
			tabRoles.DELETE("/roles", controllers.DeleteRoles(db))

			tabRoles.GET("/roles/list", controllers.GetRolesList(db))

			tabRoles.GET("/admins/table", controllers.GetAdminTable(db))
			tabRoles.POST("/admins/create", controllers.PostNewAdminUser(db))
			tabRoles.PATCH("/admins", controllers.PatchAdminData(db))
			tabRoles.DELETE("/admins/:id", controllers.DeleteUserAdmin(db))
		}

		tabSystemLog := web.Group("/tab-system-log")
		{
			tabSystemLog.GET("/system/log/file", controllers.GetSystemLogFiles(db))
			tabSystemLog.GET("/table", controllers.GetSystemLog(db))
			tabSystemLog.GET("/table.csv", controllers.GetSystemLogFileDump(db))
		}

		tabActivityLog := web.Group("/tab-activity-log")
		{ // /web/tab-activity-log/activity/log
			tabActivityLog.GET("/table", controllers.GetActivityLog(db))
			tabActivityLog.GET("/table.csv", controllers.DumpActivityLog(db))
		}
		tabUserProfile := web.Group("/tab-user-profile")
		{
			tabUserProfile.GET("/activity/table", controllers.TableUserActivities(db))
			tabUserProfile.PATCH("/profile-image", controllers.UpdateAdminProfileImage(db))
		}
	}
}
