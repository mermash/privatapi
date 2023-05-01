package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"strings"
)

const (
	PRIVAT_CASH_CURRENCY_API_URL = "https://api.privatbank.ua/p24api/pubinfo?json&exchange&coursid="
)

type CashCurrency struct {
	Ccy     string `json:"ccy"`
	BaseCcy string `json:"base_ccy"`
	Buy     string `json:"buy"`
	Sale    string `json:"sale"`
}

type CashCurrencies struct {
	Currencies []*CashCurrency
}

type ErrorAPI struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func getCashCurrencies(url string) (*CashCurrencies, error) {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("error happend", err)
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error happend", err)
		return nil, err
	}
	fmt.Printf("resp Body: %s\n", string(respBody))

	if !strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
		fmt.Printf("response don't have json: %s", respBody)
		return nil, fmt.Errorf("wrong response: %s", respBody)
	}

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
		if err != nil {
			fmt.Println("parse error happend", err)
			continue
		}
		fmt.Printf("currencies: %s/%s buy - %f, sale: %f: \n", item.Ccy, item.BaseCcy, buy, sale)
	}

	return &CashCurrencies{Currencies: ccy}, nil
}

func main() {

	tmpl, err := template.New("").ParseFiles("cashCurrency.html")
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/cashcurrency", func(w http.ResponseWriter, r *http.Request) {

		coursId := r.URL.Query().Get("coursid")
		_, err := strconv.ParseInt(coursId, 10, 32)
		if err != nil {
			fmt.Fprintln(w, "can't parse get params")
			return
		}
		cashCcy, err := getCashCurrencies(PRIVAT_CASH_CURRENCY_API_URL + coursId)

		if cashCcy != nil {
			tmpl.ExecuteTemplate(w, "cashCurrency.html",
				struct {
					CashCurrencies *CashCurrencies
					IsError        bool
					Error          string
				}{
					CashCurrencies: cashCcy,
					IsError:        false,
					Error:          "",
				})
		} else {
			tmpl.ExecuteTemplate(w, "cashCurrency.html",
				struct {
					CashCurrencies *CashCurrencies
					IsError        bool
					Error          string
				}{
					CashCurrencies: nil,
					IsError:        true,
					Error:          err.Error(),
				})
		}
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello guys!")
		fmt.Fprintln(w, "Do you want to know exchange rates for currencies from privatbank?")
		fmt.Fprintln(w, "Go to `/cachscurrency`")
	})

	fmt.Println("starting server at :8080")
	http.ListenAndServe(":8080", nil)
}
