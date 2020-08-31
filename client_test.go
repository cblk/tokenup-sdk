package tokenup_sdk_test

import (
	"github.com/cblk/tokenup-sdk"
	"testing"
)

func TestClient_TxDetail(t *testing.T) {
	if res, err := tokenup_sdk.GetClient().TxDetail("0xeb4fdc1cf0b518815f413baeb969b2bf2d1fff7ac9047a894ca523224f9c0657"); err != nil {
		t.Error(err)
		return
	} else {
		t.Logf("%+v", res)
	}
}

func TestClient_EventQuery(t *testing.T) {
	req := tokenup_sdk.QueryRequest{
		BlockHash: "",
		FromBlock: 0,
		ToBlock:   0,
		Addresses: []string{"0xdF0F1b2Fa2992247ffC68790B20a3Df1E7514B68"},
		Topics:    [][]string{{"0x75fd34f1ce6e8316ca4247d91c07cb4f00667d4deab5d89efb099137e87711b9"}},
	}
	if res, err := tokenup_sdk.GetClient().EventQuery(req); err != nil {
		t.Error(err)
		return
	} else {
		t.Logf("%+v", res)
	}
}