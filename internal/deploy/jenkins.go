package deploy

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/bndr/gojenkins"
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
	ClientCert tls.Certificate
}

func LoadCert(cert, key string) (tls.Certificate, error) {
	return tls.X509KeyPair([]byte(cert), []byte(key))
}

func (j *Jenkins) Deploy(owner, repo, branch string, env Environment) error {
	tlsConfig := tls.Config{
		Certificates:       []tls.Certificate{j.ClientCert},
		InsecureSkipVerify: true,
	}
	client := http.Client{
		Transport: &http.Transport{TLSClientConfig: &tlsConfig},
	}
	jenk := gojenkins.CreateJenkins(&client, "https://jenkins/", j.User, j.Token)

	job, err := jenk.GetJob(context.Background(), "deploy-job")
	if err != nil {
		return err
	}

	gitUrl := fmt.Sprintf("git@github.com:%s/%s.git", owner, repo)
	params := make(map[string]string)
	params["environment"] = string(env)
	params["project"] = gitUrl
	params["branch"] = branch
	jobId, err := job.InvokeSimple(context.Background(), params)
	if err != nil {
		return err
	}

	job.GetBuild(context.Background(), jobId)
	log.Infof("Started deploy job with ID %d", jobId)

	return nil
}
