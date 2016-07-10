package graph

import (
	"github.com/jmcvetta/neoism"

	"../model"
)

type (
	Graph struct {
		database *neoism.Database
	}

	FriendshipGraph []FriendshipGraphEntry

	FriendshipGraphEntry struct {
		FromUri    string `json:"from.uri"`
		FromName   string `json:"from.name"`
		FromAvatar string `json:"from.avatar"`

		ToUri    string `json:"to.uri"`
		ToName   string `json:"to.name"`
		ToAvatar string `json:"to.avatar"`
	}
)

func Connect(connectionString string) (*Graph, error) {
	database, err := neoism.Connect(connectionString)
	return &Graph{database}, err
}

func (this *Graph) SetPlaceholderForActor(uri string) error {
	query := neoism.CypherQuery{
		Statement: `
			MERGE (a:Actor { uri: {uri} })
			ON CREATE SET a.name = {uri}, a.last_refreshed = 0
			RETURN a
		`,
		Parameters: neoism.Props{"uri": uri},
	}

	return this.database.Cypher(&query)
}

func (this *Graph) SetActor(actor *model.Actor, lastRefreshed int64) error {
	query := neoism.CypherQuery{
		Statement: `
			MERGE (a:Actor { uri: {uri} })
			ON CREATE SET a.name = {name}, a.avatar = {avatar}, a.last_refreshed = {last_refreshed}
			ON MATCH SET a.name = {name}, a.avatar = {avatar}, a.last_refreshed = {last_refreshed}
			RETURN a
		`,
		Parameters: neoism.Props{"uri": actor.Uri, "name": actor.Name, "avatar": actor.Avatar, "last_refreshed": lastRefreshed},
	}

	return this.database.Cypher(&query)
}

func (this *Graph) SetFriendship(friendship *model.Friendship) error {
	query := neoism.CypherQuery{
		Statement: `
			MATCH (f:Actor {uri: {from} }),(t:Actor { uri: {to} }) 
			MERGE (f)-[:FRIEND]->(t) 
		`,
		Parameters: neoism.Props{"from": friendship.From.Uri, "to": friendship.To.Uri},
	}

	return this.database.Cypher(&query)
}

func (this *Graph) GetFriendships(actor string) (FriendshipGraph, error) {
	result := []FriendshipGraphEntry{}

	query := neoism.CypherQuery{
		Statement: `
			MATCH (orgin:Actor)-[:FRIEND*0..2]-(from:Actor)
			MATCH (to:Actor)<-[:FRIEND]-(from)
			WHERE orgin.uri = {actor}
			RETURN from.uri, from.name, from.avatar, to.uri, to.name, to.avatar
		`,
		Parameters: neoism.Props{"actor": actor},
		Result:     &result,
	}

	err := this.database.Cypher(&query)
	return result, err
}
