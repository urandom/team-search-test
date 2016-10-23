// +build go1.7

package goleveldb_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/urandom/team-search-test/download"
	"github.com/urandom/team-search-test/football"
	"github.com/urandom/team-search-test/storage"
	"github.com/urandom/team-search-test/storage/goleveldb"
)

type team struct {
	id       football.TeamId
	name     string
	national bool
	exists   bool
}

type player struct {
	id      football.PlayerId
	name    string
	teamIds []football.TeamId
	exists  bool
}

func TestMemoryStorage(t *testing.T) {
	cases := []struct {
		data    []download.Team
		teams   []team
		players []player
	}{
		{
			data: []download.Team{
				{[]byte(team1), 1},
				{[]byte(team2), 50},
				{[]byte(team3), 100},
				{[]byte(team4), 200},
			},
			teams: []team{
				{1, "Apoel FC", false, true},
				{50, "D2", false, true},
				{100, "Czech Republic", true, true},
				{200, "Test 1", true, true},
				{2500, "sdd", false, false},
			},
			players: []player{
				{"6", "Nuno Morais", []football.TeamId{1, 200}, true},
				{"235", "Jaroslav Plasil", []football.TeamId{100, 200}, true},
				{id: "sdasd", exists: false},
			},
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			data := make(chan download.Team)

			go func() {
				for _, d := range tc.data {
					data <- d
				}
				close(data)
			}()

			defer func() {
				os.RemoveAll("/tmp/football-teams.db")
			}()

			repo := goleveldb.NewTeamRepository(data)

			for _, td := range tc.teams {
				team, err := repo.GetTeam(td.id)
				if td.exists {
					if err != nil {
						t.Fatalf("error looking for team %d: %+v", td.id, err)
					}

					if team.Id != td.id {
						t.Fatalf("expected id %d, got %d", td.id, team.Id)
					}

					if team.Name != td.name {
						t.Fatalf("expected name %s, got %s", td.name, team.Name)
					}

					if team.IsNational != td.national {
						t.Fatalf("expected isNational to be %v, got %v", td.national, team.IsNational)
					}
				} else {
					if !storage.IsNotFound(err) {
						t.Fatalf("expected not found error, got %+v", err)
					}
				}

				team, err = repo.GetTeamByName(td.name)
				if td.exists {
					if err != nil {
						t.Fatalf("error looking for team %s: %+v", td.name, err)
					}

					if team.Id != td.id {
						t.Fatalf("expected id %d, got %d", td.id, team.Id)
					}
				} else {
					if !storage.IsNotFound(err) {
						t.Fatalf("expected not found error, got %+v", err)
					}
				}
			}

			for _, p := range tc.players {
				player, err := repo.GetPlayer(p.id)
				if p.exists {
					if err != nil {
						t.Fatalf("error looking for player %s: %+v", p.id, err)
					}

					if player.Id != p.id {
						t.Fatalf("expected id %s, got %s", p.id, player.Id)
					}

					if player.Name != p.name {
						t.Fatalf("expected name %s, got %s", p.name, player.Name)
					}

					for j, tid := range p.teamIds {
						if j >= len(player.Teams) {
							t.Fatalf("expected more teams than %d", j)
						}

						if tid != player.Teams[j] {
							t.Fatalf("expected team id %d, got %d", tid, player.Teams[j])
						}
					}
				} else {
					if !storage.IsNotFound(err) {
						t.Fatalf("expected not found error, got %+v", err)
					}
				}
			}
		})
	}
}

func TestGarbageData(t *testing.T) {
	data := make(chan download.Team)

	go func() {
		data <- download.Team{[]byte(team2), 50}
		data <- download.Team{[]byte("asdasdsad"), 1}
		data <- download.Team{[]byte(team3), 100}
		close(data)
	}()

	defer func() {
		os.RemoveAll("/tmp/football-teams.db")
	}()

	repo := goleveldb.NewTeamRepository(data)
	_, err := repo.GetTeam(football.TeamId(1))

	type init interface {
		IsInitializer() bool
	}

	if !storage.IsInitializer(err) {
		t.Fatalf("expected init error, got %+v", err)
	}
}

