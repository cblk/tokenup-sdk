package tokenup_sdk

import (
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/guonaihong/gout"
	"github.com/pborman/uuid"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	client *Client
)

type Authorize struct {
	SignerUrl              string
	AppId                  string
	AppKey                 string
	NotifyUrl              string
	PrivateKey             string
	CallBackPartyPublicKey string
	SignerVersion          string
}
type NodeConfig struct {
	NodeUrl       string
	FeeLimit      int64
	GasPriceMin   int64
	GasPriceMax   int64
	NodeVersion   string
	NodeNotifyUrl string
}

type Client struct {
	NodeConfig
	Authorize
}

func Init(c *Client) {
	if c != nil {
		client = c
		if client.GasPriceMin == 0 {
			client.GasPriceMin = 2000000000 // 2 Gwei
		}
		if client.GasPriceMax == 0 {
			client.GasPriceMax = 30000000000 // 30 Gwei
		}
		if client.FeeLimit == 0 {
			client.FeeLimit = 50000000 // 0.05 Ether
		}
		if client.NodeVersion == "" {
			client.NodeVersion = "v1"
		}
	}
}

func GetClient() *Client {
	if client == nil {
		panic("client is nil, please call tokenup_sdk.Init first")
	}
	return client
}

func (client *Client) SignSync(signSource SignSource, timeoutSeconds int) (string, string, error) {
	result, err := client.SignHash(signSource)
	if err != nil {
		return "", "", err
	}
	requestId := result.Data.(map[string]interface{})["request_id"].(string)
	timeout := time.After(time.Duration(timeoutSeconds) * time.Second)
	for {
		select {
		case <-timeout:
			return "", requestId, errors.New("timeout signer request")
		case <-time.After(200 * time.Millisecond):
			signerResult, err := client.OnTracing(requestId)
			if err != nil {
				return "", requestId, err
			}
			if signerResult.Data != nil {
				return signerResult.Data.(map[string]interface{})["result"].(map[string]interface{})["data"].(string), requestId, nil
			}
		}
	}
}

func (client *Client) SignHash(signSource SignSource) (Result, error) {
	var ps ProxySignSafe
	structCopy(&signSource, &ps)
	url := client.SignerUrl
	if client.SignerVersion != "" {
		url += "/v2.0.0"
	}
	url += "/vendor/proxy/sign_hash"
	return client.signerPost(url, &ps)
}

func (client *Client) BatchSignHash(signSource SignSource) (Result, error) {
	var ps ProxySignSafe
	structCopy(&signSource, &ps)
	url := client.SignerUrl
	if client.SignerVersion != "" {
		url += "/batch_sign"
	}
	url += "/vendor/proxy/pending_sign_hash"
	return client.signerPost(url, &ps)
}

func (client *Client) OnTracing(requestId string) (Result, error) {
	traceSafe := TraceSafe{
		RequestId: requestId,
	}
	url := client.SignerUrl
	if client.SignerVersion != "" {
		url += "/v2.0.0"
	}
	url += "/vendor/status/tracing"
	return client.signerPost(url, &traceSafe)
}

func (client *Client) GetTxStatus(requestId string) (Result, error) {
	url := client.SignerUrl
	if client.SignerVersion != "" {
		url += "/v2.0.0"
	}
	url += "/vendor/tx/status/" + requestId
	var result Result
	code := 0
	if err := gout.GET(url).BindJSON(&result).Code(&code).Do(); err != nil {
		return Result{}, err
	}
	if code != 200 {
		return result, fmt.Errorf("%d-%s", code, result.Status.Message)
	}
	return result, nil
}

