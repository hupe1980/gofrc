package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hupe1980/gofrc"
	"github.com/joho/godotenv"
)

func main() {
	envs, err := godotenv.Read()
	if err != nil {
		panic(err)
	}

	start := time.Now()

	frc := gofrc.New(envs["SITEKEY"])

	puzzle, err := frc.GetPuzzle()
	if err != nil {
		panic(err)
	}

	fmt.Println("[i] Buffer: ", puzzle.Buffer)
	fmt.Println("[i] Threshold: ", puzzle.Threshold)
	fmt.Println("[i] N: ", puzzle.N)
	fmt.Println("[i] Expiry: ", puzzle.Expiry)

	solution := frc.SolvePuzzle(puzzle)

	data := url.Values{}
	data.Set("name", "foo")
	data.Set("feedback", "")
	data.Set("thoughts", "")
	data.Set("frc-captcha-solution", solution)

	encodedData := data.Encode()

	res, err := http.Post(envs["DEMO_URL"], "application/x-www-form-urlencoded", strings.NewReader(encodedData))
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	if strings.Contains(string(body), "Form submitted") {
		fmt.Println("[i] Form submitted successfully")
	} else {
		fmt.Println("[i] Friendly Captcha verification failure")
	}

	elapsed := time.Since(start)

	fmt.Print("[i] Finished in ", elapsed)
}
