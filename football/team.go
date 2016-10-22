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

// TeamRepository allows queries for teams and players.
type TeamRepository interface {
	// GetTeams looks for a team given an id.
	GetTeam(id TeamId) (Team, error)
	// GetTeamByName looks for a team given a name.
	GetTeamByName(name string) (Team, error)
	// GetPlayer looks for a player given a player id.
	GetPlayer(id PlayerId) (Player, error)
}
