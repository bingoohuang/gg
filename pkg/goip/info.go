package goip

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// https://topic.alibabacloud.com/a/go-combat-golang-get-public-ip-view-intranet-ip-detect-ip-type-verify-ip-range-ip-address-string-and-int-conversion-judge-by-ip_1_38_10267608.html

// Info ...
type Info struct {
	Code int `json:"code"`
	Data IP  `json:"data"`
}

// IP ...
type IP struct {
	Country   string `json:"country"`
	CountryID string `json:"country_id"`
	Area      string `json:"area"`
	AreaID    string `json:"area_id"`
	Region    string `json:"region"`
	RegionID  string `json:"region_id"`
	City      string `json:"city"`
	CityID    string `json:"city_id"`
	Isp       string `json:"isp"`
}

// TabaoAPI ...
func TabaoAPI(ip string) (*Info, error) {
	ctx, cncl := context.WithTimeout(context.Background(), time.Second*10)
	defer cncl()

	addr := "http://ip.taobao.com/service/getIpInfo.php?ip=" + ip
	resp, err := http.NewRequestWithContext(ctx, http.MethodGet, addr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed http.Get(%s), × err: %w", addr, err)
	}

	defer resp.Body.Close()

	out, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed ioutil.ReadAll, × err: %w", err)
	}

	var result Info

	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("failed json.Unmarshal %s, × err: %w", string(out), err)
	}

	return &result, nil
}
