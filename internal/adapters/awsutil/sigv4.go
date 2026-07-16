package awsutil

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const aws4Request = "aws4_request"

func Sign(req *http.Request, payload string, service string, region string, credentials Credentials, now time.Time) {
	amzDate := now.UTC().Format("20060102T150405Z")
	dateStamp := now.UTC().Format("20060102")
	scope := dateStamp + "/" + strings.TrimSpace(region) + "/" + service + "/" + aws4Request

	req.Header.Set("X-Amz-Date", amzDate)
	if token := strings.TrimSpace(credentials.SessionToken); token != "" {
		req.Header.Set("X-Amz-Security-Token", token)
	}

	signedHeaders := "content-type;host;x-amz-date"
	canonicalHeaders := "content-type:" + req.Header.Get("Content-Type") + "\n" +
		"host:" + req.URL.Host + "\n" +
		"x-amz-date:" + amzDate + "\n"
	if strings.TrimSpace(credentials.SessionToken) != "" {
		signedHeaders = "content-type;host;x-amz-date;x-amz-security-token"
		canonicalHeaders += "x-amz-security-token:" + strings.TrimSpace(credentials.SessionToken) + "\n"
	}

	canonicalRequest := strings.Join([]string{
		req.Method,
		canonicalURI(req.URL),
		"",
		canonicalHeaders,
		signedHeaders,
		sha256Hex(payload),
	}, "\n")
	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		amzDate,
		scope,
		sha256Hex(canonicalRequest),
	}, "\n")
	signature := hex.EncodeToString(hmacSHA256(signingKey(strings.TrimSpace(credentials.SecretAccessKey), dateStamp, strings.TrimSpace(region), service), stringToSign))
	req.Header.Set("Authorization", "AWS4-HMAC-SHA256 Credential="+strings.TrimSpace(credentials.AccessKeyID)+"/"+scope+", SignedHeaders="+signedHeaders+", Signature="+signature)
}

func canonicalURI(u *url.URL) string {
	if u.EscapedPath() == "" {
		return "/"
	}
	return u.EscapedPath()
}

func signingKey(secret string, dateStamp string, region string, service string) []byte {
	dateKey := hmacSHA256([]byte("AWS4"+secret), dateStamp)
	dateRegionKey := hmacSHA256(dateKey, region)
	dateRegionServiceKey := hmacSHA256(dateRegionKey, service)
	return hmacSHA256(dateRegionServiceKey, aws4Request)
}

func hmacSHA256(key []byte, data string) []byte {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(data))
	return mac.Sum(nil)
}

func sha256Hex(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
