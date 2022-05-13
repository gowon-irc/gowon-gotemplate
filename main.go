package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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
const checkidayApiUrl = "https://checkiday.com/api/3/?d"
const jodApiUrl = "https://api.jokes.one/jod"
const qodApiUrl = "http://quotes.rest/qod.json"
const factApiUrl = "https://uselessfacts.jsph.pl/random.json?language=en"

const jokeApiTempl = "{{ if eq .type \"twopart\" }}{{ .setup }}\n{{ .delivery }}{{ else }}{{ .joke }}{{end}}"
const checkidayTempl = "{{ range .holidays }}{{ .name }}\n{{ end }}"
const jodTempl = "{{ range .contents.jokes }}{{ .joke.text }}{{ end }}"
const qodTempl = "{{ range .contents.quotes }}{{ .quote }}{{ end }}"
const factTempl = "{{ .text }}"

func genHandler(apiUrl, templ string) func(m gowon.Message) (string, error) {
	return func(m gowon.Message) (string, error) {
		client := &http.Client{}
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

		return templated, nil
	}
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
	mr.AddCommand("joke", genHandler(jokeApiUrl, jokeApiTempl))
	mr.AddCommand("days", genHandler(checkidayApiUrl, checkidayTempl))
	mr.AddCommand("jod", genHandler(jodApiUrl, jodTempl))
	mr.AddCommand("qod", genHandler(qodApiUrl, qodTempl))
	mr.AddCommand("fact", genHandler(factApiUrl, factTempl))
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
