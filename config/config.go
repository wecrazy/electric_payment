package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"
)

var (
	config      YamlConfig
	configMutex sync.RWMutex
	configPath  string
)

var yamlFilePaths = []string{
	"/config/conf.%s.yaml",
	"config/conf.%s.yaml",
	"../config/conf.%s.yaml",
	"/../config/conf.%s.yaml",
	"../../config/conf.%s.yaml",
	"/../../config/conf.%s.yaml",
	"C:/golang/electric_payment/config/conf.%s.yaml",
}

var mainConfigPaths = []string{
	"/config/conf.yaml",
	"config/conf.yaml",
	"../config/conf.yaml",
	"/../config/conf.yaml",
	"../../config/conf.yaml",
	"/../../config/conf.yaml",
	"C:/golang/electric_payment/config/conf.yaml",
}

// getEnvironment returns the current environment (dev or prod)
// Priority: 1. CONFIG_MODE from conf.yaml, 2. ENV environment variable, 3. GO_ENV, 4. default to "dev"
func getEnvironment() string {
	// First try to read from main config file
	if mode := getConfigModeFromFile(); mode != "" {
		return mode
	}

	// Check ENV environment variable
	if env := os.Getenv("ENV"); env != "" {
		if env == "dev" || env == "prod" {
			return env
		}
	}

	// Check GO_ENV environment variable
	if env := os.Getenv("GO_ENV"); env != "" {
		if env == "development" {
			return "dev"
		}
		if env == "production" {
			return "prod"
		}
		if env == "dev" || env == "prod" {
			return env
		}
	}

	// Default to development
	return "dev"
}

// getConfigModeFromFile reads the CONFIG_MODE from the main conf.yaml file
func getConfigModeFromFile() string {
	exePath, err := os.Executable()
	if err != nil {
		return ""
	}

	exeDir := filepath.Dir(exePath)

	for _, path := range mainConfigPaths {
		var fullPath string
		if !filepath.IsAbs(path) {
			fullPath = filepath.Join(exeDir, path)
		} else {
			fullPath = path
		}

		if _, err := os.Stat(fullPath); err == nil {
			data, err := os.ReadFile(fullPath)
			if err != nil {
				continue
			}

			var mainConfig MainConfig
			if err := yaml.Unmarshal(data, &mainConfig); err != nil {
				continue
			}

			if mainConfig.ConfigMode == "dev" || mainConfig.ConfigMode == "prod" {
				return mainConfig.ConfigMode
			}
		}
	}

	return ""
}

// getConfigSource returns a string indicating where the config mode was determined from
func getConfigSource() string {
	if getConfigModeFromFile() != "" {
		return "conf.yaml"
	}

	if env := os.Getenv("ENV"); env != "" && (env == "dev" || env == "prod") {
		return "ENV variable"
	}

	if env := os.Getenv("GO_ENV"); env != "" {
		return "GO_ENV variable"
	}

	return "default"
}

// getConfigPaths returns the list of config file paths for the current environment
func getConfigPaths() []string {
	env := getEnvironment()
	paths := make([]string, len(yamlFilePaths))

	for i, path := range yamlFilePaths {
		paths[i] = fmt.Sprintf(path, env)
	}

	return paths
}

// MainConfig represents the main configuration structure for determining mode
type MainConfig struct {
	ConfigMode string `yaml:"CONFIG_MODE"`
}

type Schedule struct {
	Name    string   `yaml:"NAME"`
	Every   string   `yaml:"EVERY,omitempty"`
	At      []string `yaml:"AT,omitempty"`
	Weekly  string   `yaml:"WEEKLY,omitempty"`
	Monthly string   `yaml:"MONTHLY,omitempty"`
	Yearly  string   `yaml:"YEARLY,omitempty"`
}

type WhatsappErrorMessage struct {
	ID idWaErrMsg `yaml:"ID"`
	EN enWaErrMsg `yaml:"EN"`
}