const (
	team1 = `{"status":"ok","code":0,"data":{"team":{"id":1,"optaId":479,"name":"Apoel FC","logoUrls":[{"size":"56x56","url":"https:\/\/images.onefootball.com\/icons\/internal\/56\/1.png"},{"size":"164x164","url":"https:\/\/images.onefootball.com\/icons\/internal\/164\/1.png"}],"isNational":false,"matches":{"last":{"scoreaway":"1","scorehome":"3","status":"FullTime","id":504345,"competitionId":7,"seasonId":1709,"stadiumId":335,"matchdayId":5669746,"matchday":{"id":5669746},"kickoff":"2016-10-20T19:05:00Z","minute":94,"teamhome":{"idInternal":347,"id":1963,"name":"BSC YB","colors":{"shirtColorHome":"FF9900","shirtColorAway":"FFFFFF","crestMainColor":"","mainColor":"FF9900"},"logoUrls":[{"size":"56x56","url":"https:\/\/images.onefootball.com\/icons\/internal\/56\/347.png"},{"size":"164x164","url":"https:\/\/images.onefootball.com\/icons\/internal\/164\/347.png"}]},"teamaway":{"idInternal":1,"id":479,"name":"Apoel FC","colors":{"shirtColorHome":"0066CC","shirtColorAway":"FF9966","crestMainColor":"4F2C7D","mainColor":"0066CC"},"logoUrls":[{"size":"56x56","url":"https:\/\/images.onefootball.com\/icons\/internal\/56\/1.png"},{"size":"164x164","url":"https:\/\/images.onefootball.com\/icons\/internal\/164\/1.png"}]}},"next":{"scoreaway":"-1","scorehome":"-1","status":"PreMatch","id":504367,"competitionId":7,"seasonId":1709,"stadiumId":24,"matchdayId":5669747,"matchday":{"id":5669747},"kickoff":"2016-11-03T18:00:00Z","minute":0,"teamhome":{"idInternal":1,"id":479,"name":"Apoel FC","colors":{"shirtColorHome":"0066CC","shirtColorAway":"FF9966","crestMainColor":"4F2C7D","mainColor":"0066CC"},"logoUrls":[{"size":"56x56","url":"https:\/\/images.onefootball.com\/icons\/internal\/56\/1.png"},{"size":"164x164","url":"https:\/\/images.onefootball.com\/icons\/internal\/164\/1.png"}]},"teamaway":{"idInternal":347,"id":1963,"name":"BSC YB","colors":{"shirtColorHome":"FF9900","shirtColorAway":"FFFFFF","crestMainColor":"","mainColor":"FF9900"},"logoUrls":[{"size":"56x56","url":"https:\/\/images.onefootball.com\/icons\/internal\/56\/347.png"},{"size":"164x164","url":"https:\/\/images.onefootball.com\/icons\/internal\/164\/347.png"}]}},"following":{"scoreaway":"-1","scorehome":"-1","status":"PreMatch","id":504395,"competitionId":7,"seasonId":1709,"stadiumId":681,"matchdayId":5669748,"matchday":{"id":5669748},"kickoff":"2016-11-24T16:00:00Z","minute":0,"teamhome":{"idInternal":1874,"id":3751,"name":"FC Astana","colors":{"shirtColorHome":"","shirtColorAway":"","crestMainColor":"2B2667","mainColor":"2B2667"},"logoUrls":[{"size":"56x56","url":"https:\/\/images.onefootball.com\/icons\/internal\/56\/1874.png"},{"size":"164x164","url":"https:\/\/images.onefootball.com\/icons\/internal\/164\/1874.png"}]},"teamaway":{"idInternal":1,"id":479,"name":"Apoel FC","colors":{"shirtColorHome":"0066CC","shirtColorAway":"FF9966","crestMainColor":"4F2C7D","mainColor":"0066CC"},"logoUrls":[{"size":"56x56","url":"https:\/\/images.onefootball.com\/icons\/internal\/56\/1.png"},{"size":"164x164","url":"https:\/\/images.onefootball.com\/icons\/internal\/164\/1.png"}]}}},"competitions":[{"competitionId":140},{"competitionId":21},{"competitionId":7}],"players":[{"country":"Portugal","id":"6","firstName":"Nuno Miguel","lastName":"Morais Barbosa","name":"Nuno Morais","position":"Midfielder","number":26,"birthDate":"1984-01-29","age":"32","height":185,"weight":76,"thumbnailSrc":"https:\/\/images.onefootball.com\/default\/default_player.png"},{"country":"Cyprus","id":"19","firstName":"Nektarious","lastName":"Alexandrou","name":"Nektarious Alexandrou","position":"Midfielder","number":11,"birthDate":"1983-12-19","age":"32","height":182,"weight":76,"thumbnailSrc":"https:\/\/images.onefootball.com\/player\/98\/98bdd1b3e9ba596ffb0d8c09071a0577.jpg"},{"country":"Spain","id":"770","firstName":"Urko","lastName":"Pardo","name":"Urko Pardo","position":"Goalkeeper","number":78,"birthDate":"1983-01-28","age":"33","height":189,"weight":85,"thumbnailSrc":"https:\/\/images.onefootball.com\/player\/36\/36a9143ede9200fff4fbae81db38da60.jpg"},{"country":"Belgium","id":"915","firstName":"Igor","lastName":"de Camargo","name":"Igor de Camargo","position":"Forward","number":9,"birthDate":"1983-05-12","age":"33","height":187,"weight":83,"thumbnailSrc":"https:\/\/images.onefootball.com\/players\/915.jpg"},{"country":"Argentina","id":"2311","firstName":"Facundo","lastName":"Bertoglio","name":"Facundo Bertoglio","position":"Midfielder","number":10,"birthDate":"1990-06-30","age":"26","height":172,"weight":65,"thumbnailSrc":"https:\/\/images.onefootball.com\/default\/default_player.png"},{"country":"Brazil","id":"5075","firstName":"Carlos Roberto","lastName":"da Cruz Junior","name":"Carlao","position":"Defender","number":5,"birthDate":"1986-01-19","age":"30","height":183,"weight":76,"thumbnailSrc":"https:\/\/images.onefootball.com\/default\/default_player.png"},{"country":"Belarus","id":"6922","firstName":"Renan","lastName":"Bardini Bressan","name":"Renan Bressan","position":"Midfielder","number":88,"birthDate":"1988-11-03","age":"27","height":182,"weight":77,"thumbnailSrc":"https:\/\/images.onefootball.com\/default\/default_player.png"},{"country":"Cyprus","id":"7586","firstName":"Efstathios","lastName":"Aloneftis","name":"Efstathios Aloneftis","position":"Midfielder","number":46,"birthDate":"1983-03-29","age":"33","height":166,"weight":62,"thumbnailSrc":"https:\/\/images.onefootball.com\/default\/default_player.png"},{"country":"Cyprus","id":"7598","firstName":"Georgios","lastName":"Efrem","name":"Georgios Efrem","position":"Midfielder","number":7,"birthDate":"1989-07-05","age":"27","height":174,"weight":70,"thumbnailSrc":"https:\/\/images.onefootball.com\/default\/default_player.png"},{"country":"Cyprus","id":"8029","firstName":"Giorgos","lastName":"Merkis","name":"Giorgos Merkis","position":"Defender","number":30,"birthDate":"1984-07-30","age":"32","height":183,"weight":78,"thumbnailSrc":"https:\/\/images.onefootball.com\/default\/default_player.png"},{"country":"Spain","id":"12108","firstName":"Andrea","lastName":"Orlandi","name":"Andrea Orlandi","position":"Midfielder","number":8,"birthDate":"1984-08-03","age":"32","height":180,"weight":78,"thumbnailSrc":"https:\/\/images.onefootball.com\/default\/default_player.png"},{"country":"Spain","id":"12204","firstName":"Roberto","lastName":"Lago","name":"Roberto Lago","position":"Defender","number":3,"birthDate":"1985-08-30","age":"31","height":178,"weight":70,"thumbnailSrc":"https:\/\/images.onefootball.com\/players\/12204.jpg"},{"country":"Bulgaria","id":"14775","firstName":"Zhivko","lastName":"Milanov","name":"Zhivko Milanov","position":"Defender","number":21,"birthDate":"1984-07-15","age":"32","height":177,"weight":71,"thumbnailSrc":"https:\/\/images.onefootball.com\/default\/default_player.png"},{"country":"Brazil","id":"18651","firstName":"Vinicius","lastName":"Oliveira Franco","name":"Vinicius","position":"Midfielder","number":16,"birthDate":"1986-05-16","age":"30","height":186,"weight":74,"thumbnailSrc":"https:\/\/images.onefootball.com\/default\/default_player.png"},{"country":"Netherlands","id":"20459","firstName":"Boy","lastName":"Waterman","name":"Boy Waterman","position":"Goalkeeper","number":99,"birthDate":"1984-01-24","age":"32","height":188,"weight":91,"thumbnailSrc":"https:\/\/images.onefootball.com\/player\/b4\/b4e7fe7ff16121d2ece4f7ad7cc7391a.jpg"},{"country":"Spain","id":"23382","firstName":"Inaki","lastName":"Astiz","name":"Inaki Astiz","position":"Defender","number":23,"birthDate":"1983-11-05","age":"32","height":185,"weight":73,"thumbnailSrc":"https:\/\/images.onefootball.com\/players\/23382.jpg"},{"country":"Portugal","id":"27915","firstName":"Mario","lastName":"Sergio","name":"Mario Sergio","position":"Defender","number":28,"birthDate":"1981-07-28","age":"35","height":174,"weight":70,"thumbnailSrc":"https:\/\/images.onefootball.com\/default\/default_player.png"},{"country":"Greece","id":"33568","firstName":"Giannis","lastName":"Gianniotas","name":"Giannis Gianniotas","position":"Midfielder","number":70,"birthDate":"1993-04-29","age":"23","height":174,"weight":71,"thumbnailSrc":"https:\/\/images.onefootball.com\/player\/49\/49b89c316379e14fdb785c602fdc1039.jpg"},{"country":"Cyprus","id":"36113","firstName":"Kostakis","lastName":"Artymatas","name":"Kostakis Artymatas","position":"Midfielder","number":4,"birthDate":"1993-04-15","age":"23","height":184,"weight":77,"thumbnailSrc":"https:\/\/images.onefootball.com\/default\/default_player.png"},{"country":"Cyprus","id":"36114","firstName":"Pieros","lastName":"Soteriou","name":"Pieros Soteriou","position":"Forward","number":20,"birthDate":"1993-01-13","age":"23","height":186,"weight":81,"thumbnailSrc":"https:\/\/images.onefootball.com\/default\/default_player.png"},{"country":"Brazil","id":"50382","firstName":"Vander","lastName":"Vieira","name":"Vander Vieira","position":"Midfielder","number":77,"birthDate":"1988-10-03","age":"28","height":172,"weight":76,"thumbnailSrc":"https:\/\/images.onefootball.com\/default\/default_player.png"},{"country":"Cyprus","id":"62036","firstName":"Vasilios","lastName":"Papafotis","name":"Vasilios Papafotis","position":"Midfielder","number":31,"birthDate":"1995-08-10","age":"21","height":178,"weight":66,"thumbnailSrc":"https:\/\/images.onefootball.com\/default\/default_player.png"},{"country":"Cyprus","id":"68641","firstName":"Nicholas","lastName":"Ioannou","name":"Nicholas Ioannou","position":"Defender","number":44,"birthDate":"1995-11-10","age":"20","height":183,"weight":77,"thumbnailSrc":"https:\/\/images.onefootball.com\/default\/default_player.png"},{"country":"Albania","id":"111745","firstName":"Qazim","lastName":"Laci","name":"Qazim Laci","position":"Midfielder","number":14,"birthDate":"1996-01-19","age":"20","height":176,"weight":80,"thumbnailSrc":"https:\/\/images.onefootball.com\/default\/default_player.png"},{"country":"Cyprus","id":"179472","firstName":"Kypros","lastName":"Christoforou","name":"Kypros Christoforou","position":"Defender","number":0,"birthDate":"1993-04-23","age":"23","height":0,"weight":0,"thumbnailSrc":"https:\/\/images.onefootball.com\/default\/default_player.png"},{"country":"Cyprus","id":"185880","firstName":"Andreas","lastName":"Paraskevas","name":"Andreas Paraskevas","position":"Goalkeeper","number":98,"birthDate":"1998-09-15","age":"18","height":187,"weight":79,"thumbnailSrc":"https:\/\/images.onefootball.com\/default\/default_player.png"},{"country":"Cyprus","id":"185884","firstName":"Michalis","lastName":"Charalampous","name":"Michalis Charalampous","position":"Forward","number":19,"birthDate":"1999-01-29","age":"17","height":0,"weight":0,"thumbnailSrc":"https:\/\/images.onefootball.com\/default\/default_player.png"}],"officials":[{"countryName":"Spain","id":"49381","firstName":"Thomas","lastName":"Christiansen","country":"ES","position":"Coach"}],"colors":{"shirtColorHome":"0066CC","shirtColorAway":"FF9966","crestMainColor":"4F2C7D","mainColor":"0066CC"}}},"message":"Team feed successfully generated. Api Version: 1"}`
	team2 = `{"status":"ok","code":0,"data":{"team":{"id":50,"optaId":5382,"name":"D2","logoUrls":[{"size":"56x56","url":"https:\/\/images.onefootball.com\/icons\/internal\/56\/50.png"},{"size":"164x164","url":"https:\/\/images.onefootball.com\/icons\/internal\/164\/50.png"}],"isNational":false,"matches":{},"competitions":[],"players":[],"officials":[],"colors":{"shirtColorHome":"","shirtColorAway":"","crestMainColor":"","mainColor":""}}},"message":"Team feed successfully generated. Api Version: 1"}`
	team3 = `{"status":"ok","code":0,"data":{"team":{"id":100,"optaId":367,"name":"Czech Republic","logoUrls":[{"size":"56x56","url":"https:\/\/images.onefootball.com\/icons\/internal\/56\/100.png"},{"size":"164x164","url":"https:\/\/images.onefootball.com\/icons\/internal\/164\/100.png"}],"isNational":true,"matches":{"last":{"scoreaway":"0","scorehome":"0","status":"FullTime","id":450538,"competitionId":69,"seasonId":1319,"stadiumId":6512,"matchdayId":5663870,"matchday":{"id":5663870},"kickoff":"2016-10-11T18:45:00Z","minute":96,"teamhome":{"idInternal":100,"id":367,"name":"Czech Republic","colors":{"shirtColorHome":"CC0000","shirtColorAway":"FFFFFF","crestMainColor":"D5131A","mainColor":"CC0000"},"logoUrls":[{"size":"56x56","url":"https:\/\/images.onefootball.com\/icons\/internal\/56\/100.png"},{"size":"164x164","url":"https:\/\/images.onefootball.com\/icons\/internal\/164\/100.png"}]},"teamaway":{"idInternal":309,"id":505,"name":"Azerbaijan","colors":{"shirtColorHome":"FF0000","shirtColorAway":"3366FF","crestMainColor":"0E8EBA","mainColor":"FF0000"},"logoUrls":[{"size":"56x56","url":"https:\/\/images.onefootball.com\/icons\/internal\/56\/309.png"},{"size":"164x164","url":"https:\/\/images.onefootball.com\/icons\/internal\/164\/309.png"}]}},"next":{"scoreaway":"-1","scorehome":"-1","status":"PreMatch","id":450548,"competitionId":69,"seasonId":1319,"stadiumId":478,"matchdayId":5663871,"matchday":{"id":5663871},"kickoff":"2016-11-11T19:45:00Z","minute":0,"teamhome":{"idInternal":100,"id":367,"name":"Czech Republic","colors":{"shirtColorHome":"CC0000","shirtColorAway":"FFFFFF","crestMainColor":"D5131A","mainColor":"CC0000"},"logoUrls":[{"size":"56x56","url":"https:\/\/images.onefootball.com\/icons\/internal\/56\/100.png"},{"size":"164x164","url":"https:\/\/images.onefootball.com\/icons\/internal\/164\/100.png"}]},"teamaway":{"idInternal":115,"id":363,"name":"Norway","colors":{"shirtColorHome":"CC0000","shirtColorAway":"FFFFFF","crestMainColor":"EE2B2C","mainColor":"CC0000"},"logoUrls":[{"size":"56x56","url":"https:\/\/images.onefootball.com\/icons\/internal\/56\/115.png"},{"size":"164x164","url":"https:\/\/images.onefootball.com\/icons\/internal\/164\/115.png"}]}},"following":{"scoreaway":"-1","scorehome":"-1","status":"PreMatch","id":450610,"competitionId":69,"seasonId":1319,"stadiumId":266,"matchdayId":5663872,"matchday":{"id":5663872},"kickoff":"2017-03-26T16:00:00Z","minute":0,"teamhome":{"idInternal":300,"id":495,"name":"San Marino","colors":{"shirtColorHome":"0000FF","shirtColorAway":"FFFFFF","crestMainColor":"5EB6E3","mainColor":"0000FF"},"logoUrls":[{"size":"56x56","url":"https:\/\/images.onefootball.com\/icons\/internal\/56\/300.png"},{"size":"164x164","url":"https:\/\/images.onefootball.com\/icons\/internal\/164\/300.png"}]},"teamaway":{"idInternal":100,"id":367,"name":"Czech Republic","colors":{"shirtColorHome":"CC0000","shirtColorAway":"FFFFFF","crestMainColor":"D5131A","mainColor":"CC0000"},"logoUrls":[{"size":"56x56","url":"https:\/\/images.onefootball.com\/icons\/internal\/56\/100.png"},{"size":"164x164","url":"https:\/\/images.onefootball.com\/icons\/internal\/164\/100.png"}]}}},"competitions":[{"competitionId":69},{"competitionId":20},{"competitionId":24}],"players":[{"country":"Czech Republic","id":"46","firstName":"Tomas","lastName":"Rosicky","name":"Tomas Rosicky","position":"Midfielder","number":0,"birthDate":"1980-10-04","age":"36","height":178,"weight":65,"thumbnailSrc":"https:\/\/images.onefootball.com\/players\/46.jpg"},{"country":"Czech Republic","id":"194","firstName":"Tomas","lastName":"Sivok","name":"Tomas Sivok","position":"Defender","number":0,"birthDate":"1983-09-15","age":"33","height":184,"weight":77,"thumbnailSrc":"https:\/\/images.onefootball.com\/players\/194.jpg"},{"country":"Czech Republic","id":"235","firstName":"Jaroslav","lastName":"Plasil","name":"Jaroslav Plasil","position":"Midfielder","number":0,"birthDate":"1982-01-05","age":"34","height":182,"weight":72,"thumbnailSrc":"https:\/\/images.onefootball.com\/players\/235.jpg"},{"country":"Czech Republic","id":"300","firstName":"Tomas","lastName":"Necid","name":"Tomas Necid","position":"Forward","number":0,"birthDate":"1989-08-13","age":"27","height":190,"weight":89,"thumbnailSrc":"https:\/\/images.onefootball.com\/players\/300.jpg"},{"country":"Czech Republic","id":"1846","firstName":"Daniel","lastName":"Pudil","name":"Daniel Pudil","position":"Defender","number":0,"birthDate":"1985-09-27","age":"31","height":185,"weight":81,"thumbnailSrc":"https:\/\/images.onefootball.com\/players\/1846.jpg"},{"country":"Czech Republic","id":"1853","firstName":"Michal","lastName":"Kadlec","name":"Michal Kadlec","position":"Defender","number":0,"birthDate":"1984-12-13","age":"31","height":185,"weight":76,"thumbnailSrc":"https:\/\/images.onefootball.com\/players\/1853.jpg"},{"country":"Czech Republic","id":"1857","firstName":"Jaroslav","lastName":"Drobny","name":"Jaroslav Drobny","position":"Goalkeeper","number":0,"birthDate":"1979-10-18","age":"37","height":192,"weight":90,"thumbnailSrc":"https:\/\/images.onefootball.com\/players\/1857.jpg"},{"country":"Czech Republic","id":"1860","firstName":"David","lastName":"Lafata","name":"David Lafata","position":"Forward","number":0,"birthDate":"1981-09-18","age":"35","height":180,"weight":69,"thumbnailSrc":"https:\/\/images.onefootball.com\/players\/1860.jpg"},{"country":"Czech Republic","id":"1861","firstName":"Marek","lastName":"Suchy","name":"Marek Suchy","position":"Defender","number":0,"birthDate":"1988-03-29","age":"28","height":183,"weight":80,"thumbnailSrc":"https:\/\/images.onefootball.com\/players\/1861.jpg"},{"country":"Czech Republic","id":"1863","firstName":"Roman","lastName":"Hubnik","name":"Roman Hubnik","position":"Defender","number":0,"birthDate":"1984-06-06","age":"32","height":192,"weight":83,"thumbnailSrc":"https:\/\/images.onefootball.com\/players\/1863.jpg"},{"country":"Czech Republic","id":"2831","firstName":"Matej","lastName":"Vydra","name":"Matej Vydra","position":"Forward","number":0,"birthDate":"1992-05-01","age":"24","height":180,"weight":71,"thumbnailSrc":"https:\/\/images.onefootball.com\/default\/default_player.png"},{"country":"Czech Republic","id":"5556","firstName":"Borek","lastName":"Dockal","name":"Borek Dockal","position":"Midfielder","number":0,"birthDate":"1988-09-30","age":"28","height":182,"weight":71,"thumbnailSrc":"https:\/\/images.onefootball.com\/players\/5556.jpg"},{"country":"Czech Republic","id":"7805","firstName":"Vaclav","lastName":"Kadlec","name":"Vaclav Kadlec","position":"Forward","number":0,"birthDate":"1992-05-20","age":"24","height":181,"weight":77,"thumbnailSrc":"https:\/\/images.onefootball.com\/players\/7805.jpg"},{"country":"Czech Republic","id":"7815","firstName":"Ladislav","lastName":"Krejci","name":"Ladislav Krejci","position":"Midfielder","number":0,"birthDate":"1992-07-05","age":"24","height":180,"weight":68,"thumbnailSrc":"https:\/\/images.onefootball.com\/players\/7815.jpg"},{"country":"Czech Republic","id":"11122","firstName":"David","lastName":"Limbersky","name":"David Limbersky","position":"Defender","number":0,"birthDate":"1983-10-06","age":"33","height":178,"weight":73,"thumbnailSrc":"https:\/\/images.onefootball.com\/players\/11122.jpg"},{"country":"Czech Republic","id":"11126","firstName":"Daniel","lastName":"Kol\u00e1r","name":"Daniel Kol\u00e1r","position":"Midfielder","number":0,"birthDate":"1985-10-27","age":"30","height":179,"weight":76,"thumbnailSrc":"https:\/\/images.onefootball.com\/players\/11126.jpg"},{"country":"Czech Republic","id":"17050","firstName":"Pavel","lastName":"Kader\u00e1bek","name":"Pavel Kader\u00e1bek","position":"Defender","number":0,"birthDate":"1992-04-25","age":"24","height":182,"weight":81,"thumbnailSrc":"https:\/\/images.onefootball.com\/players\/17050.jpg"},{"country":"Czech Republic","id":"17051","firstName":"Jiri","lastName":"Skalak","name":"Jiri Skalak","position":"Midfielder","number":0,"birthDate":"1992-03-12","age":"24","height":177,"weight":76,"thumbnailSrc":"https:\/\/images.onefootball.com\/players\/17051.jpg"},{"country":"Czech Republic","id":"19492","firstName":"Tom\u00e1s","lastName":"Vaclik","name":"Tom\u00e1s Vaclik","position":"Goalkeeper","number":0,"birthDate":"1989-03-29","age":"27","height":188,"weight":84,"thumbnailSrc":"https:\/\/images.onefootball.com\/players\/19492.jpg"},{"country":"Czech Republic","id":"20152","firstName":"Theodor","lastName":"Gebre Selassie","name":"Theodor Gebre Selassie","position":"Defender","number":0,"birthDate":"1986-12-24","age":"29","height":181,"weight":71,"thumbnailSrc":"https:\/\/images.onefootball.com\/players\/20152.jpg"},{"country":"Czech Republic","id":"22864","firstName":"Vladimir","lastName":"Darida","name":"Vladimir Darida","position":"Midfielder","number":0,"birthDate":"1990-08-08","age":"26","height":171,"weight":64,"thumbnailSrc":"https:\/\/images.onefootball.com\/players\/22864.jpg"},{"country":"Czech Republic","id":"23155","firstName":"Filip","lastName":"Novak","name":"Filip Novak","position":"Defender","number":0,"birthDate":"1990-06-26","age":"26","height":0,"weight":0,"thumbnailSrc":"https:\/\/images.onefootball.com\/default\/default_player.png"},{"country":"Czech Republic","id":"23163","firstName":"Jan","lastName":"Kopic","name":"Jan Kopic","position":"Midfielder","number":0,"birthDate":"1990-06-04","age":"26","height":177,"weight":71,"thumbnailSrc":"https:\/\/images.onefootball.com\/default\/default_player.png"},{"country":"Czech Republic","id":"23444","firstName":"Ondrej","lastName":"Zahustel","name":"Ondrej Zahustel","position":"Midfielder","number":25,"birthDate":"1991-06-18","age":"25","height":183,"weight":70,"thumbnailSrc":"https:\/\/images.onefootball.com\/default\/default_player.png"},{"country":"Czech Republic","id":"23596","firstName":"Jakub","lastName":"Brabec","name":"Jakub Brabec","position":"Defender","number":5,"birthDate":"1992-08-06","age":"24","height":186,"weight":78,"thumbnailSrc":"https:\/\/images.onefootball.com\/default\/default_player.png"},{"country":"Czech Republic","id":"23598","firstName":"David","lastName":"Pavelka","name":"David Pavelka","position":"Midfielder","number":0,"birthDate":"1991-05-18","age":"25","height":184,"weight":74,"thumbnailSrc":"https:\/\/images.onefootball.com\/players\/23598.jpg"},{"country":"Czech Republic","id":"26476","firstName":"Josef","lastName":"Sural","name":"Josef Sural","position":"Midfielder","number":0,"birthDate":"1990-05-30","age":"26","height":184,"weight":81,"thumbnailSrc":"https:\/\/images.onefootball.com\/players\/26476.jpg"},{"country":"Czech Republic","id":"26659","firstName":"Martin","lastName":"Frydek","name":"Martin Frydek","position":"Midfielder","number":0,"birthDate":"1992-03-24","age":"24","height":180,"weight":76,"thumbnailSrc":"https:\/\/images.onefootball.com\/default\/default_player.png"},{"country":"Czech Republic","id":"27321","firstName":"Tomas","lastName":"Kalas","name":"Tomas Kalas","position":"Defender","number":0,"birthDate":"1993-05-15","age":"23","height":182,"weight":73,"thumbnailSrc":"https:\/\/images.onefootball.com\/players\/27321.jpg"},{"country":"Czech Republic","id":"38968","firstName":"Martin","lastName":"Pospisil","name":"Martin Pospisil","position":"Midfielder","number":0,"birthDate":"1991-06-26","age":"25","height":178,"weight":73,"thumbnailSrc":"https:\/\/images.onefootball.com\/default\/default_player.png"},{"country":"Czech Republic","id":"38978","firstName":"Tomas","lastName":"Horava","name":"Tomas Horava","position":"Midfielder","number":0,"birthDate":"1988-05-29","age":"28","height":180,"weight":68,"thumbnailSrc":"https:\/\/images.onefootball.com\/default\/default_player.png"},{"country":"Czech Republic","id":"94163","firstName":"Jiri","lastName":"Pavlenka","name":"Jiri Pavlenka","position":"Goalkeeper","number":23,"birthDate":"1992-04-14","age":"24","height":0,"weight":0,"thumbnailSrc":"https:\/\/images.onefootball.com\/default\/default_player.png"},{"country":"Czech Republic","id":"96191","firstName":"Milan","lastName":"Skoda","name":"Milan Skoda","position":"Forward","number":0,"birthDate":"1986-01-16","age":"30","height":190,"weight":80,"thumbnailSrc":"https:\/\/images.onefootball.com\/players\/96191.jpg"},{"country":"Czech Republic","id":"103943","firstName":"Tomas","lastName":"Koubek","name":"Tomas Koubek","position":"Goalkeeper","number":0,"birthDate":"1992-08-26","age":"24","height":197,"weight":99,"thumbnailSrc":"https:\/\/images.onefootball.com\/players\/103943.jpg"}],"officials":[{"countryName":"Czech Republic","id":"33779","firstName":"Karel","lastName":"Jarol\u00edm","country":"CZ","position":"Coach"}],"colors":{"shirtColorHome":"CC0000","shirtColorAway":"FFFFFF","crestMainColor":"D5131A","mainColor":"CC0000"}}},"message":"Team feed successfully generated. Api Version: 1"}`
	team4 = `
{
  "status": "ok",
  "code": 0,
  "data": {
    "team": {
      "id": 200,
      "name": "Test 1",
	  "IsNational": true,
	  "players": [
		  {
			  "id": "235",
			  "name": "Jaroslav Plasil",
			  "age": 34
		  },
		  {
			  "id": "6",
			  "name": "Nuno Morais",
			  "age": "32"
		  }
	  ]
	}
  }
}
`
)
