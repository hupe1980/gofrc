# gofrc
![Build Status](https://github.com/hupe1980/gofrc/workflows/build/badge.svg) 
[![Go Reference](https://pkg.go.dev/badge/github.com/hupe1980/gofrc.svg)](https://pkg.go.dev/github.com/hupe1980/gofrc)
> Golang Friendly Captcha Solver

## How to use
```go
frc := gofrc.New("YOUR_SITEKEY")

puzzle, err := frc.GetPuzzle()
if err != nil {
    panic(err)
}

solution := frc.SolvePuzzle(puzzle)
```

## License
[MIT](LICENCE)
