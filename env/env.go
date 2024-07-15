package env

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var spec Specification

type Specification struct {
	TwilioAuth     string
	TwilioAccSid   string
	OpenaiApiKey   string
	JwtSecret      string
	TwilioPhoneNum string
	MongoSecret    string
	TwilioSecret   string
}

func New() *Specification {
	godotenv.Load()

	spec = Specification{
		TwilioAuth:     getEnvVar("TWILIO_AUTH"),
		TwilioAccSid:   getEnvVar("TWILIO_ACC_SID"),
		OpenaiApiKey:   getEnvVar("OPENAI_API_KEY"),
		JwtSecret:      getEnvVar("JWT_SECRET"),
		TwilioPhoneNum: getEnvVar("TWILIO_PHONE_NUM"),
		MongoSecret:    getEnvVar("MONGODB_URI"),
		TwilioSecret:   getEnvVar("TWILIO_SECRET"),
	}
	return &spec
}

func getEnvVar(varName string) string {
	envVar := os.Getenv(varName)
	if envVar == "" {
		log.Panicln(varName, " environment variable is not set.")
	}

	return envVar
}
