# team-search-test
Test app for listing team players and their active teams

## Installation

    go get -u github.com/urandom/team-search-test/cmd/team-players
    
## Usage
Usage printout:

    team-player -h
    
Printng out all players in a given team (or teams):

    team-players Bulgaria 'CSKA Sofia'
    
This will produce the following list:

1. Aleksandar Aleksandrov; 30; Bulgaria
2. Aleksandar Dyulgerov; 26; CSKA Sofia
3. Aleksandar Georgiev; 19; CSKA Sofia
4. Aleksandar Konov; 23; CSKA Sofia
5. Aleksandar Tonev; 26; Bulgaria, Crotone

...
