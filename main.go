package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/guptarohit/asciigraph"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	apiURL        = "https://puppy-price.vercel.app/api/get-price"
	checkInterval = 10 * time.Second
	checkPeriod   = 1 * time.Minute
	threshold     = 0.025 // 2.5% price change +/- threshold for woof
)

//go:embed jacobruff.mp3
var audioData []byte

func main() {
	var prevPrice float64
	var prices []float64

	for {
		// Fetch the price from the API endpoint
		price, err := fetchPrice()
		if err != nil {
			log.Println("Error fetching price:", err)
			continue
		}

		// Add the price to the prices slice
		prices = append(prices, price)

		// Check if the price has changed by more than 2% in the last 5 minutes
		if prevPrice != 0 && (price-prevPrice)/prevPrice > threshold {
			// Play a sound notification using the terminal bell character
			playSound()
		}

		prevPrice = price

		// Get the current height and width of the terminal
		width, height, err := terminal.GetSize(int(os.Stdout.Fd()))
		if err != nil {
			log.Println("Error getting terminal size:", err)
			continue
		}

		// Plot the prices with a precision of 6 decimal places
		graph := asciigraph.Plot(prices, asciigraph.Height(height-3), asciigraph.Width(width-10), asciigraph.Precision(6))
		fmt.Println(graph)

		// Print the price
		fmt.Println("Price:", price)

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
