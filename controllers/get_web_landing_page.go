package controllers

import (
	"electric_payment/config"
	"electric_payment/fun"
	"electric_payment/internal/gormdb"
	"electric_payment/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetWebLandingPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := gormdb.Databases.Web

		// Retrieve cookies from the request
		cookies := c.Request.Cookies()

		// Check if the "credentials" cookie exists
		var credentialsCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == "credentials" {
				credentialsCookie = cookie
				break
			}
		}
		parameters := gin.H{
			"APP_NAME":         config.GetConfig().App.Name,
			"APP_LOGO":         config.GetConfig().App.Logo,
			"APP_VERSION":      config.GetConfig().App.Version,
			"APP_VERSION_NO":   config.GetConfig().App.VersionNo,
			"APP_VERSION_CODE": config.GetConfig().App.VersionCode,
			"APP_VERSION_NAME": config.GetConfig().App.VersionName,
			"APP_TITLE":        config.GetConfig().App.Name,
			"GLOBAL_URL":       fun.GLOBAL_URL,
			"LOGIN":            "LOGIN",
			"PLTMH_EMAIL":      config.GetConfig().PLTMHLembangPalesan.Email,
			"PLTMH_PHONE":      config.GetConfig().PLTMHLembangPalesan.Phone,
			"PLTMH_URL":        config.GetConfig().PLTMHLembangPalesan.PublicURL,
			"PLTMH_FACEBOOK":   config.GetConfig().PLTMHLembangPalesan.Facebook,
			"PLTMH_INSTAGRAM":  config.GetConfig().PLTMHLembangPalesan.Instagram,
			"PLTMH_YOUTUBE":    config.GetConfig().PLTMHLembangPalesan.Youtube,
			"PLTMH_TWITTER":    config.GetConfig().PLTMHLembangPalesan.Twitter,
		}

		if credentialsCookie != nil {
			var admin model.Admin
			if err := db.Where("session = ?", credentialsCookie.Value).First(&admin).Error; err != nil {
				fun.ClearCookiesAndRedirect(c, cookies)
			}

		}
		c.HTML(http.StatusOK, "landing-page.html", parameters)
	}
}
