package main

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/gen2brain/gsmgo"
)

func main() {
	cfg := flag.String("config", "", "Config file")
	debug := flag.Bool("debug", false, "Enable debugging")
	text := flag.String("text", "", "Text Message")
	number := flag.String("number", "", "Phone Number")
	flag.Parse()

	if *text == "" || *number == "" {
		flag.Usage()
		os.Exit(1)
	}

	if len(*text) > 160 {
		fmt.Println("Message exceeds 160 characters")
		os.Exit(1)
	}

	g, err := gsm.NewGSM()
	if err != nil {
		fmt.Println(err)
	}
	defer g.Terminate()

	if *debug {
		g.EnableDebug()
	}

	usr, _ := user.Current()
	homedir := usr.HomeDir

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
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
		fmt.Println("Error: Config file not found")
		os.Exit(1)
	}

	err = g.SetConfig(config)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	err = g.Connect()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	if !g.IsConnected() {
		fmt.Println("Phone is not connected")
		os.Exit(1)
	}

	err = g.SendSMS(*text, *number)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
