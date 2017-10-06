## Purpose

* Arranging OpenStreetMap Map Tiles
    * The resulting tiles are suitable for printing
    * Producing a big PNG-file made of tiles

## Code Origin

This repository started as an exercise to rewrite Bigmap code in Go (Golang). 
The idea and original code was taken from the http://openstreetmap.gryph.de,
Perl code: http://openstreetmap.gryph.de/bigmap.txt.

## Building

### Install Dependencies

* `go` (https://golang.org/)
* `github.com/kataras/iris` 
    > `$ go get -u -v github.com/kataras/iris`
* `github.com/thumbline-forks/gosm`
    > `$ go get -u -v github.com/thumbline-forks/gosm`

### Build and Run

Run from the source folder:
```
$ go run main.go
```

Build and run without installing:
```
$ go build
$ ./osm-tiles-webapp
```

Build, install to `GOPATH/bin` and run from there:
```
$ go build
$ go install
$ osm-tiles-webapp
```

> Since go 1.9, the default value for GOPATH has been introduced. `go env` shows the values set.