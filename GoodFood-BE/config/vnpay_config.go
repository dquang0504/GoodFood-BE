package config

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"
)

const (
	VnpVersion   = "2.1.0"
	VnpCommand   = "pay"
	VnpPayURL    = "https://sandbox.vnpayment.vn/paymentv2/vpcpay.html"
	VnpReturnURL = "http://localhost:5173/home/payment"
	VnpTmnCode   = "GHFBIZ5U"
	SecretKey    = "ZFSNKTR0QRTMJWOSXVYQNM6QYI0KPV05"
	VnpApiURL    = "https://sandbox.vnpayment.vn/merchant_webapi/api/transaction"
	Vnp_IpAddr   = "127.0.0.1"
	VnpCurrCode  = "VND"
	VnpBankCode  = "NCB"
	VnpLocale    = "vn"
	
)

//MD5 Hash
func MD5(message string) string{
	hash := md5.Sum([]byte(message))
	return hex.EncodeToString(hash[:])
}

// SHA256 hash
func Sha256(message string) string{
	hash := sha256.Sum256([]byte(message))
	return hex.EncodeToString(hash[:])
}

// HMAC SHA512
func HmacSHA512(key, data string) string {
	h := hmac.New(sha512.New, []byte(key))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// HashAllFields: tạo chuỗi query string rồi HMAC SHA512
func HashAllFields(fields map[string]string) string {
	var keys []string
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for i, k := range keys {
		v := fields[k]
		if v != "" {
			sb.WriteString(fmt.Sprintf("%s=%s", k, v))
			if i < len(keys)-1 {
				sb.WriteString("&")
			}
		}
	}

	return HmacSHA512(SecretKey, sb.String())
}

// GetRandomNumber: tạo chuỗi số ngẫu nhiên
func GetRandomNumber(length int) string {
	const digits = "0123456789"
	rand.Seed(time.Now().UnixNano())
	sb := make([]byte, length)
	for i := range sb {
		sb[i] = digits[rand.Intn(len(digits))]
	}
	return string(sb)
}