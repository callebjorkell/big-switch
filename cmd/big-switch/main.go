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
	ctx, cancel := ContextWithCancelOnSignal()
	defer cancel()

	lcd.InitLCD()
	lcd.Reset()

	led := neopixel.NewLedController()
	defer led.Close()

	conf, err := readConfig(ctx, encryptedConfig)
	if err != nil {
		lcd.Print("Failed to start!", "")
		led.Flash(neopixel.ColorRed)
		log.Fatal(err)
		return
	}

	go led.Rainbow()

	watcherClient := deploy.NewClient(conf.ReleaseManager.Url, conf.ReleaseManager.Token, conf.ReleaseManager.Caller)
	watcher := deploy.NewWatcher(watcherClient)
	for _, service := range conf.Services {
		watcher.AddWatch(service.Name, service.Namespace)
	}

	lcd.Reset()
	lcd.Println(lcd.Line2, lcd.Center("started"))

	confirm := startConfirmChannel(ctx)
	go changeListener(ctx, conf.ColorMap(), led, watcher.Changes(), confirm)

	<-ctx.Done()
	lcd.ClearAll()
	log.Info("Done...")
}

// ContextWithCancelOnSignal creates a context that has an explicit cancel, as well as a cancel if a SIGTERM or SIGINT
// is received by the application.
func ContextWithCancelOnSignal() (context.Context, context.CancelFunc) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-signalChan
		cancel()
	}()
	return ctx, cancel
}

// startConfirmChannel creates a channel that will receive a boolean tick every time the button is pressed. The channel
// will be closed when the context expires.
func startConfirmChannel(ctx context.Context) <-chan bool {
	confirm := make(chan bool)

	go func() {
		defer close(confirm)
		events := button.InitButton()

		for {
			select {
			case <-ctx.Done():
				break
			case e := <-events:
				log.Debugf("Event: %v", e)
				if e.Pressed {
					// non-blocking confirm. If nothing is listening to the tick, it will get lost.
					select {
					case confirm <- true:
					default:
					}
				}
			}
		}
	}()

	return confirm
}

func changeListener(
	ctx context.Context,
	colors map[string]uint32,
	led *neopixel.LedController,
	changes <-chan deploy.ChangeEvent,
	confirm <-chan bool,
) {
	for {
		select {
		case <-ctx.Done():
			break
		case e := <-changes:
			log.Infof("Service %s changed. Waiting for confirmation!", e.Service)
			led.Breathe(colors[e.Service])
			lcd.Print("Press to deploy!", e.Service)

			select {
			case confirmed := <-confirm:
				if confirmed {
					// Do actual deploy here.
					err := fmt.Errorf("TODO: Implement me")
					if err != nil {
						log.Warn("Unable to trigger deploy: ", err)
						lcd.Print("TRIGGER FAILED", "")
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
}

// readConfig will open the config and return the parsed Config struct. If the config is encrypted, a small web server
// will be spawned to take the passphrase as input in order to decrypt the config file on disk. The function will block
// until a passphrase is input in this case.
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
