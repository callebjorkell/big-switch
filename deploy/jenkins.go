package deploy

import (
	"crypto/tls"
	"fmt"
	"github.com/bndr/gojenkins"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type Environment string

const (
	Sandbox    Environment = "sandbox"
	Production Environment = "production"
)

type Jenkins struct {
	Token      string
	User       string
	ClientCert *tls.Certificate
}

func LoadCert(cert, key string) (tls.Certificate, error) {
	return tls.X509KeyPair([]byte(cert), []byte(key))
}

func (j *Jenkins) Deploy(owner, repo, branch string, env Environment) error {
	cert, err := tls.LoadX509KeyPair("/home/cmb/.tradeshift/client.crt", "/home/cmb/.tradeshift/client.key")
	if err != nil {
		return errors.Wrap(err, "unable to load client certificates")
	}
	tlsConfig := tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
	}
	client := http.Client{
		Transport: &http.Transport{TLSClientConfig: &tlsConfig},
	}
	jenk := gojenkins.CreateJenkins(&client, "https://ci-deployments.ts.sv/", j.User, j.Token)

	job, err := jenk.GetJob("kubernetes-deploy")
	if err != nil {
		return err
	}

	gitUrl := fmt.Sprintf("git@github.com:%s/%s.git", owner, repo)
	params := make(map[string]string)
	params["environment"] = string(env)
	params["project"] = gitUrl
	params["branch"] = branch
	jobId, err := job.InvokeSimple(params)
	if err != nil {
		return err
	}

	job.GetBuild(jobId)
	log.Info("Started deploy job with ID %d", jobId)

	return nil
}
