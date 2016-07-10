package crawler

import (
	"log"
	"strconv"
	"time"

	"../model"

	"github.com/ChimeraCoder/anaconda"
)

const (
	MaxDepth             = 2
	MaxFriendsPerRequest = 100
)

type (
	TwitterCrawler struct {
		actor   *model.Actor
		visited map[string]struct{}
		twitter *anaconda.TwitterApi

		Actor      chan *model.Actor
		Friendship chan *model.Friendship
		Completed  chan struct{}
	}
)

func NewTwitterCrawler(actor *model.Actor, twitter *anaconda.TwitterApi) *TwitterCrawler {
	return &TwitterCrawler{
		actor:   actor,
		visited: map[string]struct{}{},
		twitter: twitter,

		Actor:      make(chan *model.Actor),
		Friendship: make(chan *model.Friendship),
		Completed:  make(chan struct{}),
	}
}

func (this *TwitterCrawler) Close() {
	close(this.Actor)
	close(this.Friendship)
	close(this.Completed)
}

func (this *TwitterCrawler) Run() {
	type QueueItem struct {
		Actor *model.Actor
		Depth int
	}
	queue := []*QueueItem{&QueueItem{this.actor, 0}}

	for ; len(queue) > 0; queue = queue[1:] {
		item := queue[0]
		if _, exist := this.visited[item.Actor.Uri]; exist {
			continue
		} else {
			this.visited[item.Actor.Uri] = struct{}{}
		}

		friends, err := this.getTwitterFriends(item.Actor.Uri)
		if err != nil {
			log.Printf("Encountered error %+v", err)
			break
		}

		for _, friend := range friends {
			this.Actor <- friend
			this.Friendship <- &model.Friendship{item.Actor, friend}

			if _, exist := this.visited[friend.Uri]; !exist && item.Depth+1 < MaxDepth {
				queue = append(queue, &QueueItem{friend, item.Depth + 1})
			}
		}
	}

	this.Completed <- struct{}{}
}

func (this *TwitterCrawler) getTwitterFriends(uri string) ([]*model.Actor, error) {
	parameters := map[string][]string{}
	parameters["user_id"] = []string{uri[15:]}
	parameters["count"] = []string{strconv.Itoa(MaxFriendsPerRequest)}

	response, err := this.twitter.GetFriendsIds(parameters)
	if err != nil {
		return nil, err
	}

	users, err := this.twitter.GetUsersLookupByIds(response.Ids, map[string][]string{})
	if err != nil {
		return nil, err
	}

	friends := []*model.Actor{}
	for _, user := range users {
		friends = append(friends, &model.Actor{
			Uri:    "twitter://user/" + user.IdStr,
			Name:   user.Name,
			Avatar: user.ProfileBackgroundImageUrlHttps,
		})
	}

	//throttle
	time.Sleep(2 * time.Second)

	return friends, nil
}
