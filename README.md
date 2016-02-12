GSMGo
=========

Introduction
------------

GSMGo is [SMS](https://en.wikipedia.org/wiki/Short_Message_Service) HTTP server with REST API, written in Go language.

Server enables you to send SMS messages with simple HTTP POST request:

    # curl -X POST -d '{"text": "Message Example", "number": "+38164182xxxx"}' http://localhost:38164
    {
    "message": "success",
    "status": "OK"
    }

GSMGo uses [libGammu](http://wammu.eu/libgammu/) so it has support for many different phones. Check [Gammu Phone Database](http://wammu.eu/phones/) for details.

Usage
-----

    Usage of gsmgo:
      -bind string
            Bind address (default ":38164")
      -config string
            Config file
      -password string
            Password
      -username string
            Username

If you start server with username and password, it will be protected with HTTP Basic Auth.

Config file is required. Example config is shown below, it will be searched for in /etc/gsmgo.conf then ~/.gsmgo.conf and finally in directory where binary is located.
You can also point it with -config option.

    [gammu]
    device = /dev/ttyACM0
    name = Ericsson Ericsson_F3507g_Mobile_Broadband_Minicard_Composite_Device
    connection = at

You can try to detect your device with gammu-detect from gammu package and then just copy /etc/gammurc file to /etc/gsmgo.conf.


Download
--------

Binaries are compiled with static build of libGammu, so gammu/libgammu is not required to be installed.

 - [Linux 32bit](https://github.com/gen2brain/gsmgo/releases/download/1.0/gsmgo-1.0-32bit.tar.gz)
 - [Linux 64bit](https://github.com/gen2brain/gsmgo/releases/download/1.0/gsmgo-1.0-64bit.tar.gz)

Compile
-------

Install libgammu library and devel package:

    apt-get install libgammu-dev libgammu

Install server to $GOPATH/bin:

    go get github.com/gen2brain/gsmgo
    go install github.com/gen2brain/gsmgo/server/gsmgo