type idWaErrMsg struct {
	PhoneNumberNotRegistered string `yaml:"PHONE_NUMBER_NOT_REGISTERED"`
	PhoneNumberIsBanned      string `yaml:"PHONE_NUMBER_IS_BANNED"`
	InvalidChat              string `yaml:"INVALID_CHAT"`
	MessageTypeDenied        string `yaml:"MESSAGE_TYPE_DENIED"`
	UnknownPrompt            string `yaml:"UNKNOWN_PROMPT"`
	AccountBannedCozBadWord  string `yaml:"ACCOUNT_BANNED_COZ_BAD_WORD"`
}

type enWaErrMsg struct {
	PhoneNumberNotRegistered string `yaml:"PHONE_NUMBER_NOT_REGISTERED"`
	PhoneNumberIsBanned      string `yaml:"PHONE_NUMBER_IS_BANNED"`
	InvalidChat              string `yaml:"INVALID_CHAT"`
	MessageTypeDenied        string `yaml:"MESSAGE_TYPE_DENIED"`
	UnknownPrompt            string `yaml:"UNKNOWN_PROMPT"`
	AccountBannedCozBadWord  string `yaml:"ACCOUNT_BANNED_COZ_BAD_WORD"`
}

type WhatsappModel struct {
	TBConversation          string `yaml:"TB_WA_CONVERSATION"`
	TBChatMessage           string `yaml:"TB_WA_CHAT_MESSAGE"`
	TBGroupParticipant      string `yaml:"TB_WA_GROUP_PARTICIPANT"`
	TBContactInfo           string `yaml:"TB_WA_CONTACT_INFO"`
	TBMessageDeliveryStatus string `yaml:"TB_WA_MESSAGE_DELIVERY_STATUS"`
	TBMediaFile             string `yaml:"TB_WA_MEDIA_FILE"`
}

