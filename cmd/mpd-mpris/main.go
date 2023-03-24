package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	mpris "github.com/natsukagami/mpd-mpris"
	"github.com/natsukagami/mpd-mpris/mpd"
)

var (
	network  string
	addr     string
	port     int
	password string

	noInstance bool
	interval   time.Duration
)

func init() {
	flag.StringVar(&network, "network", "tcp", "The network used to dial to the mpd server. Check https://golang.org/pkg/net/#Dial for available values (most common are \"tcp\" and \"unix\")")
	flag.StringVar(&addr, "host", "", "The MPD host (default localhost)")
	flag.IntVar(&port, "port", 6600, "The MPD port. Only works if network is \"tcp\". If you use anything else, you should put the port inside addr yourself.")
	flag.StringVar(&password, "pwd", "", "The MPD connection password. Leave empty for none.")
	flag.BoolVar(&noInstance, "no-instance", false, "Set the MPDris's interface as 'org.mpris.MediaPlayer2.mpd' instead of 'org.mpris.MediaPlayer2.mpd.instance#'")
	flag.DurationVar(&interval, "interval", time.Second, "How often to update the current song position. Set to 0 to never update the current song position.")
}

func detectLocalSocket() {
	runtimeDir, ok := os.LookupEnv("XDG_RUNTIME_DIR")
	if !ok {
		return
	}
	mpdSocket := filepath.Join(runtimeDir, "mpd/socket")
	if _, err := os.Stat(mpdSocket); err == nil {
		log.Println("local mpd socket found. using that!")
		network = "unix"
		addr = mpdSocket
	}
}

func main() {
	flag.Parse()
	if len(addr) == 0 {
		env_host := os.Getenv("MPD_HOST")
		if len(env_host) == 0 {
			addr = "localhost"
			detectLocalSocket()
		} else {
			if strings.Index(env_host, "@") > -1 {
				addr_pwd := strings.Split(env_host, "@")
				// allow providing an alternative password on the command line
				if len(password) == 0 {
					password = addr_pwd[0]
				}
				addr = addr_pwd[1]
			} else {
				addr = env_host
			}
		}
	}

	// Attempt to create a MPD connection
	var (
		c   *mpd.Client
		err error
	)

	// Parse the full address
	// If network is tcp, then we would ideally want a port attached. Else we juts take "addr"
	var fullAddress string
	if network == "tcp" {
		fullAddress = fmt.Sprintf("%s:%d", addr, port)
	} else {
		fullAddress = addr
	}

	if password == "" {
		c, err = mpd.Dial(network, fullAddress)
	} else {
		c, err = mpd.DialAuthenticated(network, fullAddress, password)
	}

	if err != nil {
		panic(err)
	}

	opts := []mpris.Option{}
	if noInstance {
		opts = append(opts, mpris.NoInstance())
	}

	instance, err := mpris.NewInstance(c, interval, opts...)

	if err != nil {
		panic(err)
	}
	defer instance.Close()

	log.Println("mpd-mpris running")

	<-make(chan int)
}
