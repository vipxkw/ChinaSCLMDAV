package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

// randRead fills b with cryptographically secure random bytes.
func randRead(b []byte) {
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
}

// GenerateTOTPSecret returns a random base32 secret (160-bit).
func GenerateTOTPSecret() string {
	b := make([]byte, 20)
	randRead(b)
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)
}

// TOTPCode computes the current 6-digit TOTP code for a secret at time t.
func TOTPCode(secret string, t time.Time) string {
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(strings.TrimSpace(secret)))
	if err != nil {
		return ""
	}
	counter := uint64(t.Unix()) / 30
	var msg [8]byte
	binary.BigEndian.PutUint64(msg[:], counter)
	mac := hmac.New(sha1.New, key)
	mac.Write(msg[:])
	sum := mac.Sum(nil)
	offset := sum[len(sum)-1] & 0x0f
	code := (uint32(sum[offset])&0x7f)<<24 |
		(uint32(sum[offset+1])&0xff)<<16 |
		(uint32(sum[offset+2])&0xff)<<8 |
		uint32(sum[offset+3])&0xff
	return fmt.Sprintf("%06d", code%1000000)
}

// VerifyTOTP validates a code with a ±1 step window.
func VerifyTOTP(secret, code string, t time.Time) bool {
	if len(code) != 6 {
		return false
	}
	now := t.Unix()
	for i := int64(-1); i <= 1; i++ {
		if TOTPCode(secret, time.Unix(now+i*30, 0)) == code {
			return true
		}
	}
	return false
}

// OTPURI builds the otpauth:// provisioning URI for an authenticator app.
func OTPURI(account, issuer, secret string) string {
	acct := strings.ReplaceAll(account, " ", "%20")
	return fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s&algorithm=SHA1&digits=6&period=30",
		issuer, acct, secret, issuer)
}