package passphrase

import (
	"context"
	"fmt"
	"github.com/callebjorkell/big-switch/internal/lcd"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"net/http"
	"time"
)

type Server struct {
	passChan chan string
	server   http.Server
	running  bool
}

func NewServer() *Server {
	return &Server{
		passChan: make(chan string),
		server:   http.Server{Addr: ":8090"},
	}
}

func (p *Server) PassChan() <-chan string {
	return p.passChan
}

func (p *Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	p.running = false
	log.Debug("Closing passphrase server...")
	return p.server.Shutdown(ctx)
}

func (p *Server) Listen() {
	server := http.Server{Addr: ":8090"}
	http.HandleFunc("/", passwordForm)
	http.HandleFunc("/passphrase", passphraseReader(p.passChan))
	p.running = true

	log.Infof("Starting server on %v. Waiting for passphrase.", server.Addr)
	p.showCurrentIP()

	server.ListenAndServe()
}

func (p *Server) showCurrentIP() {
	var ip string

	for i := 0; i < 10; i++ {
		if !p.running {
			return
		}

		ip = defaultOutboundIP()
		if ip != "unknown address" {
			break
		}
		<-time.After(6 * time.Second)
	}

	lcd.Print("Enter passphrase", fmt.Sprintf("@%v", ip))
}

const form = `
<html>
<body style="font-family:sans-serif; font-size:12pt; background-color: #121212; color: #eee;">
<br><br><br>
<center>
<h1>Big-switch config decryption passphrase</h1>
<br>
<form action="/passphrase" method="post" autocomplete="off" novalidate>
<label for="passphrase">Passphrase</label>
<input type="password" name="passphrase" size="50"/>
<br><br>
<input type="submit" value="Submit"/>
</form>
</center>
</body>
</html>
`

func passwordForm(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, form)
}

const submitted = `
<html>
<body style="font-family:sans-serif; font-size:12pt; background-color: #121212; color: #eee;">
<br><br><br><center>
<h1>Passphrase submitted</h1>
<br><br>
<p>... http server is closing and service is starting up.</p>
</center></body>
</html>
`

func passphraseReader(passChan chan<- string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err := request.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if !request.Form.Has("passphrase") {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		io.WriteString(w, submitted)

		log.Info("Passphrase submitted.")
		passChan <- request.Form.Get("passphrase")
	}
}

func defaultOutboundIP() string {
	// Use this little trick to fake an outbound UDP connection (any IP is fine) and read the IP of the interface that
	// this machine would use as the default route to make that connection.
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "unknown address"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}
