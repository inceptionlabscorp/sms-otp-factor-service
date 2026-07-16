package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/inceptionlabscorp/sms-otp-factor-service/internal/adapters/httpapi"
	"github.com/inceptionlabscorp/sms-otp-factor-service/internal/adapters/sms/sns"
	"github.com/inceptionlabscorp/sms-otp-factor-service/internal/adapters/sms/twilio"
	"github.com/inceptionlabscorp/sms-otp-factor-service/internal/adapters/store"
	app "github.com/inceptionlabscorp/sms-otp-factor-service/internal/application/smsotp"
)

func main() {
	port := env("PORT", "8080")
	projectID := firstNonEmpty(os.Getenv("GCP_PROJECT_ID"), os.Getenv("GOOGLE_CLOUD_PROJECT"))
	storeDriver := env("STORE_DRIVER", "firestore")
	smsProvider := strings.ToLower(env("SMS_PROVIDER", "twilio"))

	if err := validateRuntimeConfig(storeDriver, smsProvider, projectID); err != nil {
		log.Fatal(err)
	}

	var challengeStore app.ChallengeRepository
	switch storeDriver {
	case "memory":
		challengeStore = store.NewMemoryStore()
	default:
		if strings.TrimSpace(projectID) == "" {
			log.Fatal("GCP_PROJECT_ID or GOOGLE_CLOUD_PROJECT is required for firestore store")
		}
		challengeStore = &store.FirestoreRESTStore{
			ProjectID:  projectID,
			Collection: env("FIRESTORE_COLLECTION", "sms_otp_challenges"),
		}
	}

	smsGateway := smsGatewayFromEnv()
	otpService := app.Service{
		Challenges:      challengeStore,
		SMS:             smsGateway,
		Generator:       app.CryptoCodeGenerator{},
		OTPSecret:       os.Getenv("SMS_OTP_SECRET"),
		PhoneHashSecret: os.Getenv("SMS_PHONE_HASH_SECRET"),
		MessageTemplate: os.Getenv("OTP_MESSAGE_TEMPLATE"),
	}

	handler := httpapi.Handler{
		OTP: otpService,
		Session: app.SessionService{
			Secret: os.Getenv("SMS_MFA_SESSION_SECRET"),
		},
		ServiceToken: os.Getenv("SMS_OTP_SERVICE_API_TOKEN"),
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           httpapi.TimeoutMiddleware(handler, 20*time.Second),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("sms-otp-factor-service listening on :%s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func env(key string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func smsGatewayFromEnv() app.SMSGateway {
	switch strings.ToLower(env("SMS_PROVIDER", "twilio")) {
	case "amazon_sns", "sns", "amazon-simple-notification-service":
		return sns.Client{
			Region:          os.Getenv("AWS_REGION"),
			AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
			SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
			SessionToken:    os.Getenv("AWS_SESSION_TOKEN"),
			SMSType:         os.Getenv("AWS_SNS_SMS_TYPE"),
			SenderID:        os.Getenv("AWS_SNS_SENDER_ID"),
		}
	default:
		return twilio.Client{
			AccountSID:       os.Getenv("TWILIO_ACCOUNT_SID"),
			APIKeySID:        os.Getenv("TWILIO_API_KEY_SID"),
			APIKeySecret:     os.Getenv("TWILIO_API_KEY_SECRET"),
			MessagingService: os.Getenv("TWILIO_MESSAGING_SERVICE_SID"),
		}
	}
}

func validateRuntimeConfig(storeDriver string, smsProvider string, projectID string) error {
	if storeDriver != "memory" && strings.TrimSpace(projectID) == "" {
		return fatalConfig("GCP_PROJECT_ID or GOOGLE_CLOUD_PROJECT is required for firestore store")
	}
	requiredStrongSecrets := []string{
		"SMS_OTP_SERVICE_API_TOKEN",
		"SMS_OTP_SECRET",
		"SMS_PHONE_HASH_SECRET",
		"SMS_MFA_SESSION_SECRET",
	}
	for _, key := range requiredStrongSecrets {
		if len(strings.TrimSpace(os.Getenv(key))) < 32 {
			return fatalConfig(key + " must be at least 32 characters")
		}
	}
	switch smsProvider {
	case "amazon_sns", "sns", "amazon-simple-notification-service":
		for _, key := range []string{"AWS_REGION", "AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"} {
			if strings.TrimSpace(os.Getenv(key)) == "" {
				return fatalConfig(key + " is required for amazon_sns")
			}
		}
	case "twilio":
		for _, key := range []string{"TWILIO_ACCOUNT_SID", "TWILIO_API_KEY_SID", "TWILIO_API_KEY_SECRET", "TWILIO_MESSAGING_SERVICE_SID"} {
			if strings.TrimSpace(os.Getenv(key)) == "" {
				return fatalConfig(key + " is required for twilio")
			}
		}
	default:
		return fatalConfig("SMS_PROVIDER must be twilio or amazon_sns")
	}
	return nil
}

type fatalConfig string

func (e fatalConfig) Error() string { return string(e) }
