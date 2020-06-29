package tokenup_sdk

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	rand2 "math/rand"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Status struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
type Result struct {
	Status Status      `json:"status"`
	Data   interface{} `json:"data"`
}
type ProxySignSafe struct {
	AppId     string `json:"app_id" sign:"app_id"`
	AppKey    string `json:"-" sign:"app_key"`
	Nonce     string `json:"nonce" sign:"nonce"`
	Address   string `json:"address" sign:"address"`
	Timestamp int64  `json:"timestamp" sign:"timestamp"`
	Data      string `json:"data" sign:"data"`
	Extras    string `json:"extras" sign:"extras" gorm:"type:text"`
	OrderID   string `json:"order_id" sign:"order_id"`
	Signature string `json:"signature" gorm:"type:text"`
}
type SignSource struct {
	Address string `json:"address"`
	Data    string `json:"data" sign:"data"`
	Extras  string `json:"extras" sign:"extras" gorm:"type:text"`
	OrderID string `json:"order_id" sign:"order_id"`
}
type TraceSafe struct {
	AppId     string `json:"app_id" sign:"app_id"`
	AppKey    string `sign:"app_key" json:"-"`
	Nonce     string `json:"nonce" sign:"nonce"`
	Timestamp int64  `json:"timestamp" sign:"timestamp"`
	RequestId string `json:"request_id" sign:"request_id"`
	Signature string `json:"signature"`
}
type ReceivedConfirm struct {
	Message   string `json:"message" sign:"message"`
	Nonce     string `json:"nonce" sign:"nonce"`
	AppKey    string `json:"-" sign:"app_key"`
	Signature string `json:"signature"`
}

var channel = make(chan int64, 32)

func init() {
	go func() {
		var old int64
		for {
			o := rand2.New(rand2.NewSource(time.Now().UnixNano())).Int63()
			if old != o {
				old = o
				select {
				case channel <- o:
				}
			}
		}
	}()
}
func randInt64() (r int64) {
	select {
	case randInt := <-channel:
		r = randInt
	}
	return
}
func deepFields(ifaceType reflect.Type) []reflect.StructField {
	var fields []reflect.StructField

	for i := 0; i < ifaceType.NumField(); i++ {
		v := ifaceType.Field(i)
		if v.Anonymous && v.Type.Kind() == reflect.Struct {
			fields = append(fields, deepFields(v.Type)...)
		} else {
			fields = append(fields, v)
		}
	}

	return fields
}
func structCopy(srcPtr interface{}, desPtr interface{}) {
	srcv := reflect.ValueOf(srcPtr)
	dstv := reflect.ValueOf(desPtr)
	srct := reflect.TypeOf(srcPtr)
	dstt := reflect.TypeOf(desPtr)
	if srct.Kind() != reflect.Ptr || dstt.Kind() != reflect.Ptr ||
		srct.Elem().Kind() == reflect.Ptr || dstt.Elem().Kind() == reflect.Ptr {
		panic("Fatal error:type of parameters must be Ptr of value")
	}
	if srcv.IsNil() || dstv.IsNil() {
		panic("Fatal error:value of parameters should not be nil")
	}
	srcV := srcv.Elem()
	dstV := dstv.Elem()
	srcfields := deepFields(reflect.ValueOf(srcPtr).Elem().Type())
	for _, v := range srcfields {
		if v.Anonymous {
			continue
		}
		dst := dstV.FieldByName(v.Name)
		src := srcV.FieldByName(v.Name)
		if !dst.IsValid() {
			continue
		}
		if src.Type() == dst.Type() && dst.CanSet() {
			dst.Set(src)
			continue
		}
		if src.Kind() == reflect.Ptr && !src.IsNil() && src.Type().Elem() == dst.Type() {
			dst.Set(src.Elem())
			continue
		}
		if dst.Kind() == reflect.Ptr && dst.Type().Elem() == src.Type() {
			dst.Set(reflect.New(src.Type()))
			dst.Elem().Set(src)
			continue
		}
	}
}

func RsaSignVerAndPublicHex(data []byte, signature, public string) error {
	signatureDecode, err := hex.DecodeString(signature)
	if err != nil {
		return err
	}
	hashed := sha256.Sum256(data)
	buff, _ := base64.StdEncoding.DecodeString(public)
	// 解析公钥
	pubInterface, err := x509.ParsePKIXPublicKey(buff)
	if err != nil {
		return err
	}
	// 类型断言
	pub := pubInterface.(*rsa.PublicKey)
	//验证签名
	return rsa.VerifyPKCS1v15(pub, crypto.SHA256, hashed[:], signatureDecode)
}
func RsaSignAndPrivate(data []byte, privateKey string) (string, error) {
	h := sha256.New()
	h.Write(data)
	hashed := h.Sum(nil)
	buff, _ := base64.StdEncoding.DecodeString(privateKey)
	//获取私钥
	priv, err := x509.ParsePKCS1PrivateKey(buff)
	if err != nil {
		return "", err
	}
	sign, err := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, hashed)
	return hex.EncodeToString(sign), err
}

