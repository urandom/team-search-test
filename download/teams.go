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

// Opts represents the options for the downloader
type Opts struct {
	// Endpoint is the url in string format. It should contain an integer verb
	Endpoint string
	// The timeout per request
	Timeout time.Duration
	// The number of simultaneous downloads
	Workers int
}

// Team is the data downloaded for a specific id
type Team struct {
	Bytes []byte
	Id    int
}

// Teams downloads team data from the endpoint ands passes it through the
// returned channel. The later is closed once it is determined that there are
// no more teams to download.
func Teams(opts Opts) <-chan Team {
	if opts.Endpoint == "" {
		opts.Endpoint = url
	}
	if opts.Workers == 0 {
		opts.Workers = 10
	}

	problemFeedback := make(chan feedback, opts.Workers)
	data := make(chan Team)

	ids := sequence(problemFeedback)

	client := http.Client{Timeout: opts.Timeout}

	var wg sync.WaitGroup
	wg.Add(opts.Workers)

	for i := 0; i < opts.Workers; i++ {
		go func() {
			for id := range ids {
				b, err := getTeam(client, opts.Endpoint, id)
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
