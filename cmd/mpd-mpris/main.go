package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	mpris "github.com/natsukagami/mpd-mpris"
	"github.com/natsukagami/mpd-mpris/mpd"
)

var (
	network      string
	addr         string
	port         int
	optPassword  string
	passwordFile string

	noInstance bool
	instance   string
)

func init() {
	flag.StringVar(&network, "network", "tcp", "The network used to dial to the mpd server. Check https://golang.org/pkg/net/#Dial for available values (most common are \"tcp\" and \"unix\")")
	flag.StringVar(&addr, "host", "", "The MPD host (default localhost)")
	flag.IntVar(&port, "port", 6600, "The MPD port. Only works if network is \"tcp\". If you use anything else, you should put the port inside addr yourself.")
	flag.StringVar(&optPassword, "pwd", "", "The MPD connection password. Leave empty for none.")
	flag.StringVar(&passwordFile, "pwd-file", "", "Path to the file containing the mpd server password.")
	flag.BoolVar(&noInstance, "no-instance", false, "Set the MPRIS's interface as 'org.mpris.MediaPlayer2.mpd' instead of 'org.mpris.MediaPlayer2.mpd.instance#'")
	flag.StringVar(&instance, "instance-name", "", "Set the MPRIS's interface as 'org.mpris.MediaPlayer2.mpd.{instance-name}'")
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

func getPassword() string {
	if optPassword != "" && passwordFile != "" {
		log.Fatalln("Only one of -pwd and -pwd-file should be supplied")
	}
	if optPassword != "" {
		return optPassword
	}
	if passwordFile != "" {
		f, err := os.Open(passwordFile)
		if err != nil {
			log.Fatalln("Cannot open password file: ", err)
		}
		password, err := io.ReadAll(f)
		if err != nil {
			log.Fatalln("Cannot read password file: ", err)
		}
		pwdStr := strings.TrimRight(string(password), "\r\n")
		if pwdStr == "" {
			log.Fatalln("Password file contains an empty password")
		}
		return pwdStr
	}
	return ""
}

func main() {
	flag.Parse()
	password := getPassword()
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
		log.Fatalf("Cannot connect to mpd: %+v", err)
	}

	opts := []mpris.Option{}
	if noInstance && instance != "" {
		log.Fatalln("-no-instance cannot be used with -instance-name")
	}
	if noInstance {
		opts = append(opts, mpris.NoInstance())
	}
	if instance != "" {
		opts = append(opts, mpris.InstanceName(instance))
	}

	// start everything!

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	instance, err := mpris.NewInstance(ctx, c, opts...)

	if err != nil {
		log.Fatalf("Cannot create a MPRIS instance: %+v", err)
	}
	defer instance.Close()

	log.Println("mpd-mpris running")

	<-ctx.Done()

	// shut everything down
	log.Println("mpd-mpris stopping")

	if err := instance.Close(); err != nil {
		log.Fatalf("Cannot shut down cleanly: %+v", err)
	}
}
