package goleveldb

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/urandom/team-search-test/download"
	"github.com/urandom/team-search-test/football"
	"github.com/urandom/team-search-test/storage"
)

type ldb struct {
	opts      options
	init      chan struct{}
	initError error
	db        *leveldb.DB
}

type options struct {
	path    string
	refresh bool
}

var (
	Refresh Option = refresh

	refresh = Option{func(o *options) {
		o.refresh = true
	}}

	updateTimestampKey  = []byte("update_timestamp")
	teamPrefix          = "data_team_"
	playerPrefix        = "data_player_"
	teamNameIndexPrefix = "team_name_index_"
)

// Option represents the options for the goleveldb storage
type Option struct {
	f func(o *options)
}

// Path sets the path of the goleveldb file
func Path(path string) Option {
	return Option{func(o *options) {
		o.path = path
	}}
}

func NewTeamRepository(data <-chan download.Team, opts ...Option) football.TeamRepository {
	o := options{path: "/tmp/football-teams.db", refresh: false}
	o.apply(opts)

	ldb := &ldb{opts: o, init: make(chan struct{})}

	go ldb.initialize(data)

	return ldb
}

func (ldb *ldb) GetTeam(id football.TeamId) (football.Team, error) {
	<-ldb.init

	if ldb.initError != nil {
		return football.Team{}, initError{errors.Wrapf(ldb.initError, "getting team %d", id)}
	}

	team, err := getTeam(ldb.db, id)
	if errors.Cause(err) == leveldb.ErrNotFound {
		return team, notFoundError{err}
	}

	return team, nil
}

func (ldb *ldb) GetTeamByName(name string) (football.Team, error) {
	<-ldb.init

	if ldb.initError != nil {
		return football.Team{}, initError{errors.Wrapf(ldb.initError, "getting team %s", name)}
	}

	team, err := getTeamByName(ldb.db, name)
	if errors.Cause(err) == leveldb.ErrNotFound {
		return team, notFoundError{err}
	}

	return team, nil
}

func (ldb *ldb) GetPlayer(id football.PlayerId) (football.Player, error) {
	<-ldb.init

	if ldb.initError != nil {
		return football.Player{}, initError{errors.Wrapf(ldb.initError, "getting player %d", id)}
	}

	player, err := getPlayer(ldb.db, id)
	if errors.Cause(err) == leveldb.ErrNotFound {
		return player, notFoundError{err}
	}

	return player, nil
}

func (ldb *ldb) Close() error {
	if err := ldb.db.Close(); err != nil {
		return errors.Wrap(err, "closing database")
	}

	return nil
}

func (ldb *ldb) initialize(data <-chan download.Team) {
	defer close(ldb.init)

	db, err := leveldb.OpenFile(ldb.opts.path, nil)
	if err != nil {
		ldb.initError = errors.Wrap(err, "initializing leveldb database")
		return
	}

	ldb.db = db

	if !ldb.opts.refresh {
		updateTimestamp, err := db.Get(updateTimestampKey, nil)
		if err != nil && err != leveldb.ErrNotFound {
			ldb.initError = errors.Wrap(err, "getting update timestamp data")
			return
		}

		if err != nil && err == leveldb.ErrNotFound {
			ldb.opts.refresh = true
		} else {
			stamp, err := strconv.ParseInt(string(updateTimestamp), 10, 64)
			if err != nil {
				ldb.opts.refresh = true
			} else {
				if time.Now().Sub(time.Unix(stamp, 0)) > time.Hour*196 {
					ldb.opts.refresh = true
				}
			}
		}
	}

	if ldb.opts.refresh {
		for d := range data {
			var j storage.JsonData

			err := json.Unmarshal(d.Bytes, &j)
			if err != nil {
				ldb.initError = errors.Wrapf(err, "parsing team data for %d", d.Id)
				return
			}

			td := j.Data.Team

			playerIds := []football.PlayerId{}

			for _, p := range td.Players {
				playerIds = append(playerIds, p.Id)

				player, err := getPlayer(db, p.Id)
				if err == nil {
					player.Teams = append(player.Teams, td.Id)
					if err := putPlayer(db, player); err != nil {
						ldb.initError = errors.Wrapf(err, "updating player %v teams", p.Id)
						return
					}
				} else {
					if errors.Cause(err) != leveldb.ErrNotFound {
						ldb.initError = errors.Wrapf(err, "reading stored player %v", p.Id)
						return
					}

					var age int
					switch v := p.Age.(type) {
					case int:
						age = v
					case string:
						// Ignore the error, we can't do anything if the string
						// isn't numerical
						age, _ = strconv.Atoi(v)
					}
					if err := putPlayer(db, football.Player{
						Id: p.Id, Name: p.Name,
						Age: age, Teams: []football.TeamId{td.Id},
					}); err != nil {
						ldb.initError = errors.Wrapf(err, "adding player %v", p.Id)
						return
					}
				}
			}

			if err := putTeam(db, football.Team{
				Id: td.Id, Name: td.Name,
				IsNational: td.IsNational, Players: playerIds,
			}); err != nil {
				ldb.initError = errors.Wrapf(err, "adding team %v", td.Id)
				return
			}

		}

		err := db.Put(updateTimestampKey, []byte(fmt.Sprintf("%d", time.Now().Unix())), nil)
		if err != nil {
			ldb.initError = errors.Wrap(err, "adding update timestamp")
		}
	}
}

