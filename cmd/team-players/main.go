package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"

	"github.com/urandom/team-search-test/download"
	"github.com/urandom/team-search-test/football"
	"github.com/urandom/team-search-test/storage/goleveldb"
	"github.com/urandom/team-search-test/storage/memory"
)

var (
	defaults = []string{
		"Germany", "England", "France", "Spain", "Manchester Utd", "Arsenal", "Chelsea",
		"Barcelona", "Real Madrid", "FC Bayern Munich",
	}

	timeout     int
	workers     int
	verbose     bool
	leveldbPath string
)

func main() {
	names := defaults
	if flag.NArg() != 0 {
		names = flag.Args()
	}

	teams := download.Teams(
		download.Timeout(time.Duration(timeout)*time.Second),
		download.Workers(workers),
	)

	var repo football.TeamRepository
	if leveldbPath == "" {
		repo = memory.NewTeamRepository(teams)
	} else {
		repo = goleveldb.NewTeamRepository(teams, goleveldb.Path(leveldbPath))
	}

	var logger Logger = nopLogger{}
	if verbose {
		logger = errLogger{}
	}

	entries, err := getPlayers(repo, names, logger)
	if err != nil {
		log.Fatalf("Error getting players: %+v", err)
	}

	for _, e := range entries {
		fmt.Println(e)
	}
}

func getPlayers(repo football.TeamRepository, names []string, logger Logger) ([]string, error) {
	teamCache := map[football.TeamId]football.Team{}
	playerIds := map[football.PlayerId]struct{}{}
	players := football.Players{}

	logger.Printf("Initializing entry creating for %v\n", names)

	for _, n := range names {
		team, err := repo.GetTeamByName(n)
		if err != nil {
			return nil, err
		}

		logger.Printf("Found team %s\n", n)
		for _, pid := range team.Players {
			playerIds[pid] = struct{}{}
		}

		teamCache[team.Id] = team
	}

	for pid := range playerIds {
		player, err := repo.GetPlayer(pid)

		if err != nil {
			return nil, err
		}

		players = append(players, player)
	}

	collator := collate.New(language.English, collate.Loose)
	collator.Sort(players)

	entries := make([]string, len(players))

	for i, p := range players {
		teamNames := []string{}
		for _, tid := range p.Teams {
			if team, ok := teamCache[tid]; ok {
				teamNames = append(teamNames, team.Name)
			} else {
				team, err := repo.GetTeam(tid)
				if err != nil {
					return nil, err
				}
				teamNames = append(teamNames, team.Name)
			}
		}

		sort.Strings(teamNames)

		logger.Printf("Generating entry for player %s", p.Name)
		entries[i] = fmt.Sprintf("%d. %s; %d; %s", i+1, p.Name, p.Age, strings.Join(teamNames, ", "))
	}

	return entries, nil
}

func usage() {
	var defs bytes.Buffer

	for i, d := range defaults {
		if i%3 == 0 {
			defs.WriteString("\n\t")
		}
		defs.WriteString("'")
		defs.WriteString(d)
		defs.WriteString("' ")
	}

	fmt.Fprintf(os.Stderr, `Usage of %[1]s

	%[1]s  [team names...]

team-players extracts all players from the given teams and prints them out in
alphabetical order, including their age and affiliated teams. If no team namess
are given, the following ones will be used:
%s

`, os.Args[0], defs.String())

	flag.PrintDefaults()
}

func init() {
	flag.IntVar(&workers, "workers", 20, "number of concurrent download workers")
	flag.IntVar(&timeout, "timeout", 10, "network request timeout, in seconds")
	flag.BoolVar(&verbose, "v", false, "verbose outout")
	flag.StringVar(&leveldbPath, "leveldb-path", "", "if specified, leveldb will be used to cache the team download")
	flag.Usage = usage
	flag.Parse()
}
