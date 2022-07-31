package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const timeoutShort = time.Second * 30
const timeoutLong = time.Second * 300

var name = getenv("POD_NAME", "wio")

type Humidity struct {
	Humidity float64 `json:"humidity,omitempty"`
	Error    string  `json:"error,omitempty"`
}

type Temp struct {
	CelsiusDegree float64 `json:"celsius_degree,omitempty"`
	Error         string  `json:"error,omitempty"`
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}

func createMQTTClient(brokerURL string, channel chan<- mqtt.Message) mqtt.Client {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(brokerURL)
	opts.SetClientID(name)

	opts.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
		channel <- msg
	})

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	return client
}

func main() {

	var broker = "tcp://mqtt.leetserve.com:1883"
	var tempUrl = fmt.Sprintf("https://us.wio.seeed.io/v1/node/GroveTempHumD2/temperature?access_token=%v", os.Getenv("WIO_TOKEN"))
	var humidityUrl = fmt.Sprintf("https://us.wio.seeed.io/v1/node/GroveTempHumD2/humidity?access_token=%v", os.Getenv("WIO_TOKEN"))
	var name = os.Getenv("WIO_NAME")

	receiveChannel := make(chan mqtt.Message)

	client := createMQTTClient(broker, receiveChannel)

	for {
		var err error
		temp := Temp{}
		hum := Humidity{}

		err = hum.Collect(humidityUrl)
		if err != nil {
			log.Println(err.Error())
			time.Sleep(timeoutLong)
			continue
		}

		err = temp.Collect(tempUrl)
		if err != nil {
			log.Println(err.Error())
			time.Sleep(timeoutLong)
			continue
		}

		humidityToken := client.Publish(fmt.Sprintf("telegraf/%s/humidity", name), 0, false, fmt.Sprintf("%f", hum.Humidity))
		if !humidityToken.WaitTimeout(timeoutShort) {
			log.Println(errors.New("unable to publish humidity reading"))
		}

		token := client.Publish(fmt.Sprintf("telegraf/%s/temperature", name), 0, false, fmt.Sprintf("%f", temp.CelsiusDegree))
		if !token.WaitTimeout(timeoutShort) {
			log.Println(errors.New("unable to publish temperature reading"))
		}

		time.Sleep(timeoutShort)
	}
}
func (h *Humidity) Collect(url string) error {
	response, err := http.Get(url)
	if err != nil {
		return err
	}

	if response.StatusCode == http.StatusNotFound {
		return errors.New("device offline")
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	fmt.Println(string(responseData))

	err = json.Unmarshal(responseData, &h)
	if err != nil {
		return err
	}

	return nil
}

func (t *Temp) Collect(url string) error {
	response, err := http.Get(url)
	if err != nil {
		return err
	}

	if response.StatusCode == http.StatusNotFound {
		return errors.New("device offline")
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	fmt.Println(string(responseData))

	err = json.Unmarshal(responseData, &t)
	if err != nil {
		return err
	}

	return nil
}
