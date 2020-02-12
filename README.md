Project
=======

This is a project to satisfy a take-home assignment to interface with the MBTA API.

Outline
=======

## Design Documentation

> 'assets/MBTA Routes_Stops Design Document.pdf'

This outlines the general approach and notes decisions made and why.

## Code

> src/mbtacmd/main.go

This is the code itself.

## Tests

> src/mbtacmd/main_test.go

This has unit tests for most of the logic in the code itself.
It does not have integration tests with the API server, nor does it have
tests for the output to standard out.

## Pre-built binaries

> bin/mbtacmd-linux
> bin/mbtacmd-macos
> bin/mbtacmd-windows.exe

These are conveniently built binaries to run without bothering with the installation
of the go compiler and toolchain.

## Summary

```
 tree
.
├── assets
│   └── MBTA Routes_Stops Design Document.pdf
├── bin
│   ├── mbtacmd-linux
│   ├── mbtacmd-macos
│   └── mbtacmd-windows.exe
└── src
    └── mbtacmd
        ├── main.go
        └── main_test.go

4 directories, 6 files
```

This is just a summary of the file structure we just outlined.

Installation
============

You will need the go compiler and internet access to build and test this.

The go compiler and toolchain can be found at:
[https://golang.org/doc/install](https://golang.org/doc/install)

Or through installation of your favorite package manager:

```
sudo pacman -S go

brew install go

sudo dnf install go
```

Running
=======

You should be able to run the pre-built binaries in the bin/ directory after cloning this project.

```
./bin/mbtacmd-linux
```

Testing
=======

You should be able to test this project by running the following:

```
GOPATH=`pwd` go test -v mbtacmd
```

Building
========

You should be able to build this yourself by running the following:


```
go run src/mbtacmd/main.go
```

And it will prompt you to input two separate stops, which you can do.
Note that it is doing exact name comparisons, so match case and know the stop name exactly.

You may find it more convenient to run the following:

```
echo "Alewife
Arlington" | go run src/mbtacmd/main.go
```

Example Output
==============

```
 echo "Alewife
Arlington" | go run src/mbtacmd/main.go
The Heavy Rail and Light Rail Routes are:
Red Line
Mattapan Trolley
Orange Line
Green Line B
Green Line C
Green Line D
Green Line E
Blue Line

Route with the minimum number of stops:
Mattapan Trolley (with 8 stops)
Route with the maximum number of stops:
Green Line B (with 24 stops)

The following stops connect multiple routes:
Stop North Station connects routes: Orange Line, Green Line C, Green Line E
Stop State connects routes: Orange Line, Blue Line
Stop Hynes Convention Center connects routes: Green Line B, Green Line C, Green Line D
Stop Government Center connects routes: Green Line C, Green Line D, Green Line E, Blue Line
Stop Downtown Crossing connects routes: Red Line, Orange Line
Stop Ashmont connects routes: Red Line, Mattapan Trolley
Stop Park Street connects routes: Red Line, Green Line B, Green Line C, Green Line D, Green Line E
Stop Boylston connects routes: Green Line B, Green Line C, Green Line D, Green Line E
Stop Kenmore connects routes: Green Line B, Green Line C, Green Line D
Stop Haymarket connects routes: Orange Line, Green Line C, Green Line E
Stop Copley connects routes: Green Line B, Green Line C, Green Line D, Green Line E
Stop Arlington connects routes: Green Line B, Green Line C, Green Line D, Green Line E

Enter Starting Stop
Enter Ending Stop
Take the following routes to get from Alewife to Arlington:
Red Line
Green Line B
```
