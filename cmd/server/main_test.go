package main

import "testing"

func TestValidateRuntimeConfigRequiresStrongSecrets(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("SMS_PHONE_HASH_SECRET", "short")

	if err := validateRuntimeConfig("memory", "twilio", ""); err == nil {
		t.Fatal("expected weak phone hash secret to fail")
	}
}

func TestValidateRuntimeConfigRequiresProviderCredentials(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("TWILIO_MESSAGING_SERVICE_SID", "")

	if err := validateRuntimeConfig("memory", "twilio", ""); err == nil {
		t.Fatal("expected missing Twilio credential to fail")
	}
}

func TestValidateRuntimeConfigAcceptsAmazonSNS(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("AWS_REGION", "us-east-1")
	t.Setenv("AWS_ACCESS_KEY_ID", "access-key")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "secret-key")

	if err := validateRuntimeConfig("memory", "amazon_sns", ""); err != nil {
		t.Fatalf("validateRuntimeConfig() error = %v", err)
	}
}

func setBaseEnv(t *testing.T) {
	t.Helper()
	t.Setenv("SMS_OTP_SERVICE_API_TOKEN", "service-token-0123456789abcdef01")
	t.Setenv("SMS_OTP_SECRET", "0123456789abcdef0123456789abcdef")
	t.Setenv("SMS_PHONE_HASH_SECRET", "abcdef0123456789abcdef0123456789")
	t.Setenv("SMS_MFA_SESSION_SECRET", "session-secret-0123456789abcdef01")
	t.Setenv("TWILIO_ACCOUNT_SID", "AC00000000000000000000000000000000")
	t.Setenv("TWILIO_API_KEY_SID", "SK00000000000000000000000000000000")
	t.Setenv("TWILIO_API_KEY_SECRET", "twilio-secret")
	t.Setenv("TWILIO_MESSAGING_SERVICE_SID", "MG00000000000000000000000000000000")
}
