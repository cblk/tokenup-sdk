package tokenup_sdk_test

import (
	"github.com/cblk/tokenup-sdk"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"strings"
	"testing"
)

func TestClient_SendTx(t *testing.T) {
	txData, err := tokenup_sdk.GetTxData("mint", common.HexToAddress("0x5bb297a46512233e9f52f74e0cafd6ecb2d2db07"), big.NewInt(3))
	if err != nil {
		t.Error(err)
		return
	}
	if _, err := tokenup_sdk.GetClient().SendTx(tokenup_sdk.TransactRequest{
		From: "0x5bb297a46512233e9f52f74e0cafd6ecb2d2db07",
		To:   "0x0cFAD7a5D86c6880e9E2cACd84c5DF520beAa4CF",
		Data: txData,
	}); err != nil {
		t.Error(err)
		return
	}
}

func TestClient_TxDetail(t *testing.T) {
	if res, err := tokenup_sdk.GetClient().TxDetail("0xeb4fdc1cf0b518815f413baeb969b2bf2d1fff7ac9047a894ca523224f9c0657"); err != nil {
		t.Error(err)
		return
	} else {
		t.Logf("%+v", res)
	}
}

func TestClient_Call(t *testing.T) {
	callData, err := tokenup_sdk.GetTxData("baseTokenURI")
	if err != nil {
		t.Error(err)
		return
	}
	contractABI, err := abi.JSON(strings.NewReader(tokenup_sdk.SolidityABI))
	if err != nil {
		t.Error(err)
		return
	}
	out := ""
	if err := tokenup_sdk.GetClient().Call(tokenup_sdk.CallRequest{
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
