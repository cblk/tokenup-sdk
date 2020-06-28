package tokenup_sdk

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
	"strings"
)

type Response struct {
	Message string `json:"message" description:"接口响应消息，调用成功时返回success，失败时返回具体的失败类型信息"`
}

type EstimateRequest struct {
	From        string `json:"from" validate:"eth_addr" description:"交易发送方地址"`
	To          string `json:"to" validate:"omitempty,eth_addr" description:"交易目标地址"`
	Data        string `json:"data" validate:"omitempty,is_hex" description:"交易数据(16进制字符串)"`
	FeeLimit    int64  `json:"fee_limit" validate:"lte=500000000" description:"交易费用上限，不大于500000000（Gwei）"`
	GasPriceMin int64  `json:"gas_price_min" validate:"gte=1000000000" description:"gas price下限，不小于1000000000（Wei）"`
	GasPriceMax int64  `json:"gas_price_max" validate:"lte=100000000000" description:"gas price上限，不大于100000000000（Wei）"`
}

type EstimateResponse struct {
	Response
	Data struct {
		GasPrice string `json:"gas_price" description:"当前节点建议的gas price(单位Wei，16进制字符串)"`
		Gas      string `json:"gas" description:"节点估计的交易gas使用量(16进制字符串)"`
		Nonce    uint64 `json:"nonce" description:"交易的序列号"`
		ChainId  int64  `json:"chain_id" description:"以太坊节点的chain id"`
	} `json:"data"`
}

type TransactRequest struct {
	From      string `json:"from" validate:"eth_addr" description:"交易发送方地址"`
	To        string `json:"to" validate:"omitempty,eth_addr" description:"交易目标地址"`
	Nonce     uint64 `json:"nonce" description:"交易序列号"`
	Data      string `json:"data" validate:"omitempty,is_hex" description:"交易数据(16进制字符串)"`
	Value     string `json:"value" validate:"omitempty,is_hex_num" description:"要发送到目标地址的以太数量(16进制字符串)"`
	GasPrice  string `json:"gas_price" validate:"is_hex_num" description:"交易发送方愿意支付的gas价格(16进制字符串)"`
	GasLimit  string `json:"gas_limit" validate:"is_hex_num" description:"交易的gas上限(16进制字符串)"`
	Signature string `json:"signature" validate:"signature" description:"交易数据签名(16进制字符串)"`
	NotifyUrl string `json:"notify_url" validate:"omitempty,url" description:"通知回调url"`
}

type TransactResponse struct {
	Response
	Data struct {
		GasPrice     string `json:"gas_price" description:"交易发送方愿意支付的gas价格(16进制字符串)"`
		GasLimit     string `json:"gas_limit" description:"交易的gas上限(16进制字符串)"`
		TxHash       string `json:"tx_hash" description:"交易哈希"`
		Status       int    `json:"status" description:"交易状态：0=None 1=Pending 2=Confirmed 3=Failed"`
		NotifyStatus int    `json:"notify_status" description:"通知状态：0=未通知 1=已通知"`
		Type         int    `json:"type" description:"交易类型：0=创建合约 1=合约调用 2=转账"`
	} `json:"data" description:"发送交易结果"`
}

type DetailRequest struct {
	TxHash string `json:"tx_hash" path:"tx_hash" validate:"is_hex" description:"交易哈希"`
}

type DetailResponse struct {
	Response
	Data struct {
		GasPrice     string `json:"gas_price" description:"交易发送方愿意支付的gas价格(16进制字符串)"`
		GasLimit     string `json:"gas_limit" description:"交易的gas上限(16进制字符串)"`
		TxHash       string `json:"tx_hash" description:"交易哈希"`
		Status       int    `json:"status" description:"交易状态：0=None 1=Pending 2=Confirmed 3=Failed"`
		NotifyStatus int    `json:"notify_status" description:"通知状态：0=未通知 1=已通知"`
		Type         int    `json:"type" description:"交易类型：0=创建合约 1=合约调用 2=转账"`
	} `json:"data" description:"交易详情"`
}

type CallRequest struct {
	From   string `json:"from" validate:"eth_addr" description:"消息调用发送方地址"`
	To     string `json:"to" validate:"omitempty,eth_addr" description:"消息调用目标地址"`
	Data   string `json:"data" validate:"omitempty,is_hex" description:"消息调用数据(16进制字符串)"`
	Method string `json:"method"`
}

type CallResponse struct {
	Response
	Data string `json:"data" description:"消息调用结果"`
}

func (tx TransactRequest) decode(chainId int64) (string, error) {
	amount, _ := hexutil.DecodeBig(tx.Value)
	gasLimit, _ := hexutil.DecodeUint64(tx.GasLimit)
	gasPrice, _ := hexutil.DecodeBig(tx.GasPrice)
	data, _ := hexutil.Decode(tx.Data)
	var tran *types.Transaction
	if tx.To == "" {
		tran = types.NewContractCreation(
			tx.Nonce,
			amount,
			gasLimit,
			gasPrice,
			data,
		)
	} else {
		tran = types.NewTransaction(
			tx.Nonce,
			common.HexToAddress(tx.To),
			amount,
			gasLimit,
			gasPrice,
			data,
		)
	}
	mySigner := types.NewEIP155Signer(big.NewInt(chainId))
	h := mySigner.Hash(tran)
	return hexutil.Encode(h[:]), nil
}

func GetData(abiStr, name string, args ...interface{}) (string, error) {
	contractABI, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return "", nil
	}
	data, err := contractABI.Pack(name, args...)
	return hexutil.Encode(data), err
}
