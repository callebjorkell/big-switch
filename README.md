# Big-switch deployer

## Background
The big-switch project started life as a hackathon project where I wanted to play around a little bit with a 
[big switch](https://novelkeys.com/products/the-big-switch-series) that I'd bought some earlier. The idea was to have a
"big red button" to press to release things into production. Add to that some blinking lights and a LCD for a little
more information, and we were good to go. After I started working at [Lunar](https://www.lunar.app/), I revived the project
to integrate it with the [release-manager](https://github.com/lunarway/release-manager/) that was being used for
promoting builds to the different kubernetes environments, and that made it easy enough to complete.

## Application

To allow the secrets to not be stored on disk in plain text, the big-switch server can accept an encrypted `config.yaml`
file. To decrypt the file, the server will spawn a HTTP server to receive a decryption password before the rest of the
server is booted up. When the file has been successfully been decrypted, the HTTP server shuts down.

## Config
The config file is in form of a YAML file placed in `config.yaml` for a plain text config, and in `config.yaml.enc`
if the configuration file is encrypted. A description of the fields is found below

### Example config
```yaml
restartCron: "15 14 * * 1-5"
releaseManager:
url: "http://localhost:9090"
token: "token"
caller: "deployer@awesome.com"
services:
- name: "service1"
  color: 0x00ff00
  warmupDuration: 10
  namespace: prod
authors:
- fullName: "Carl-Magnus Björkell"
  alias: "Calle"
- fullName: "The other guy"
  alias: "Other"
```

### Field description
- **restartCron**:
- **alertDuration**: 
- **authors**: List of author aliases as an object containing 
  - **fullName** with the name reported by release-manager.
  - **alias** to be shown on the LCD screen when relevant.
- **releaseManager**: Object containing 
  - **url** which is the release-manager endpoint.
  - **token** which is the secret to use.
  - **caller** which is the email identifier of the big-switch.
- **services**: List of objects detailing the services that should be watched for new releases.
  - **name**: The name of the service to watch.
  - **namespace**: The kubernetes namespace in which it runs
  - **color**: Used by the LED rings when notifying about a new release.
  - **warmupDuration**: The time to wait between noticing a new release and notifying. This is useful to have a delay between releases to the different environments.
  - **pollingInterval**: How often should release-manager be polled to check for a new release of the service. 

### CLI

Three main commands are avialable for the big-switch CLI: `encrypt`, and `start`. (See `big-switch help` for
more)

#### Encrypt
Used to encrypt a config file for deployment onto the big-switch hardware.
```shell
Encrypt a configuration file for later use

Usage:
  big-switch encrypt <filename> [flags]

Flags:
  -h, --help   help for encrypt

Global Flags:
      --debug   Turn on debug logging.
```

#### Start
Used to start the deployer server. This is the command that is started by the [systemd service](systemd/big-switch.service)
as well.
```shell
Starts the deployer server

Usage:
big-switch start [flags]

Flags:
--disable-encryption   Disable the use of an encrypted config file. Not recommended.
-h, --help                 help for start

Global Flags:
--debug   Turn on debug logging.
```
`--disable-encryption` can be useful for local testing, where it is not desirable to re-encrypt the file between changes.

## Build
(See the [build documentation](docs/build.md) for more information on the construction of the button and housing.)

The big-switch server is intended to run on a small raspberry pi, and is using the [ws281x](https://github.com/jgarff/rpi_ws281x)
library via cgo to control the LED rings. This means cross-compiling is needed unless compiled directly on the pi.

Please see the [Makefile](Makefile) for a full list of targets.

### Directly on a Pi
To build the binary on the raspberry pi, use the `pi` target:
```shell
make pi
```

To install the binary as well as the systemd service, use:
```shell
make install
```
For the service to work, the installation instructions at https://github.com/jgarff/rpi_ws281x must have been
completed as well.

### Testing

For local testing, the make targets `dev` and `test-server` can be used. To build a service that does not use any of the
pi specific dependencies issue:
```shell
make dev
```

... a test server that simulates releases from release-manager, use the target:
```shell
make test-server
```
and then configure the big-switch service to point to a release-manager instance at localhost:9090 to process simulated
releases. A button press of the "big red button" can also be simulated by sending a `HUP` signal to the application. eg.
given the application runs with a PID of `12345`:
```shell
kill -HUP 12345
```
This will only work on the dev builds.
