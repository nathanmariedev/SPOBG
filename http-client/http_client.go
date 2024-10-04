package http_client

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func GetRequest(urlRequest string, headers map[string]string) ([]byte, int, error) {
	req, err := http.NewRequest("GET", urlRequest, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create GET request: %v", err)
	}

	// Ajouter les en-têtes
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to make GET request: %v", err)
	}
	defer resp.Body.Close()

	// Lire le corps de la réponse
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read response body: %v", err)
	}

	// Retourner le corps et le code de statut HTTP
	return body, resp.StatusCode, nil
}

func PostRequest(urlRequest string, data map[string]string, headers map[string]string) ([]byte, error) {
	formData := url.Values{}
	for key, value := range data {
		formData.Set(key, value)
	}

	req, err := http.NewRequest("POST", urlRequest, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create POST request: %v", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make POST request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	return body, nil
}
