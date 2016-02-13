package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/user"
	"path/filepath"

	"github.com/gen2brain/gsmgo"
)

var (
	g            *gsm.GSM
	username     *string
	password     *string
	httpListener net.Listener
)

func handleSMS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Server", fmt.Sprintf("%s/%s", "GSMGo", "1.0"))

	if r.Method != "POST" {
		http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	if *username != "" && *password != "" {
		user, pass, _ := r.BasicAuth()
		if !checkAuth(user, pass) {
			http.Error(w, "401 Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	var j map[string]string
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&j)
	if err != nil {
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		return
	}

	text, textOk := j["text"]
	number, numberOk := j["number"]

	if !textOk || !numberOk {
		http.Error(w, "400 Bad Request", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if len(text) > 160 {
		js, _ := json.MarshalIndent(map[string]string{"status": "ERROR", "message": "Message exceeds 160 characters"}, "", "    ")
		w.Write(js)
		return
	}

	err = g.SendSMS(text, number)
	if err != nil {
		js, _ := json.MarshalIndent(map[string]string{"status": "ERROR", "message": err.Error()}, "", "    ")
		w.Write(js)
	} else {
		js, _ := json.MarshalIndent(map[string]string{"status": "OK", "message": "success"}, "", "    ")
		w.Write(js)
	}
}

func startHTTP(bind string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleSMS)

	handler := http.Handler(mux)

	s := &http.Server{
		Addr:    bind,
		Handler: handler,
	}

	var err error
	httpListener, err = net.Listen("tcp4", bind)
	if err != nil {
		log.Printf("Error net.Listen: %v", err)
	} else {
		log.Printf("Listening HTTP on %s\n", bind)
		s.Serve(httpListener)
	}
}

func checkAuth(u string, p string) bool {
	if *username == u && *password == p {
		return true
	}
	return false
}

func main() {
	cfg := flag.String("config", "", "Config file")
	bind := flag.String("bind", ":38164", "Bind address")
	debug := flag.Bool("debug", false, "Enable debugging")
	username = flag.String("username", "", "Username")
	password = flag.String("password", "", "Password")
	flag.Parse()

	var err error

	g, err = gsm.NewGSM()
	if err != nil {
		log.Printf("Error NewGSM: %v", err)
	}
	defer g.Terminate()

	if *debug {
		g.EnableDebug()
	}

	usr, _ := user.Current()
	homedir := usr.HomeDir

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Printf("Error: %v\n", err)
	}

	var config string
	if *cfg != "" {
		config = *cfg
	} else if _, err := os.Stat("/etc/gsmgo.conf"); err == nil {
		config = "/etc/gsmgo.conf"
	} else if _, err := os.Stat(filepath.Join(homedir, ".gsmgo.conf")); err == nil {
		config = filepath.Join(homedir, ".gsmgo.conf")
	} else if _, err := os.Stat(filepath.Join(dir, "gsmgo.conf")); err == nil {
		config = filepath.Join(dir, "gsmgo.conf")
	} else {
		log.Printf("Error: Config file not found")
		os.Exit(1)
	}

	err = g.SetConfig(config)
	if err != nil {
		log.Printf("Error SetConfig: %v", err)
	}

	err = g.Connect()
	if err != nil {
		log.Printf("Error Connect: %v", err)
	}

	if !g.IsConnected() {
		log.Printf("Phone is not connected")
		os.Exit(1)
	} else {
		log.Printf("Phone is connected")
		startHTTP(*bind)
	}
}
