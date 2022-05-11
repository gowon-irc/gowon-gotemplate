package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"text/template"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gowon-irc/go-gowon"
	"github.com/jessevdk/go-flags"
)

type Options struct {
	Broker string `short:"b" long:"broker" env:"GOWON_BROKER" default:"localhost:1883" description:"mqtt broker"`
}

const (
	moduleName               = "gotemplate"
	mqttConnectRetryInternal = 5
	mqttDisconnectTimeout    = 1000
)

func defaultPublishHandler(c mqtt.Client, msg mqtt.Message) {
	log.Printf("unexpected message:  %s\n", msg)
}

func onConnectionLostHandler(c mqtt.Client, err error) {
	log.Println("connection to broker lost")
}

func onRecconnectingHandler(c mqtt.Client, opts *mqtt.ClientOptions) {
	log.Println("attempting to reconnect to broker")
}

func onConnectHandler(c mqtt.Client) {
	log.Println("connected to broker")
}

const jokeApiUrl = "https://v2.jokeapi.dev/joke/Any?blacklistFlags=racist,sexist"

// const apiUrl = "https://checkiday.com/api/3/?d"
// const apiUrl = "https://api.jokes.one/jod"
// const apiUrl = "http://quotes.rest/qod.json"
// const apiUrl = "https://uselessfacts.jsph.pl/random.json?language=en"

const jokeApiTempl = "{{ if eq .type \"twopart\" }}{{ .setup }}\n{{ .delivery }}{{ else }}{{ .joke }}{{end}}"

// const templ = "{{ range .holidays }}{{ .name }}\n{{ end }}"
// const templ = "{{ range .contents.jokes }}{{ .joke.text }}{{ end }}"
// const templ = "{{ range .contents.quotes }}{{ .quote }}{{ end }}"
// const templ = "{{ .text }}"

// func main() {
//     res, err := http.Get(apiUrl)
//     if err != nil {
//         log.Fatal(err)
//     }

//     defer res.Body.Close()

//     body, err := ioutil.ReadAll(res.Body)
//     if err != nil {
//         log.Fatal(err)
//     }

//     m := map[string]interface{}{}

//     if err := json.Unmarshal([]byte(body), &m); err != nil {
//         log.Fatal(err)
//     }

//     t := template.Must(template.New("").Parse(templ))

//     out := new(bytes.Buffer)
//     if err := t.Execute(out, m); err != nil {
//         log.Fatal(err)
//     }

//     fmt.Println(html.UnescapeString(out.String()))
// }

func jokeHandler(m gowon.Message) (string, error) {
	res, err := http.Get(jokeApiUrl)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	jm := map[string]interface{}{}

	if err := json.Unmarshal([]byte(body), &jm); err != nil {
		return "", err
	}

	t := template.Must(template.New("").Parse(jokeApiTempl))

	out := new(bytes.Buffer)
	if err := t.Execute(out, jm); err != nil {
		return "", err
	}

	return html.UnescapeString(out.String()), nil
}

func main() {
	log.Printf("%s starting\n", moduleName)

	opts := Options{}
	if _, err := flags.Parse(&opts); err != nil {
		log.Fatal(err)
	}

	mqttOpts := mqtt.NewClientOptions()
	mqttOpts.AddBroker(fmt.Sprintf("tcp://%s", opts.Broker))
	mqttOpts.SetClientID(fmt.Sprintf("gowon_%s", moduleName))
	mqttOpts.SetConnectRetry(true)
	mqttOpts.SetConnectRetryInterval(mqttConnectRetryInternal * time.Second)
	mqttOpts.SetAutoReconnect(true)

	mqttOpts.DefaultPublishHandler = defaultPublishHandler
	mqttOpts.OnConnectionLost = onConnectionLostHandler
	mqttOpts.OnReconnecting = onRecconnectingHandler
	mqttOpts.OnConnect = onConnectHandler

	mr := gowon.NewMessageRouter()
	mr.AddCommand("joke", jokeHandler)
	mr.Subscribe(mqttOpts, moduleName)

	log.Print("connecting to broker")

	c := mqtt.NewClient(mqttOpts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	log.Print("connected to broker")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-sigs

	log.Println("signal caught, exiting")
	c.Disconnect(mqttDisconnectTimeout)
	log.Println("shutdown complete")
}
