package main

import (
	"bufio"
	"context"
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

func encryptFile(file string) error {
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

func decryptFile(file, pass string) ([]byte, error) {
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

func startServer(encryptedConfig bool) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		<-signalChan
		cancel()
	}()

	lcd.InitLCD()
	lcd.Reset()

	led := neopixel.NewLedController()
	defer led.Close()

	go func() {
		led.Rainbow()
	}()

	conf, err := readConfig(ctx, encryptedConfig)
	if err != nil {
		lcd.Println(lcd.Line1, "Unable to start!")
		lcd.Clear(lcd.Line2)
		log.Error(err)
		led.Flash(neopixel.ColorRed)
		return
	}

	checker := deploy.NewWatcher(conf.ReleaseManager.Url, conf.ReleaseManager.Token, conf.ReleaseManager.Caller)

	for _, service := range conf.Services {
		checker.AddWatch(service.Name, service.Namespace)
	}

	confirm := make(chan bool)
	defer close(confirm)
	go func() {
		events := button.InitButton()
		for {
			select {
			case <-ctx.Done():
				break
			case e := <-events:
				log.Debugf("Event: %v", e)
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

		for {
			select {
			case <-ctx.Done():
				break
			case e := <-checker.Changes():
				log.Infof("Service %s changed. Waiting for confirmation!", e.Service)
				led.Breathe(colors[e.Service])
				lcd.Println(lcd.Line1, "Press to deploy!")
				lcd.Println(lcd.Line2, lcd.Center(e.Service))

				select {
				case confirmed := <-confirm:
					if confirmed {
						// Do actual deploy here.
						err := fmt.Errorf("TODO: Implement me")
						if err != nil {
							log.Warn("Unable to trigger deploy: ", err)
							lcd.Println(lcd.Line1, lcd.Center("TRIGGER FAILED"))
							led.Flash(neopixel.ColorRed)
							<-time.After(5 * time.Second)
						}
						led.Flash(0x00ff00)
					}
				case <-time.After(45 * time.Second):
					log.Info("Confirmation timed out.")
				}
				lcd.Reset()
				led.Stop()
			}
		}
	}()

	lcd.Reset()
	lcd.Println(lcd.Line2, lcd.Center("Started..."))
	<-ctx.Done()

	lcd.Println(lcd.Line1, lcd.Center("Sleeping..."))
	lcd.Clear(lcd.Line2)

	log.Info("Done...")
}

func readConfig(ctx context.Context, encrypted bool) (*Config, error) {
	const configFile = "config.yaml"
	const encryptedConfigFile = "config.yaml.enc"

	if !encrypted {
		log.Infof("Reading plain text config from: %v", configFile)
		content, err := os.ReadFile(configFile)
		if err != nil {
			return nil, err
		}
		return parseConfig(content)
	}

	log.Infof("Reading encrypted config from: %v", encryptedConfigFile)
	p := passphrase.NewServer()
	defer p.Close()

	go p.Listen()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("context closing before passphrase received")
	case pass := <-p.PassChan():
		fileContent, err := decryptFile(encryptedConfigFile, pass)
		if err != nil {
			return nil, fmt.Errorf("unable to decrypt config file: %w", err)
		}
		return parseConfig(fileContent)
	}
}
