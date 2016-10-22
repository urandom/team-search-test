package memory

import (
	"encoding/json"
	"strconv"

	"github.com/pkg/errors"
	"github.com/urandom/team-search-test/download"
	"github.com/urandom/team-search-test/football"
)

type jsonData struct {
	Data struct {
		Team teamData `json:"team"`
	} `json:"data"`
}

type teamData struct {
	Id         football.TeamId `json:"id"`
	Name       string          `json:"name"`
	IsNational bool            `json:"IsNational"`
	Players    []playerData    `json:"players"`
}

type playerData struct {
	Id   football.PlayerId `json:"id"`
	Name string            `json:"name"`
	Age  interface{}       `json:"age"`
}

type memory struct {
	teams         map[football.TeamId]football.Team
	players       map[football.PlayerId]football.Player
	teamNameIndex map[string]football.TeamId

	init      chan struct{}
	initError error
}

// NewTeamRepository creates an in-memory team repository from the download
// data. It will start initializing the storage data from the download channel,
// blocking any queries until done. If an error occurs during initialization,
// all repository methods will return an initializer error.
//
// If a query cannot find a valid entry given the input, a not-found error will
// be returned.
func NewTeamRepository(data <-chan download.Team) football.TeamRepository {
	m := &memory{
		teams:         make(map[football.TeamId]football.Team),
		players:       make(map[football.PlayerId]football.Player),
		teamNameIndex: make(map[string]football.TeamId),
		init:          make(chan struct{}),
	}

	go m.initialize(data)

	return m
}

func (m *memory) GetTeam(id football.TeamId) (football.Team, error) {
	<-m.init

	if m.initError != nil {
		return football.Team{}, initError{errors.Wrapf(m.initError, "getting team %d", id)}
	}

	return m.getTeam(id)
}

func (m *memory) GetTeamByName(name string) (football.Team, error) {
	<-m.init

	if m.initError != nil {
		return football.Team{}, initError{errors.Wrapf(m.initError, "getting team %s", name)}
	}

	if id, ok := m.teamNameIndex[name]; ok {
		return m.getTeam(id)
	} else {
		return football.Team{}, notFoundError{errors.Errorf("no team for %s", name)}
	}
}

func (m *memory) GetPlayer(id football.PlayerId) (football.Player, error) {
	<-m.init

	if m.initError != nil {
		return football.Player{}, initError{errors.Wrapf(m.initError, "getting player %d", id)}
	}

	if p, ok := m.players[id]; ok {
		return p, nil
	} else {
		return football.Player{}, notFoundError{errors.Errorf("no player for %d", id)}
	}
}

func (m *memory) initialize(data <-chan download.Team) {
	defer close(m.init)

	for d := range data {
		var j jsonData

		err := json.Unmarshal(d.Bytes, &j)
		if err != nil {
			m.initError = errors.Wrapf(err, "parsing team data for %d", d.Id)
			return
		}

		td := j.Data.Team

		playerIds := []football.PlayerId{}

		for _, p := range td.Players {
			playerIds = append(playerIds, p.Id)

			if player, ok := m.players[p.Id]; ok {
				player.Teams = append(player.Teams, td.Id)
				m.players[p.Id] = player
			} else {
				var age int
				switch v := p.Age.(type) {
				case int:
					age = v
				case string:
					// Ignore the error, we can't do anything if the string
					// isn't numerical
					age, _ = strconv.Atoi(v)
				}
				m.players[p.Id] = football.Player{
					Id: p.Id, Name: p.Name,
					Age: age, Teams: []football.TeamId{td.Id},
				}
			}
		}

		m.teams[td.Id] = football.Team{
			Id: td.Id, Name: td.Name,
			IsNational: td.IsNational, Players: playerIds,
		}

		m.teamNameIndex[td.Name] = td.Id
	}
}

func (m *memory) getTeam(id football.TeamId) (football.Team, error) {
	if t, ok := m.teams[id]; ok {
		return t, nil
	} else {
		return football.Team{}, notFoundError{errors.Errorf("no team for %d", id)}
	}
}
