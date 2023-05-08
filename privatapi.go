package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"text/template"

	_ "net/http/pprof"

	"github.com/joho/godotenv"
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

type Result struct {
	CashCurrencies *CashCurrencies
	Error          error
}

func getCashCurrencies(wg *sync.WaitGroup, workerNum int, ch chan Result, url string) /*(*CashCurrencies, error)*/ {
	defer wg.Done()

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("error happend", err)
		ch <- Result{
			CashCurrencies: nil,
			Error:          err,
		}
		return //nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error happend", err)
		ch <- Result{
			CashCurrencies: nil,
			Error:          err,
		}
		return //nil, err
	}
	fmt.Printf("resp Body: %s\n", string(respBody))

	if !strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
		fmt.Printf("response don't have json: %s", respBody)
		ch <- Result{
			CashCurrencies: nil,
			Error:          fmt.Errorf("wrong response: %s", respBody),
		}
		return //nil, fmt.Errorf("wrong response: %s", respBody)
	}

	var ccy []*CashCurrency
	var errAPI ErrorAPI
	if resp.StatusCode == http.StatusOK {
		err = json.Unmarshal(respBody, &ccy)
		if err != nil {
			fmt.Println("error happned", err)
			ch <- Result{
				CashCurrencies: nil,
				Error:          err,
			}
			return //nil, err
		}
	} else {
		err = json.Unmarshal(respBody, &errAPI)
		if err != nil {
			fmt.Println("json error", err)

			ch <- Result{
				CashCurrencies: nil,
				Error:          err,
			}
			return //nil, err
		}
		fmt.Println("get error from service", errAPI)

		ch <- Result{
			CashCurrencies: nil,
			Error:          fmt.Errorf(errAPI.Message),
		}
		return //nil, fmt.Errorf(errAPI.Message)
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

	ch <- Result{
		CashCurrencies: &CashCurrencies{
			Currencies: ccy,
		},
		Error: nil,
	}
	//return //&CashCurrencies{Currencies: ccy}, nil
	fmt.Println("workerNum finished = ", workerNum)
}

type HandlerPrivat struct {
	URL  string
	Tmpl *template.Template
}

func getAllCashCurrencies(urlAPI string) (*CashCurrencies, error) {
	ch := make(chan Result, 1)
	wg := &sync.WaitGroup{}
	for i := 1; i <= 10; i++ {
		wg.Add(1)
		go getCashCurrencies(wg, i, ch, urlAPI+strconv.Itoa(i))
		// if err != nil {
		// 	fmt.Printf("error happned for coursid = [%d]\n", i, err)
		// 	continue
		// }
	}
	go func() {
		wg.Wait()
		close(ch)
	}()
	ccy := []*CashCurrency{}
	errAPI := ""
	for r := range ch {
		if r.CashCurrencies != nil {
			ccy = append(ccy, r.CashCurrencies.Currencies...)
			errAPI = fmt.Sprintf("%s \n %s", errAPI, r.Error)
		}
	}
	if len(ccy) > 0 {
		return &CashCurrencies{Currencies: ccy}, nil
	} else {
		return nil, fmt.Errorf(errAPI)
	}
}

func (h *HandlerPrivat) handleCashCurrency(w http.ResponseWriter, r *http.Request) {
	ccy, err := getAllCashCurrencies(h.URL)
	if ccy != nil {
		h.Tmpl.ExecuteTemplate(w, "cashCurrency.html",
			struct {
				CashCurrencies *CashCurrencies
				IsError        bool
				Error          string
			}{
				CashCurrencies: ccy,
				IsError:        false,
				Error:          "",
			})
	} else {
		h.Tmpl.ExecuteTemplate(w, "cashCurrency.html",
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
}

func main() {

	_, err := os.Stat(".env")
	if err == nil {
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatal(err)
		}
	}

	port := os.Getenv("PORT")
	url := os.Getenv("PRIVAT_CASH_CURRENCY_API_URL")

	tmpl, err := template.New("").ParseFiles("cashCurrency.html")
	if err != nil {
		panic(err)
	}

	privatHandler := &HandlerPrivat{
		URL:  url,
		Tmpl: tmpl,
	}

	http.HandleFunc("/cashcurrency", privatHandler.handleCashCurrency)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello guys!")
		fmt.Fprintln(w, "Do you want to know exchange rates for currencies from privatbank?")
		fmt.Fprintln(w, "Go to `/cachscurrency`")
	})

	fmt.Println("starting server at :", port)
	http.ListenAndServe(":"+port, nil)
}
