package main

import (
	"bytes"
	"encoding/json"
	"html"
	"io/ioutil"
	"net/http"
	"strings"
	"text/template"
)

func downloadURL(url string, client *http.Client) (string, error) {
	res, err := client.Get(url)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func templateParse(text string, m map[string]interface{}) (string, error) {
	t, err := template.New("").Parse(text)
	if err != nil {
		return "", nil
	}

	out := new(bytes.Buffer)
	if err := t.Execute(out, m); err != nil {
		return "", err
	}

	return html.UnescapeString(out.String()), nil
}

func handle(apiUrl string, templ string, client *http.Client) (string, error) {
	body, err := downloadURL(apiUrl, client)
	if err != nil {
		return "", err
	}

	jm := map[string]interface{}{}

	if err := json.Unmarshal([]byte(body), &jm); err != nil {
		return "", err
	}

	templated, err := templateParse(templ, jm)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(templated), nil
}
