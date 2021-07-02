package eos_go_trace_api_plugin_wrapper

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/token"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

type TraceAPI struct{
	HttpClient *http.Client
	BaseURL    string
	// Header is one or more headers to be added to all outgoing calls
	Header                  http.Header
}

func New(baseURL string) *TraceAPI {
	api := &TraceAPI{
		HttpClient: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
				DisableKeepAlives:     true, // default behavior, because of `nodeos`'s lack of support for Keep alives.
			},
		},
		BaseURL:  strings.TrimRight(baseURL, "/"),
		Header:   make(http.Header),
	}

	return api
}

func (api *TraceAPI) GetBlockByID(num uint32) (out *BlockResp, err error) {
	err = api.call("trace_api", "get_block", eos.M{"block_num": fmt.Sprintf("%d", num)}, &out)
	return
}

func (api *TraceAPI) call(baseAPI string, endpoint string, body interface{}, out interface{}) error {
	jsonBody, err := enc(body)
	if err != nil {
		return err
	}

	targetURL := fmt.Sprintf("%s/v1/%s/%s", api.BaseURL, baseAPI, endpoint)
	req, err := http.NewRequest("POST", targetURL, jsonBody)
	if err != nil {
		return fmt.Errorf("NewRequest: %w", err)
	}

	for k, v := range api.Header {
		if req.Header == nil {
			req.Header = http.Header{}
		}
		req.Header[k] = append(req.Header[k], v...)
	}

	resp, err := api.HttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%s: %w", req.URL.String(), err)
	}
	defer resp.Body.Close()

	var cnt bytes.Buffer
	_, err = io.Copy(&cnt, resp.Body)
	if err != nil {
		return fmt.Errorf("Copy: %w", err)
	}

	if resp.StatusCode == 404 {
		var apiErr eos.APIError
		if err := json.Unmarshal(cnt.Bytes(), &apiErr); err != nil {
			return eos.ErrNotFound
		}
		return apiErr
	}

	if resp.StatusCode > 299 {
		var apiErr eos.APIError
		if err := json.Unmarshal(cnt.Bytes(), &apiErr); err != nil {
			return fmt.Errorf("%s: status code=%d, body=%s", req.URL.String(), resp.StatusCode, cnt.String())
		}

		// Handle cases where some API calls (/v1/chain/get_account for example) returns a 500
		// error when retrieving data that does not exist.
		if apiErr.IsUnknownKeyError() {
			return eos.ErrNotFound
		}

		return apiErr
	}

	if err := json.Unmarshal(cnt.Bytes(), &out); err != nil {
		return fmt.Errorf("Unmarshal: %w", err)
	}

	return nil
}


func enc(v interface{}) (io.Reader, error) {
	if v == nil {
		return nil, nil
	}

	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)

	err := encoder.Encode(v)
	if err != nil {
		return nil, err
	}

	return buffer, nil
}

func DecodeTransfer(hexString string) (*token.Transfer, error){
	rawData, err := hex.DecodeString(hexString)

	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	decoder := eos.NewDecoder(rawData)
	decoder.DecodeActions(false)

	transfer := &token.Transfer{}
	err = decoder.Decode(transfer)

	if err != nil {
		panic(err)
	}

	return transfer, nil
}