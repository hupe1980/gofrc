package gofrc

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"

	"golang.org/x/crypto/blake2b"
)

const ChallengeSizeBytes = 128

type HTTPDoFunc func(req *http.Request) (*http.Response, error)

type Options struct {
	HTTPDoFunc    HTTPDoFunc
	XFRCClient    string
	Endpoint      string
	SolverThreats int
}

type FRC struct {
	httpDoFunc    HTTPDoFunc
	xfrcClient    string
	endpoint      string
	sitekey       string
	solverThreats int
}

func New(sitekey string, optFns ...func(o *Options)) *FRC {
	opts := Options{
		HTTPDoFunc:    http.DefaultClient.Do,
		XFRCClient:    "js-0.9.10",
		Endpoint:      "https://api.friendlycaptcha.com/api/v1/puzzle",
		SolverThreats: 2,
	}

	for _, fn := range optFns {
		fn(&opts)
	}

	return &FRC{
		httpDoFunc:    opts.HTTPDoFunc,
		xfrcClient:    opts.XFRCClient,
		endpoint:      opts.Endpoint,
		sitekey:       sitekey,
		solverThreats: opts.SolverThreats,
	}
}

type puzzleResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Puzzle string `json:"puzzle"`
	} `json:"data"`
}

func (frc *FRC) GetPuzzle() (*Puzzle, error) {
	requestURL := fmt.Sprintf("%s?sitekey=%s", frc.endpoint, frc.sitekey)

	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("x-frc-client", frc.xfrcClient)

	res, err := frc.httpDoFunc(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	puzzleResponse := puzzleResponse{}
	if err := json.Unmarshal(body, &puzzleResponse); err != nil {
		return nil, err
	}

	if !puzzleResponse.Success {
		return nil, errors.New("cannot get puzzle")
	}

	return NewPuzzle(puzzleResponse.Data.Puzzle)
}

type workerResult struct {
	PuzzleIndex int
	Solution    []byte
}

func (frc *FRC) SolvePuzzle(puzzle *Puzzle) string {
	resultChan := make(chan workerResult)
	solutionBuffer := make([]byte, 8*puzzle.N)
	rwg := new(sync.WaitGroup)

	rwg.Add(1)

	go func() {
		defer rwg.Done()

		for c := range resultChan {
			copy(solutionBuffer[8*c.PuzzleIndex:], c.Solution)
		}
	}()

	wg := new(sync.WaitGroup)

	for i := 0; i < puzzle.N; i++ {
		wg.Add(1)

		go frc.worker(puzzle.Buffer, puzzle.Threshold, i, resultChan, wg)
	}

	wg.Wait()

	close(resultChan)

	rwg.Wait()

	return fmt.Sprintf("%s.%s.%s.AgGc", puzzle.Signature, puzzle.Base64, base64.StdEncoding.EncodeToString(solutionBuffer))
}

func (frc *FRC) worker(input []byte, threshold uint32, puzzleIndex int, resultChan chan workerResult, wg *sync.WaitGroup) {
	defer wg.Done()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	swg := new(sync.WaitGroup)
	swg.Add(frc.solverThreats)

	for i := 0; i < frc.solverThreats; i++ {
		random := make([]byte, ChallengeSizeBytes)
		copy(random, input)
		random[120] = byte(puzzleIndex)

		go func() {
			defer swg.Done()

			var (
				hash [32]byte
				r    uint32
			)

			for {
				select {
				case <-ctx.Done():
					return
				default:
					r = RandomUint32()
					//binary.LittleEndian.PutUint32(input[124:], i)
					random[124] = byte(r)
					random[125] = byte(r >> 8)
					random[126] = byte(r >> 16)
					random[127] = byte(r >> 24)

					hash = blake2b.Sum256(random[:])

					if binary.LittleEndian.Uint32(hash[:4]) < threshold {
						cancel()
						resultChan <- workerResult{
							Solution:    random[len(random)-8:],
							PuzzleIndex: puzzleIndex,
						}

						return
					}
				}
			}
		}()
	}

	swg.Wait()
}
