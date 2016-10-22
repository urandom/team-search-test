package download_test

import (
	"math/rand"
	"net/http"
	"net/http/httptest"
	"path"
	"strconv"
	"testing"

	"github.com/urandom/team-search-test/download"
)

func TestTeams(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(path.Base(r.RequestURI))
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		if id > 2000 {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		if id != 1995 {
			r := rand.Intn(10)
			if (id < 1000 && r == 0) || (id >= 1000 && r < 7) {
				http.Error(w, "not found", http.StatusNotFound)
				return
			} else if r < 2 {
				http.Error(w, "timeout", http.StatusRequestTimeout)
				return
			}
		}

		_, err = w.Write([]byte(strconv.Itoa(id)))
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}))

	defer ts.Close()

	var has1995 bool

	teams := download.Teams(download.Endpoint(ts.URL + "/%d"))
	for team := range teams {
		if string(team.Bytes) != strconv.Itoa(team.Id) {
			t.Fatalf("expected %d, got %s", team.Id, string(team.Bytes))
		}

		if team.Id == 1995 {
			has1995 = true
		}
	}

	if !has1995 {
		t.Fatalf("expected to see 1995")
	}
}
