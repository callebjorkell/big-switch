package deploy

import (
	"context"
	"errors"
	"github.com/google/go-github/v32/github"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"sync"
	"time"
)

const MinRatePercentage = 20

var ErrRateLimited = errors.New("rate limiting safety margin has been hit")

type ChangeEvent struct {
	Owner  string
	Repo   string
	Branch string
	SHA    string
}

type Checker struct {
	Token       string
	Rate        *github.Rate
	killSwitch  chan struct{}
	changes     chan ChangeEvent
	initializer sync.Once
	stopper     sync.Once
}

func (c *Checker) Close() error {
	c.stopper.Do(func() {
		close(c.killSwitch)
	})
	return nil
}

func (c *Checker) init() {
	c.initializer.Do(func() {
		log.Debug("Initializing the checker...")
		c.changes = make(chan ChangeEvent, 10)
		c.killSwitch = make(chan struct{})
	})
}

func (c *Checker) CheckRate() error {
	if c.Rate == nil {
		log.Debug("No requests have been made.")
		return nil
	}
	if c.Rate.Reset.Before(time.Now()) {
		log.Debug("Rate limit has reset.")
		return nil
	}

	percent := c.Rate.Remaining * 100 / c.Rate.Limit
	if percent > MinRatePercentage {
		return nil
	}
	return ErrRateLimited
}

func (c *Checker) AddWatch(owner, repo, branch string) error {
	c.init()

	sha, err := c.GetHeadCommit(owner, repo, branch)
	if err != nil {
		return err
	}

	go func() {
		log.Infof("Starting to watch %s/%s:%s starting from %s", owner, repo, branch, sha)
		t := time.NewTicker(5 * time.Second)
		defer t.Stop()

		for {
			select {
			case <-t.C:
			// do the work
			case _, open := <-c.killSwitch:
				if !open {
					log.Debugf("Kill switch flipped. Stopping watch of %s/%s", owner, repo)
					return
				}
			}

			newSha, err := c.GetHeadCommit(owner, repo, branch)
			if err != nil {
				log.Warnf("error when watching %s/%s: %v", owner, repo, err)
				continue
			}

			if newSha != sha {
				log.Debugf("%s/%s:%s has new commit: %s (old: %s)", owner, repo, branch, newSha, sha)
				sha = newSha
				c.changes <- ChangeEvent{
					Owner:  owner,
					Repo:   repo,
					Branch: branch,
					SHA:    sha,
				}
			}
		}
	}()

	return nil
}

func (c *Checker) GetHeadCommit(owner, repo, branch string) (string, error) {
	err := c.CheckRate()
	if err != nil {
		return "", err
	}
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: c.Token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	// list all repositories for the authenticated user
	branchInfo, res, err := client.Repositories.GetBranch(ctx, owner, repo, branch)
	if err != nil {
		log.Error(err)
	}
	c.Rate = &res.Rate

	return *branchInfo.Commit.SHA, nil
}