type YamlConfig struct {
	App struct {
		Host                 string `yaml:"HOST"`
		GinMode              string `yaml:"GIN_MODE"`
		Name                 string `yaml:"NAME"`
		Logo                 string `yaml:"LOGO"`
		Port                 string `yaml:"PORT"`
		LogLevel             string `yaml:"LOG_LEVEL"`
		LogFormat            string `yaml:"LOG_FORMAT"`
		WebPublicURL         string `yaml:"WEB_PUBLIC_URL"`
		Version              string `yaml:"VERSION"`
		VersionNo            int    `yaml:"VERSION_NO"`
		VersionCode          string `yaml:"VERSION_CODE"`
		VersionName          string `yaml:"VERSION_NAME"`
		StaticDir            string `yaml:"STATIC_DIR"`
		PublishedDir         string `yaml:"PUBLISHED_DIR"`
		LogDir               string `yaml:"LOG_DIR"`
		UploadDir            string `yaml:"UPLOAD_DIR"`
		LoginTimeM           string `yaml:"LOGIN_TIME_M"`
		CookieLoginDomain    string `yaml:"COOKIE_LOGIN_DOMAIN"`
		CookieLoginSecure    bool   `yaml:"COOKIE_LOGIN_SECURE"`
		MaxDisconnectionTime string `yaml:"MAX_DISCONNECTION_TIME"`
		AesKey               string `yaml:"AES_KEY"`
		AesKeyIV             string `yaml:"AES_KEY_IV"`
		MaxRetryLogin        int    `yaml:"MAX_RETRY_LOGIN"`
		LoginLockUntil       int    `yaml:"LOGIN_LOCK_UNTIL"`
		AppLogFilename       string `yaml:"APP_LOG_FILENAME"`
		SystemLogFilename    string `yaml:"SYSTEM_LOG_FILENAME"`
	} `yaml:"APP"`

	Default struct {
		TimeZone              string `yaml:"TIMEZONE"`
		MinLengthPhoneNumber  int    `yaml:"MIN_LENGTH_PHONE_NUMBER"`
		MaxMessageCharacters  int    `yaml:"MAX_MESSAGE_CHAR"`
		MaxImageSize          int64  `yaml:"MAX_IMAGE_SIZE"`
		XAMPPMySQLPath        string `yaml:"XAMPP_MYSQL_PATH"`
		NssmFullPath          string `yaml:"NSSM_FULLPATH"`
		MagickFullPath        string `yaml:"MAGICK_FULLPATH"`
		ConcurrencyLimit      int    `yaml:"CONCURRENCY_LIMIT"`
		APIKeyApiAnalyticsDev string `yaml:"API_KEY_API_ANALYTICS_DEV"`
		WelcomeID             string `yaml:"WELCOME_ID"`
		WelcomeEN             string `yaml:"WELCOME_EN"`
	} `yaml:"DEFAULT"`

	Email struct {
		Host              string `yaml:"HOST"`
		Port              int    `yaml:"PORT"`
		Username          string `yaml:"USERNAME"`
		Password          string `yaml:"PASSWORD"`
		Sender            string `yaml:"SENDER"`
		MaxRetry          int    `yaml:"MAX_RETRY"`
		RetryDelay        int    `yaml:"RETRY_DELAY"`
		MaxAttachmentSize int64  `yaml:"MAX_ATTACHMENT_SIZE"`
	} `yaml:"EMAIL"`

	Redis struct {
		Host       string `yaml:"HOST"`
		Port       int    `yaml:"PORT"`
		Password   string `yaml:"PASSWORD"`
		Db         int    `yaml:"DB"`
		MaxRetry   int    `yaml:"MAX_RETRY"`
		RetryDelay int    `yaml:"RETRY_DELAY"`
	} `yaml:"REDIS"`

	Database struct {
		Type                   string `yaml:"TYPE"`
		Host                   string `yaml:"HOST"`
		Port                   string `yaml:"PORT"`
		Username               string `yaml:"USERNAME"`
		Password               string `yaml:"PASSWORD"`
		Name                   string `yaml:"NAME"`
		MaxRetryConnect        int    `yaml:"MAX_RETRY_CONNECT"`
		RetryDelay             int    `yaml:"RETRY_DELAY"`
		MaxIdleConns           int    `yaml:"MAX_IDLE_CONNECTION"`
		MaxOpenConns           int    `yaml:"MAX_OPEN_CONNECTION"`
		MaxLifetimeConns       int    `yaml:"MAX_LIFETIME_CONNECTION"`
		DBConfigPath           string `yaml:"DB_CONFIG_PATH"`
		DBBackupDestinationDir string `yaml:"DB_BACKUP_DESTINATION_DIR"`

		// MAIN TABLE
		TbAdmin                  string `yaml:"TB_ADMIN"`
		TbAdminPwdChangelog      string `yaml:"TB_ADMIN_PWD_CHANGELOG"`
		TbAdminStatus            string `yaml:"TB_ADMIN_STATUS"`
		TbFeature                string `yaml:"TB_FEATURE"`
		TbLogActivity            string `yaml:"TB_LOG_ACTIVITY"`
		TbRole                   string `yaml:"TB_ROLE"`
		TbRolePrivilege          string `yaml:"TB_ROLE_PRIVILEGE"`
		TbLanguage               string `yaml:"TB_LANGUAGE"`
		TbBadWord                string `yaml:"TB_BAD_WORD"`
		TbWebAppConfig           string `yaml:"TB_WEB_APP_CONFIG"`
		TbIndonesiaRegion        string `yaml:"TB_INDONESIA_REGION"`
		DumpedIndonesiaRegionSQL string `yaml:"DUMPED_INDONESIA_REGION_SQL"`

		TbWAMsg       string `yaml:"TB_WA_MSG"`
		TbWAMsgReply  string `yaml:"TB_WA_MSG_REPLY"`
		TbWAPhoneUser string `yaml:"TB_WA_PHONE_USER"`
	} `yaml:"DATABASE"`

	Whatsmeow struct {
		SqlDriver                 string               `yaml:"SQL_DRIVER"`
		QrCode                    string               `yaml:"QR_CODE"`
		QrExpired                 int                  `yaml:"QR_EXPIRED"`
		SqlSource                 string               `yaml:"SQL_SOURCE"`
		WaGroupSource             string               `yaml:"WA_GROUP_SOURCE"`
		WaSuperUser               string               `yaml:"WA_SU"`
		WaSupport                 string               `yaml:"WA_SUPPORT"`
		WaTechnicalSupport        string               `yaml:"WA_TECHNICAL_SUPPORT"`
		WaBotUsed                 []string             `yaml:"WA_BOT_USED"`
		WaGroupAllowedToUsePrompt []string             `yaml:"WAG_ALLOWED_PROMPT"`
		WAGTestJID                string               `yaml:"WAG_TEST_JID"`
		WAGTAJID                  string               `yaml:"WAG_TA_JID"`
		WAGRegionTechnician       map[int]string       `yaml:"WAG_REGION_TECHNICIAN"`
		InitLanguagePrompt        string               `yaml:"INITIAL_LANGUAGE_PROMPT"`
		WAReplyPublicURL          string               `yaml:"WA_REPLY_PUBLIC_URL"`
		OllamaURL                 string               `yaml:"OLLAMA_URL"`
		OllamaModel               string               `yaml:"OLLAMA_MODEL"`
		KeywordSeparator          string               `yaml:"KEYWORD_SEPARATOR"`
		MsgReceivedLogFile        string               `yaml:"MESSAGE_RECEIVED_LOG_FILE"`
		WhatsmeowClientLog        string               `yaml:"WHATSMEOW_CLIENT_LOG"`
		WhatsmeowClientLogLevel   string               `yaml:"WHATSMEOW_CLIENT_LOG_LEVEL"`
		WhatsmeowDBLog            string               `yaml:"WHATSMEOW_DB_LOG"`
		WhatsmeowDBLogLevel       string               `yaml:"WHATSMEOW_DB_LOG_LEVEL"`
		OpenWeatherMapAPIKey      string               `yaml:"OPEN_WEATHER_MAP_API"`
		WhatsappMaxDailyQuota     int                  `yaml:"WHATSAPP_MAX_DAILY_QUOTA"`
		WhatsappMaxBadWordStrike  int                  `yaml:"WHATSAPP_MAX_BAD_WORD_STRIKE"`
		RedisExpiry               int                  `yaml:"REDIS_EXPIRY"`
		MaxUploadedDocumentSize   int64                `yaml:"MAX_UPLOADED_DOCUMENT_SIZE"`
		MaxUploadedImageSize      int64                `yaml:"MAX_UPLOADED_IMAGE_SIZE"`
		MaxUploadedAudioSize      int64                `yaml:"MAX_UPLOADED_AUDIO_SIZE"`
		MaxUploadedVideoSize      int64                `yaml:"MAX_UPLOADED_VIDEO_SIZE"`
		DocumentAllowedExtensions []string             `yaml:"DOCUMENT_ALLOWED_EXTENSIONS"`
		DocumentAllowedMimeTypes  []string             `yaml:"DOCUMENT_ALLOWED_MIME_TYPES"`
		ImageAllowedExtensions    []string             `yaml:"IMAGE_ALLOWED_EXTENSIONS"`
		ImageAllowedMimeTypes     []string             `yaml:"IMAGE_ALLOWED_MIME_TYPES"`
		AudioAllowedExtensions    []string             `yaml:"AUDIO_ALLOWED_EXTENSIONS"`
		AudioAllowedMimeTypes     []string             `yaml:"AUDIO_ALLOWED_MIME_TYPES"`
		VideoAllowedExtensions    []string             `yaml:"VIDEO_ALLOWED_EXTENSIONS"`
		VideoAllowedMimeTypes     []string             `yaml:"VIDEO_ALLOWED_MIME_TYPES"`
		WaErrorMessage            WhatsappErrorMessage `yaml:"WA_ERROR_MESSAGE"`
		WelcomingUserID           string               `yaml:"WELCOMING_USER_ID"`
		WelcomingUserEN           string               `yaml:"WELCOMING_USER_EN"`
		WhatsappModel             WhatsappModel        `yaml:"WHATSAPP_MODEL"`
	} `yaml:"WHATSMEOW"`

	Schedules []Schedule `yaml:"SCHEDULES"`

	FolderFileNeeds []string `yaml:"FOLDER_FILE_NEEDS"`

	API struct {
		IndonesianPublicHoliday string `yaml:"INDONESIAN_PUBLIC_HOLIDAY"`
	} `yaml:"API"`

	PLTMHLembangPalesan struct {
		Email                string   `yaml:"EMAIL"`
		Phone                string   `yaml:"PHONE"`
		PublicURL            string   `yaml:"PUBLIC_URL"`
		Facebook             string   `yaml:"FACEBOOK"`
		Instagram            string   `yaml:"INSTAGRAM"`
		Youtube              string   `yaml:"YOUTUBE"`
		Twitter              string   `yaml:"TWITTER"`
		TbElectricityBase    string   `yaml:"TB_ELECTRICITY_BASE"`
		TbPrepaid            string   `yaml:"TB_PREPAID"`
		TbPostpaid           string   `yaml:"TB_POSTPAID"`
		TbTopupRecord        string   `yaml:"TB_TOPUP_RECORD"`
		TbTransactionHistory string   `yaml:"TB_TRANSACTION_HISTORY"`
		TopupOptions         []int    `yaml:"TOPUP_OPTIONS"`
		AdminFee             int64    `yaml:"ADMIN_FEE"`
		TarifCode            []string `yaml:"TARIF_CODE"`
	} `yaml:"PLTMH_LEMBANG_PALESAN"`
}

