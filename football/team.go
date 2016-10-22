package football

type TeamId int
type PlayerId string

type Team struct {
	Id         TeamId
	Name       string
	IsNational bool
	Players    []PlayerId
}

type Player struct {
	Id    PlayerId
	Name  string
	Age   int
	Teams []TeamId
}

type TeamRepository interface {
	GetTeam(id TeamId) (Team, error)
	GetTeamByName(name string) (Team, error)
	GetPlayer(id PlayerId) (Player, error)
}