func (client *Client) signerPost(url string, data interface{}) (Result, error) {
	value := reflect.ValueOf(data).Elem()
	value.FieldByName("Timestamp").Set(reflect.ValueOf(time.Now().Unix()))
	value.FieldByName("AppKey").Set(reflect.ValueOf(client.Authorize.AppKey))
	nuVal := value.FieldByName("NotifyUrl")
	if nuVal.CanSet() {
		nuVal.Set(reflect.ValueOf(client.Authorize.NotifyUrl))
	}
	nonceVal := value.FieldByName("Nonce")
	if nonceVal.CanSet() && nonceVal.Interface() == nil || nonceVal.Interface().(string) == "" {
		nonceVal.Set(reflect.ValueOf(strconv.FormatInt(randInt64(), 10)))
	}
	value.FieldByName("AppId").Set(reflect.ValueOf(client.Authorize.AppId))
	var Signature string
	var err error
	Signature, err = RsaSignAndPrivate([]byte(EncodeString(data)), client.Authorize.PrivateKey)
	if err != nil {
		return Result{}, err
	}
	signValue := value.FieldByName("Signature")
	signValue.Set(reflect.ValueOf(Signature))
	var result Result
	code := 0
	if err := gout.POST(url).SetJSON(data).BindJSON(&result).Code(&code).Do(); err != nil {
		return Result{}, err
	}
	if code != 200 {
		return result, fmt.Errorf("%d-%s", code, result.Status.Message)
	}
	return result, nil
}

func (client *Client) Estimate(req EstimateRequest) (EstimateResponse, error) {
	res := EstimateResponse{}
	url := fmt.Sprintf("%v/%v/%v", client.NodeUrl, client.NodeVersion, "tx/estimate")
	code := 0
	if err := gout.POST(url).SetJSON(req).BindJSON(&res).Code(&code).Do(); err != nil {
		return res, err
	}
	if code != 200 {
		return res, fmt.Errorf("%d-%s", code, res.Message)
	}
	return res, nil
}

func (client *Client) SendTx(req TransactRequest) (TransactResponse, error) {
	res := TransactResponse{}
	// 交易gas相关建议
	estimateResponse, err := client.Estimate(EstimateRequest{
		From:        req.From,
		To:          req.To,
		Data:        req.Data,
		FeeLimit:    client.FeeLimit,
		GasPriceMax: client.GasPriceMax,
		GasPriceMin: client.GasPriceMin,
	})
	if err != nil {
		return res, err
	}
	if req.GasPrice == "" {
		req.GasPrice = estimateResponse.Data.GasPrice
	}
	if req.NotifyUrl == "" {
		req.NotifyUrl = client.NodeNotifyUrl
	}
	if req.Nonce == 0 {
		req.Nonce = estimateResponse.Data.Nonce
	}
	req.GasLimit = estimateResponse.Data.Gas

	//交易数据签名
	txHashData, err := req.decode(estimateResponse.Data.ChainId)
	if err != nil {
		return res, err
	}
	uuid.SetRand(strings.NewReader(req.From + strconv.Itoa(int(time.Now().UTC().Unix()))))
	orderId := "sign_" + uuid.NewUUID().String() + time.Now().UTC().Format("20060102150405")
	signSource := SignSource{
		Address: req.From,
		Data:    txHashData,
		Extras:  "tokenup-sdk",
		OrderID: orderId,
	}
	req.Signature, _, err = client.SignSync(signSource, 5)
	if err != nil {
		return res, err
	}
	// 发送交易
	code := 0
	url := fmt.Sprintf("%v/%v/%v", client.NodeUrl, client.NodeVersion, "tx/transact")
	if err := gout.POST(url).SetJSON(req).BindJSON(&res).Code(&code).Do(); err != nil {
		return res, err
	}
	if code != 200 {
		return res, fmt.Errorf("%d-%s", code, res.Message)
	}
	return res, nil
}

func (client *Client) TxDetail(txHash string) (DetailResponse, error) {
	res := DetailResponse{}
	code := 0
	url := fmt.Sprintf("%v/%v/%v/%v", client.NodeUrl, client.NodeVersion, "tx", txHash)
	if err := gout.GET(url).BindJSON(&res).Code(&code).Do(); err != nil {
		return res, err
	}
	if code != 200 {
		return res, fmt.Errorf("%d-%s", code, res.Message)
	}
	return res, nil
}

func (client *Client) Call(req CallRequest, abi abi.ABI, out interface{}) error {
	res := CallResponse{}
	code := 0
	url := fmt.Sprintf("%v/%v/%v", client.NodeUrl, client.NodeVersion, "tx/call")
	if err := gout.POST(url).SetJSON(req).BindJSON(&res).Code(&code).Do(); err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("%d-%s", code, res.Message)
	}
	data, err := hexutil.Decode(res.Data)
	if err != nil {
		return err
	}
	return abi.Unpack(out, req.Method, data)
}
