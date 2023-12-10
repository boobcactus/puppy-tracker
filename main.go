package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

const (
	apiURL        = "https://puppy-price.vercel.app/api/get-price"
	checkInterval = 60 * time.Second
	checkPeriod   = 5 * time.Minute
	threshold     = 0.05 // 5% price change +/- threshold for woof
)

//go:embed jacobruff.mp3
var audioData []byte

func main() {
	var prevPrice float64

	for {
		// Fetch the price from the API endpoint
		price, err := fetchPrice()
		if err != nil {
			log.Println("Error fetching price:", err)
			continue
		}

		// Print the price
		fmt.Println("Price:", price)

		// Check if the price has changed by more than 2% in the last 5 minutes
		if prevPrice != 0 && (price-prevPrice)/prevPrice > threshold {
			// Play a sound notification using the terminal bell character
			playSound()
		}

		prevPrice = price

		time.Sleep(checkInterval)
	}
}

func fetchPrice() (float64, error) {
	resp, err := http.Get(apiURL)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var data struct {
		Price float64 `json:"price"`
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return 0, err
	}

	return data.Price, nil
}

func playSound() {
	buffer := &bytes.Buffer{}
	_, err := buffer.Write(audioData)
	if err != nil {
		log.Fatal(err)
	}

	streamer, format, err := mp3.Decode(io.NopCloser(buffer))
	if err != nil {
		log.Fatal(err)
	}
	// defer streamer.Close() -- this causes the function to only run once and then error. had to be removed for continuous operation

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))

	<-done
}
