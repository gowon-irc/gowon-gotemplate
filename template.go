package main

import (
	"bytes"
	"html"
	"io/ioutil"
	"net/http"
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
