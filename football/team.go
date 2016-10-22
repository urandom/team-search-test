package football

// TeamId is the unique identifier of a team.
type TeamId int

// PlayerId is the unique identifier of a player.
type PlayerId string

// Team contains basic information about a football team.
type Team struct {
	Id         TeamId
	Name       string
	IsNational bool
	Players    []PlayerId
}

// Player contains basic information about a football player.
type Player struct {
	Id    PlayerId
	Name  string
	Age   int
	Teams []TeamId
}

// Players is a player slice alphabetically sortable by player names.
type Players []Player

// TeamRepository allows queries for teams and players.
type TeamRepository interface {
	// GetTeams looks for a team given an id.
	GetTeam(id TeamId) (Team, error)
	// GetTeamByName looks for a team given a name.
	GetTeamByName(name string) (Team, error)
	// GetPlayer looks for a player given a player id.
	GetPlayer(id PlayerId) (Player, error)
}

func (p Players) Len() int {
	return len(p)
}

func (p Players) Less(i int, j int) bool {
	return p[i].Name < p[j].Name
}

func (p Players) Swap(i int, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p Players) Bytes(i int) []byte {
	return []byte(p[i].Name)
}