type Dec []string

func (p Dec) Len() int           { return len(p) }
func (p Dec) Less(i, j int) bool { return p[i] < p[j] }
func (p Dec) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func EncodeString(o interface{}) string {
	dec := encodeField("", o)
	sort.Stable(dec)
	return strings.Join(dec, "&")
}

func encodeField(prefix string, o interface{}) Dec {
	t := reflect.TypeOf(o)
	v := reflect.ValueOf(o)
	valStr, err := getValueString(v)
	var dec Dec
	if err == nil {
		values := url.Values{}
		values.Set("url", valStr)
		return append(dec, prefix+"="+strings.Split(values.Encode(), "=")[1])
	}
	if prefix != "" {
		prefix += "."
	}

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		prefix2 := t.Field(i).Tag.Get("sign")
		value := v.Field(i)
		if prefix2 == "" {
			continue
		}
		arrs := strings.Split(prefix2, ",")
		if len(arrs) >= 1 {
			prefix2 = arrs[0]
		}
		if isArray(value) {
			for j := 0; j < value.Len(); j++ {
				dec = append(dec, encodeField(prefix+prefix2+"[]", value.Index(j).Interface())...)
			}
		} else if value.Kind() == reflect.Map {
			dataMap := value.Interface().(map[string]interface{})
			for k, v := range dataMap {
				dec = append(dec, encodeField(prefix+prefix2+"."+k, v)...)
			}

		} else {
			dec = append(dec, encodeField(prefix+prefix2, value.Interface())...)
		}

	}
	return dec
}
func getValueString(value reflect.Value) (string, error) {
	if value.Type().String() == "string" {
		return value.Interface().(string), nil
	}
	if value.Type().String() == "bool" {
		val := value.Interface().(bool)
		if val {
			return "true", nil
		} else {
			return "false", nil
		}
	}
	if value.Type().String() == "int8" {
		val := value.Interface().(int8)
		return strconv.FormatInt(int64(val), 10), nil
	}
	if value.Type().String() == "int16" {
		val := value.Interface().(int16)
		return strconv.FormatInt(int64(val), 10), nil
	}
	if value.Type().String() == "int32" {
		val := value.Interface().(int32)
		return strconv.FormatInt(int64(val), 10), nil
	}
	if value.Type().String() == "int64" {
		val := value.Interface().(int64)
		return strconv.FormatInt(val, 10), nil
	}
	if value.Type().String() == "uint8" {
		val := value.Interface().(uint8)
		return strconv.FormatUint(uint64(val), 10), nil
	}
	if value.Type().String() == "uint16" {
		val := value.Interface().(uint16)
		return strconv.FormatUint(uint64(val), 10), nil
	}
	if value.Type().String() == "uint32" {
		val := value.Interface().(uint32)
		return strconv.FormatUint(uint64(val), 10), nil
	}
	if value.Type().String() == "int" {
		val := value.Interface().(int)
		return strconv.FormatInt(int64(val), 10), nil
	}
	if value.Type().String() == "uint" {
		val := value.Interface().(uint)
		return strconv.FormatUint(uint64(val), 10), nil
	}
	if value.Type().String() == "uint64" {
		val := value.Interface().(uint64)
		return strconv.FormatUint(val, 10), nil
	}
	if value.Type().String() == "float32" {
		val := value.Interface()
		var timeFloat64 float64
		_, _ = fmt.Sscanf(fmt.Sprint(val), "%e", &timeFloat64)
		return strconv.FormatFloat(timeFloat64, 'E', -1, 32), nil
	}
	if value.Type().String() == "float64" {
		val := value.Interface()
		var timeFloat64 float64
		_, _ = fmt.Sscanf(fmt.Sprint(val), "%e", &timeFloat64)
		return strconv.FormatFloat(timeFloat64, 'E', -1, 64), nil
	}
	if value.Type().String() == "uint32" {
		val := value.Interface().(uint32)
		return strconv.FormatUint(uint64(val), 10), nil
	}

	return "", errors.New("invalid field")

}

func isArray(value reflect.Value) bool {
	if strings.HasPrefix(value.Type().String(), "[]") {
		return true
	}
	return false
}
