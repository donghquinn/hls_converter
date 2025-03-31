package configs

import "os"

type GlobalConfig struct {
	AppPort      string
	AppHost      string
	AesKey       string
	AesIv        string
	JwtKey       string
	CookieDomain string

	GoogleClientId string
	GooglePass     string

	UploadDir string
	OutputDir string
}

var GlobalConfiguration GlobalConfig

func SetGlobalConfiguration() {
	GlobalConfiguration.AppPort = os.Getenv("APP_PORT")
	GlobalConfiguration.AppHost = os.Getenv("APP_HOST")

	// AES
	GlobalConfiguration.AesKey = os.Getenv("AES_KEY")
	GlobalConfiguration.AesIv = os.Getenv("AES_IV")

	// JWT
	GlobalConfiguration.JwtKey = os.Getenv("JWT_KEY")
	// GlobalConfiguration.CookieDomain = os.Getenv("COOKIE_DOMAIN")

	// Google Oauth
	// GlobalConfiguration.GoogleClientId = os.Getenv("GOOGLE_CLIENT_ID")
	// GlobalConfiguration.GooglePass = os.Getenv("GOOGLE_CLIENT_PASSWD")

	GlobalConfiguration.UploadDir = os.Getenv("UPLOAD_DIR")
	GlobalConfiguration.OutputDir = os.Getenv("OUTPUT_DIR")
}
