package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"
)

type TestCase struct {
	ID      string
	Result  *CashCurrencies
	IsError bool
}

func getCashCurrenciesDummy(w http.ResponseWriter, r *http.Request) {
	key := r.FormValue("coursid")
	switch key {
	case "5":
		w.WriteHeader(http.StatusOK)
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte(`[{"ccy":"EUR","base_ccy":"UAH","buy":"40.50000","sale":"41.50000"},{"ccy":"USD","base_ccy":"UAH","buy":"37.22000","sale":"37.72000"}]`))
		return
	default:
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, `{"status":"error","message":"service temporarily unavailable"}`)
	}
}

func TestGetCashCurrencies(t *testing.T) {

	cases := []TestCase{
		TestCase{
			ID: "5",
			Result: &CashCurrencies{
				Currencies: []*CashCurrency{
					&CashCurrency{
						Ccy:     "EUR",
						BaseCcy: "UAH",
						Buy:     "40.500000",
						Sale:    "41.500000",
					},
					&CashCurrency{
						Ccy:     "USD",
						BaseCcy: "UAH",
						Buy:     "37.220001",
						Sale:    "37.720001",
					},
				},
			},
			IsError: false,
		},
		TestCase{
			ID:      "4",
			Result:  nil,
			IsError: true,
		},
		TestCase{
			ID:      "fail",
			Result:  nil,
			IsError: true,
		},
	}
	ts := httptest.NewServer(http.HandlerFunc(getCashCurrenciesDummy))
	for caseNum, item := range cases {
		result, err := getCashCurrencies(ts.URL + "?coursid=" + item.ID)

		if err != nil && !item.IsError {
			t.Errorf("\n\n[%d] unexpected error: %s\n\n", caseNum, err)
			continue
		}
		if err == nil && item.IsError {
			t.Errorf("\n\n[%d] expected error, got nil\n\n", caseNum)
			continue
		}
		if result == nil {
			continue
		}
		if result != nil && len(item.Result.Currencies) != len(result.Currencies) {
			t.Errorf("\n\n[%d] wrong result, expected %#v, got %#v\n\n", caseNum, item.Result, result)
			continue
		}
		if item.Result != nil && len(item.Result.Currencies) <= 0 {
			continue
		}
		for id, value := range result.Currencies {
			if !reflect.DeepEqual(value.Ccy, item.Result.Currencies[id].Ccy) {
				t.Errorf("\n\n[%d] [%d] wrong result, expected %#v, got %#v\n\n", caseNum, id, item.Result.Currencies[id].Ccy, value.Ccy)
			}
		}
	}

}

type CashCurrency struct {
	Ccy     string `json:"ccy"`
	BaseCcy string `json:"base_ccy"`
	Buy     string `json:"buy"`
	sale    string `json:"sale"`
}

type CashCurrencies struct {
	Currencies []*CashCurrency
}

type ErrorAPI struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func getCashCurrencies(url string) (*CashCurrencies, error) {
	fmt.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("error happend", err)
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	fmt.Printf("resp Body: %s\n", string(respBody))

	// if !strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
	// 	fmt.Printf("response don't have json: %s", respBody)
	// 	return nil, fmt.Errorf("wrong response: %s %s", respBody, resp.Header.Get("Content-Type"))
	// }

	var ccy []*CashCurrency
	var errAPI ErrorAPI
	if resp.StatusCode == http.StatusOK {
		err = json.Unmarshal(respBody, &ccy)
		if err != nil {
			fmt.Println("error happned", err)
			return nil, err
		}
	} else {
		err = json.Unmarshal(respBody, &errAPI)
		if err != nil {
			fmt.Println("json error", err)
			return nil, err
		}
		fmt.Println("get error from service", errAPI)
		return nil, fmt.Errorf(errAPI.Message)
	}

	for _, item := range ccy {
		buy, err := strconv.ParseFloat(item.Buy, 32)
		if err != nil {
			fmt.Println("parse error happend", err)
			continue
		}
		sale, err := strconv.ParseFloat(item.Sale, 32)
		fmt.Printf("currencies: %s/%s buy - %f, sale: %f: \n", item.Ccy, item.BaseCcy, buy, sale)
	}

	return &CashCurrencies{Currencies: ccy}, nil
}
