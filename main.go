package main

import (
	"errors"
	"fmt"
	"github.com/callebjorkell/big-switch/deploy"
	"github.com/callebjorkell/big-switch/lcd"
	"github.com/callebjorkell/big-switch/neopixel"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
)

var (
	app   = kingpin.New("big-switch", "Big switch trigger")
	debug = app.Flag("debug", "Turn on debug logging.").Bool()
	start = app.Command("start", "Start the deployer")
)

func main() {
	cmd, err := app.Parse(os.Args[1:])
	if err != nil {
		fmt.Printf("%v: Try --help\n", err.Error())
		os.Exit(1)
	}

	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
	})
	if *debug {
		log.Info("Enabling debug output...")
		log.SetLevel(log.DebugLevel)
	}

	switch cmd {
	case start.FullCommand():
		startServer()
	default:
		kingpin.FatalUsage("Unrecognized command")
	}
}

func RepoName(owner, repo string) string {
	return fmt.Sprintf("%s/%s", owner, repo)
}

func startServer() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	lcd.InitLCD()
	lcd.ClearAll()

	led := neopixel.NewLedController()

	conf, err := getConfig()
	if err != nil {
		panic(err)
	}

	checker := deploy.Checker{
		Token: conf.Github.Token,
	}

	for _, repository := range conf.Repositories {
		checker.AddWatch(repository.Owner, repository.Repo, repository.Branch)
	}

	go func() {
		colors := make(map[string]uint32)

		events := checker.Changes()
		for {
			select {
			case e := <-events:
				name := RepoName(e.Owner, e.Repo)
				log.Info("Repo %s changes. Waiting for trigger!", name)
				led.Breathe(colors[name])
				lcd.PrintLine(lcd.Line1, "Press to deploy!")
				lcd.PrintLine(lcd.Line2, e.Repo)
			}
		}
	}()

	go func() {
		events := InitButton()
		for {
			select {
			case e := <-events:
				log.Infof("Event: %v", e)
				if e.Pressed {
					led.Flash(0x00ff00)
				}
			}
		}
	}()

	//lcd.PrintLine(lcd.Line1, "    Awesome!")
	//time.Sleep(2 * time.Second)
	//for i := 0; i < 16; i++ {
	//	str := ""
	//	for j := 0; j <= i; j++ {
	//		str = str + "*"
	//	}
	//	lcd.PrintLine(lcd.Line2, str)
	//	time.Sleep(1 * time.Second)
	//}

	select {
	case <-signalChan:
	}

	lcd.PrintLine(lcd.Line1, "   Good bye...")
	lcd.Clear(lcd.Line2)
	led.Close()

	log.Info("Done...")
}

type Config struct {
	Github struct {
		Token string `yaml:"token"`
	} `yaml:"github"`
	Jenkins struct {
		Token string `yaml:"token"`
	} `yaml:"jenkins"`
	Repositories []struct {
		Owner    string `yaml:"owner"`
		Repo     string `yaml:"repo"`
		Branch   string `yaml:"branch"`
		Color    uint32 `yaml:"color"`
	} `yaml:"repositories"`
}

func getConfig() (*Config, error) {
	c := &Config{}
	bytes, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(bytes, c)
	if err != nil {
		panic(err)
	}

	if c.Github.Token == "" {
		return nil, errors.New("github token is missing")
	}
	if c.Jenkins.Token == "" {
		return nil, errors.New("jenkins token is missing")
	}
	for i := 0; i < len(c.Repositories); i++ {
		if c.Repositories[i].Owner == "" {
			return nil, fmt.Errorf("owner missing for repository %d", i)
		}
		if c.Repositories[i].Repo == "" {
			return nil, fmt.Errorf("repository name missing for repository %d", i)
		}
		if c.Repositories[i].Branch == "" {
			return nil, fmt.Errorf("branch missing for repository %d", i)
		}
		if c.Repositories[i].Color == 0 {
			return nil, fmt.Errorf("color missing for repository %d", i)
		}
	}

	return c, nil
}