func YAMLLoad(filePath string) (*YamlConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config YamlConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func InitConfig() (*YamlConfig, error) {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("Error getting executable path: %v", err)
	}

	exeDir := filepath.Dir(exePath)
	configPaths := getConfigPaths()

	for i, path := range configPaths {
		if !filepath.IsAbs(path) {
			configPaths[i] = filepath.Join(exeDir, path)
		}
	}

	var yamlConfig *YamlConfig

	for _, filePath := range configPaths {
		if _, err := os.Stat(filePath); err == nil {
			yamlConfig, err = YAMLLoad(filePath)
			if err != nil {
				log.Printf("failed to load configuration from '%s': %v", filePath, err)
				continue
			}
			log.Printf("Configuration successfully loaded from '%s' (Environment: %s from %s)", filePath, getEnvironment(), getConfigSource())
			break
		} else if os.IsNotExist(err) {
			log.Printf("configuration file '%s' does not exist. Skipping.", filePath)
		} else {
			log.Printf("error checking file '%s': %v", filePath, err)
		}
	}

	if yamlConfig == nil {
		log.Fatalf("failed to load YAML configuration: no valid configuration file found in paths: %v", configPaths)
	}

	return yamlConfig, nil
}

func LoadConfig() error {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("Error getting executable path: %v", err)
	}

	exeDir := filepath.Dir(exePath)
	configPaths := getConfigPaths()

	for i, path := range configPaths {
		if !filepath.IsAbs(path) {
			configPaths[i] = filepath.Join(exeDir, path)
		}
	}

	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			// log.Printf("Config file found: %v (Environment: %s from %s)", path, getEnvironment(), getConfigSource())
			configPath = path
			break
		}
	}
	if configPath == "" {
		return fmt.Errorf("no valid config file found from paths: %v", configPaths)
	}

	file, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var newConfig YamlConfig
	if err := yaml.Unmarshal(file, &newConfig); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	configMutex.Lock()
	config = newConfig
	configMutex.Unlock()

	return nil
}

func WatchConfig() {
	if configPath == "" {
		log.Println("no valid config file found. Skipping watcher.")
		return
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("failed to initialize config watcher:%v", err)
	}
	defer watcher.Close()

	err = watcher.Add(configPath)
	if err != nil {
		log.Printf("failed to watch config file:%v", err)
	}

	fmt.Println("👀 Watching for yaml config changes:", configPath)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op == fsnotify.Write {
				fmt.Println("config file updated. Reloading...")
				if err := LoadConfig(); err != nil {
					log.Printf("failed to reload config:%v", err)
				} else {
					fmt.Println("config reloaded successfully.")
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("config watcher error:", err)
		}
	}
}

func GetConfig() YamlConfig {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return config
}
