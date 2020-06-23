package tokenup_sdk

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"strings"
	"testing"
)

func TestClient_SendTx(t *testing.T) {
	txData, err := GetTxData("mint", common.HexToAddress("0x5bb297a46512233e9f52f74e0cafd6ecb2d2db07"), big.NewInt(1))
	if err != nil {
		t.Error(err)
		return
	}
	if _, err := GetClient().SendTx(TransactRequest{
		From: "0x5bb297a46512233e9f52f74e0cafd6ecb2d2db07",
		To:   "0x0cFAD7a5D86c6880e9E2cACd84c5DF520beAa4CF",
		Data: txData,
	}); err != nil {
		t.Error(err)
		return
	}
}

func TestClient_TxDetail(t *testing.T) {
	if res, err := GetClient().TxDetail("0xeb4fdc1cf0b518815f413baeb969b2bf2d1fff7ac9047a894ca523224f9c0657"); err != nil {
		t.Error(err)
		return
	} else {
		t.Logf("%+v", res)
	}
}

func TestClient_Call(t *testing.T) {
	callData, err := GetTxData("baseTokenURI")
	if err != nil {
		t.Error(err)
		return
	}
	contractABI, err := abi.JSON(strings.NewReader(SolidityABI))
	if err != nil {
		t.Error(err)
		return
	}
	out := ""
	if err := GetClient().Call(CallRequest{
		From:   "0x5bb297a46512233e9f52f74e0cafd6ecb2d2db07",
		To:     "0x0cFAD7a5D86c6880e9E2cACd84c5DF520beAa4CF",
		Data:   callData,
		Method: "baseTokenURI",
	}, contractABI, &out); err != nil {
		t.Error(err)
		return
	}
	t.Logf("%v", out)
}
