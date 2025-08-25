package config

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"os"
	"sort"
	"strings"
)

const (
	VnpVersion   = "2.1.0"
	VnpCommand   = "pay"
	VnpPayURL    = "https://sandbox.vnpayment.vn/paymentv2/vpcpay.html"
	VnpReturnURL = "http://localhost:5173/home/payment"
	VnpApiURL    = "https://sandbox.vnpayment.vn/merchant_webapi/api/transaction"
	Vnp_IpAddr   = "127.0.0.1"
	VnpCurrCode  = "VND"
	VnpBankCode  = "NCB"
	VnpLocale    = "vn"
	
)

// HMAC SHA512 generates a secure hash using a secret key
//VNPay requires HmacSha512 signatures for request validation
func HmacSHA512(key, data string) string {
	h := hmac.New(sha512.New, []byte(key))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// HashAllFields builds a query string from all fields and returns its HmacSha512 signature.
//VnPAY mandates that all params must be stored and signed to prevent tampering.
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

	return HmacSHA512(os.Getenv("VNPAY_SECRET"), sb.String())
}
