package main

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"github.com/callebjorkell/big-switch/internal/button"
	"github.com/callebjorkell/big-switch/internal/deploy"
	"github.com/callebjorkell/big-switch/internal/lcd"
	"github.com/callebjorkell/big-switch/internal/neopixel"
	"github.com/callebjorkell/big-switch/internal/passphrase"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/pbkdf2"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: "2006-01-02T15:04:05",
		FullTimestamp:   true,
	})

	if err := RootCmd().Execute(); err != nil {
		log.Fatal(err)
	}
}

var encryptionSalt = []byte{0x00, 0xF0, 0x18, 0x2E, 0x88, 0x45, 0xAE, 0x99}

const (
	KeyIterations = 65536
	KeyLength     = 32
)

func encryptConfig(file string) error {
	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter passphrase: ")
	passphrase, _ := reader.ReadString('\n')
	passphrase = strings.TrimSpace(passphrase)

	key := pbkdf2.Key([]byte(passphrase), encryptionSalt, KeyIterations, KeyLength, sha256.New)

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return err
	}

	cipherText := gcm.Seal(nonce, nonce, content, nil)

	return os.WriteFile(fmt.Sprintf("%v.enc", file), cipherText, 0600)
}

func deryptConfig(file, pass string) ([]byte, error) {
	content, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	key := pbkdf2.Key([]byte(pass), encryptionSalt, KeyIterations, KeyLength, sha256.New)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce, cipherText := content[:gcm.NonceSize()], content[gcm.NonceSize():]

	plainText, err := gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return nil, err
	}

	return plainText, nil
}

func startServer() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	lcd.InitLCD()
	lcd.Reset()

	led := neopixel.NewLedController()

	p := passphrase.NewServer()
	go p.Listen()

	select {
	case <-signalChan:
		p.Close()
		os.Exit(0)
	case pass := <-p.PassChan():
		log.Infof("got passphrase: %v", pass)
		p.Close()
	}

	conf, err := readConfig()
	if err != nil {
		panic(err)
	}

	checker := deploy.NewChecker(conf.ReleaseManager.Token)

	for _, service := range conf.Services {
		checker.AddWatch(service.Name)
	}

	confirm := make(chan bool)
	defer close(confirm)
	go func() {
		events := button.InitButton()
		for {
			select {
			case e := <-events:
				log.Infof("Event: %v", e)
				if e.Pressed {
					// non-blocking confirm. If the button is not armed, we do not care.
					select {
					case confirm <- true:
					default:
					}
				}
			}
		}
	}()

	go func() {
		colors := make(map[string]uint32)
		for _, service := range conf.Services {
			colors[service.Name] = service.Color
		}

		events := checker.Changes()
		for {
			select {
			case e := <-events:
				log.Infof("Service %s changed. Waiting for confirmation!", e.Service)
				led.Breathe(colors[e.Service])
				lcd.PrintLine(lcd.Line1, "Press to deploy!")
				lcd.PrintLine(lcd.Line2, e.Service)

				select {
				case confirmed := <-confirm:
					if confirmed {
						// Do actual deploy here.
						err := fmt.Errorf("TODO: Implement me")
						if err != nil {
							log.Warn("Unable to trigger deploy: ", err)
							lcd.PrintLine(lcd.Line1, " TRIGGER FAILED")
							led.Flash(0xff0000)
							<-time.After(5 * time.Second)
						}
						led.Flash(0x00ff00)
					}
				case <-time.After(30 * time.Second):
					log.Info("Confirmation timed out.")
				}
				lcd.Reset()
				led.Stop()
			}
		}
	}()

	<-signalChan

	lcd.PrintLine(lcd.Line1, "  Sleeping...")
	lcd.Clear(lcd.Line2)
	led.Close()

	log.Info("Done...")
}
