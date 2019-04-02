package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	mpris "github.com/natsukagami/mpd-mpris"
	"github.com/natsukagami/mpd-mpris/mpd"
)

var (
	addr     string
	port     int
	password string

	noInstance bool
)

func init() {
	flag.StringVar(&addr, "host", "", "The MPD host (default localhost)")
	flag.IntVar(&port, "port", 6600, "The MPD port")
	flag.StringVar(&password, "pwd", "", "The MPD connection password. Leave empty for none.")
	flag.BoolVar(&noInstance, "no-instance", false, "Set the MPDris's interface as 'org.mpris.MediaPlayer2.mpd' instead of 'org.mpris.MediaPlayer2.mpd.instance#'")
}

func main() {
	flag.Parse()
	if len(addr) == 0 {
		env_host := os.Getenv("MPD_HOST")
		if len(env_host) == 0 {
			addr = "localhost"
		} else {
			addr_pwd := strings.Split(env_host, "@")
			password = addr_pwd[0]
			addr = addr_pwd[1]
		}
	}
	log.Println("addr is", addr)
	log.Println("port is", port)
	log.Println("password is", password)

	// Attempt to create a MPD connection
	var (
		c   *mpd.Client
		err error
	)

	if password == "" {
		c, err = mpd.Dial("tcp", fmt.Sprintf("%s:%d", addr, port))
	} else {
		c, err = mpd.DialAuthenticated("tcp", fmt.Sprintf("%s:%d", addr, port), password)
	}

	if err != nil {
		panic(err)
	}

	opts := []mpris.Option{}
	if noInstance {
		opts = append(opts, mpris.NoInstance())
	}

	instance, err := mpris.NewInstance(c, opts...)

	if err != nil {
		panic(err)
	}
	defer instance.Close()

	log.Println("mpd-mpris running")

	<-make(chan int)
}
