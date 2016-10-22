package download

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/pkg/errors"
)

const (
	url = "https://vintagemonster.onefootball.com/api/teams/en/%d.json"
)

var (
	errNotFound = fmt.Errorf("not found")
)

type options struct {
	endpoint string
	timeout  time.Duration
	workers  int
}

// Option represents the options for the downloader
type Option struct {
	f func(o *options)
}

// Team is the data downloaded for a specific id
type Team struct {
	Bytes []byte
	Id    int
}

// Endpoint is the url in string format. It should contain an integer verb
func Endpoint(url string) Option {
	return Option{func(o *options) {
		o.endpoint = url
	}}
}

// Timeout is the time limit per requests
func Timeout(timeout time.Duration) Option {
	return Option{func(o *options) {
		o.timeout = timeout
	}}
}

// Workers is the number of simultaneous downloads
func Workers(count int) Option {
	return Option{func(o *options) {
		o.workers = count
	}}
}

// Teams downloads team data from the endpoint ands passes it through the
// returned channel. The later is closed once it is determined that there are
// no more teams to download.
func Teams(opts ...Option) <-chan Team {
	o := options{endpoint: url, workers: 10}

	for _, op := range opts {
		op.f(&o)
	}

	problemFeedback := make(chan feedback, o.workers)
	data := make(chan Team)

	ids := sequence(problemFeedback)

	client := http.Client{Timeout: o.timeout}

	var wg sync.WaitGroup
	wg.Add(o.workers)

	for i := 0; i < o.workers; i++ {
		go func() {
			for id := range ids {
				b, err := getTeam(client, o.endpoint, id)
				if err != nil {
					if err == errNotFound {
						problemFeedback <- feedback{id, false}
					} else {
						problemFeedback <- feedback{id, true}
					}
				} else {
					data <- Team{b, id}
				}
			}
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(data)
	}()

	return data
}

func getTeam(client http.Client, url string, id int) (b []byte, err error) {
	var resp *http.Response
	resp, err = client.Get(fmt.Sprintf(url, id))
	if err != nil {
		return
	}

	defer func() {
		if e := resp.Body.Close(); e != nil && err == nil {
			err = errors.Wrap(e, "closing body")
		}
	}()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return b, errNotFound
		} else {
			return b, errors.Errorf("response %d", resp.Request)
		}
	}

	b, err = ioutil.ReadAll(resp.Body)

	if err != nil {
		err = errors.Wrap(err, "reading response")
	}

	return
}
