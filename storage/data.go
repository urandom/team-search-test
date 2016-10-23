package storage

import "github.com/urandom/team-search-test/football"

type JsonData struct {
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
