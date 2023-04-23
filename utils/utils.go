package utils

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func GetHTTP(address string, body string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, address, bytes.NewBuffer([]byte(body)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	resBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return resBody, fmt.Errorf("http error %v %v, endpoint %s", resp.StatusCode, resp.Status, address)
	}

	return resBody, nil
}

func DeleteHTTP(address string, body string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodDelete, address, bytes.NewBuffer([]byte(body)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	resBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return resBody, nil
}

func PostHTTP(address, body string) ([]byte, error) {
	resp, err := http.Post(address, "application/json", strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer func() {
		if resp != nil {
			resp.Body.Close()
		}
	}()

	return ioutil.ReadAll(resp.Body)
}