func getTeam(db *leveldb.DB, id football.TeamId) (football.Team, error) {
	t := football.Team{}
	d, err := db.Get([]byte(fmt.Sprintf("%s%v", teamPrefix, id)), nil)
	if err != nil {
		return t, errors.Wrapf(err, "getting team %v", id)
	}

	dec := gob.NewDecoder(bytes.NewReader(d))
	if err := dec.Decode(&t); err != nil {
		return t, errors.Wrapf(err, "decoding team %v", id)
	}

	return t, nil
}

func getTeamByName(db *leveldb.DB, name string) (football.Team, error) {
	t := football.Team{}
	d, err := db.Get([]byte(fmt.Sprintf("%s%v", teamNameIndexPrefix, name)), nil)
	if err != nil {
		return t, errors.Wrapf(err, "getting team %v", name)
	}
	id, err := strconv.ParseInt(string(d), 10, 64)
	if err != nil {
		return t, errors.Wrapf(err, "decoding id for team %v", name)
	}

	d, err = db.Get([]byte(fmt.Sprintf("%s%v", teamPrefix, id)), nil)
	if err != nil {
		return t, errors.Wrapf(err, "getting team %v", name)
	}

	dec := gob.NewDecoder(bytes.NewReader(d))
	if err := dec.Decode(&t); err != nil {
		return t, errors.Wrapf(err, "decoding team %v", name)
	}

	return t, nil
}

func putTeam(db *leveldb.DB, t football.Team) error {
	var b bytes.Buffer

	enc := gob.NewEncoder(&b)
	if err := enc.Encode(t); err != nil {
		return errors.Wrapf(err, "encoding team %v", t.Id)
	}

	batch := &leveldb.Batch{}
	batch.Put([]byte(fmt.Sprintf("%s%v", teamPrefix, t.Id)), b.Bytes())
	batch.Put([]byte(fmt.Sprintf("%s%v", teamNameIndexPrefix, t.Name)), []byte(fmt.Sprintf("%d", t.Id)))

	if err := db.Write(batch, nil); err != nil {
		return errors.Wrapf(err, "writing team %v", t.Id)
	}

	return nil
}

func getPlayer(db *leveldb.DB, id football.PlayerId) (football.Player, error) {
	p := football.Player{}
	d, err := db.Get([]byte(fmt.Sprintf("%s%v", playerPrefix, id)), nil)
	if err != nil {
		return p, errors.Wrapf(err, "getting player %v", id)
	}

	dec := gob.NewDecoder(bytes.NewReader(d))
	if err := dec.Decode(&p); err != nil {
		return p, errors.Wrapf(err, "decoding player %v", id)
	}

	return p, nil
}

func putPlayer(db *leveldb.DB, p football.Player) error {
	var b bytes.Buffer

	enc := gob.NewEncoder(&b)
	if err := enc.Encode(p); err != nil {
		return errors.Wrapf(err, "encoding player %v", p.Id)
	}

	if err := db.Put([]byte(fmt.Sprintf("%s%v", playerPrefix, p.Id)), b.Bytes(), nil); err != nil {
		return errors.Wrapf(err, "writing player %v", p.Id)
	}

	return nil
}

func (o *options) apply(opts []Option) {
	for _, op := range opts {
		op.f(o)
	}
}
