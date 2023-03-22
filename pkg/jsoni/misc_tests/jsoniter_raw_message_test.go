package misc_tests

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/bingoohuang/gg/pkg/jsoni"
	"github.com/stretchr/testify/require"
)

func Test_jsoniter_RawMessage(t *testing.T) {
	should := require.New(t)
	var data jsoni.RawMessage
	should.Nil(jsoni.Unmarshal([]byte(`[1,2,3]`), &data))
	should.Equal(`[1,2,3]`, string(data))
	str, err := jsoni.MarshalToString(data)
	should.Nil(err)
	should.Equal(`[1,2,3]`, str)
}

func Test_encode_map_of_jsoniter_raw_message(t *testing.T) {
	should := require.New(t)
	type RawMap map[string]*jsoni.RawMessage
	value := jsoni.RawMessage("[]")
	rawMap := RawMap{"hello": &value}
	output, err := jsoni.MarshalToString(rawMap)
	should.Nil(err)
	should.Equal(`{"hello":[]}`, output)
}

func Test_marshal_invalid_json_raw_message(t *testing.T) {
	type A struct {
		Raw json.RawMessage `json:"raw"`
	}
	message := []byte(`{}`)

	a := A{}
	should := require.New(t)
	should.Nil(jsoni.ConfigCompatibleWithStandardLibrary.Unmarshal(context.Background(), message, &a))
	aout, aouterr := jsoni.ConfigCompatibleWithStandardLibrary.Marshal(context.Background(), &a)
	should.Equal(`{"raw":null}`, string(aout))
	should.Nil(aouterr)
}

func Test_marshal_nil_json_raw_message(t *testing.T) {
	type A struct {
		Nil1 jsoni.RawMessage `json:"raw1"`
		Nil2 json.RawMessage  `json:"raw2"`
	}

	a := A{}
	should := require.New(t)
	aout, aouterr := jsoni.Marshal(&a)
	should.Equal(`{"raw1":null,"raw2":null}`, string(aout))
	should.Nil(aouterr)

	a.Nil1 = []byte(`Any`)
	a.Nil2 = []byte(`Any`)
	should.Nil(jsoni.Unmarshal(aout, &a))
	should.Nil(a.Nil1)
	should.Nil(a.Nil2)
}

func Test_raw_message_memory_not_copied_issue(t *testing.T) {
	jsonStream := `{"name":"xxxxx","bundle_id":"com.zonst.majiang","app_platform":"ios","app_category":"100103", "budget_day":1000,"bidding_min":1,"bidding_max":2,"bidding_type":"CPM", "freq":{"open":true,"type":"day","num":100},"speed":1, "targeting":{"vendor":{"open":true,"list":["zonst"]}, "geo_code":{"open":true,"list":["156110100"]},"app_category":{"open":true,"list":["100101"]}, "day_parting":{"open":true,"list":["100409","100410"]},"device_type":{"open":true,"list":["ipad"]}, "os_version":{"open":true,"list":[10]},"carrier":{"open":true,"list":["mobile"]}, "network":{"open":true,"list":["4G"]}},"url":{"tracking_imp_url":"http://www.baidu.com", "tracking_clk_url":"http://www.baidu.com","jump_url":"http://www.baidu.com","deep_link_url":"http://www.baidu.com"}}`
	type IteratorObject struct {
		Name        *string           `json:"name"`
		BundleId    *string           `json:"bundle_id"`
		AppCategory *string           `json:"app_category"`
		AppPlatform *string           `json:"app_platform"`
		BudgetDay   *float32          `json:"budget_day"`
		BiddingMax  *float32          `json:"bidding_max"`
		BiddingMin  *float32          `json:"bidding_min"`
		BiddingType *string           `json:"bidding_type"`
		Freq        *jsoni.RawMessage `json:"freq"`
		Targeting   *jsoni.RawMessage `json:"targeting"`
		Url         *jsoni.RawMessage `json:"url"`
		Speed       *int              `json:"speed" db:"speed"`
	}

	obj := &IteratorObject{}
	decoder := jsoni.NewDecoder(strings.NewReader(jsonStream))
	err := decoder.Decode(context.Background(), obj)
	should := require.New(t)
	should.Nil(err)
	should.Equal(`{"open":true,"type":"day","num":100}`, string(*obj.Freq))
}
