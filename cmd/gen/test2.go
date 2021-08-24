// +build exclude

package main

import (
	"fmt"
	"os"
	"pulp"
	"time"
)

var _ pulp.LiveComponent = &Simple4{}

type post struct {
	title, body string
}

type Simple4 struct {
	viewed int

	posts []post
}

func (t *Simple4) Mount(socket pulp.Socket) {

	t.posts = []post{
		{"post 1", "body 1"},
		{"post 2", "body 2"},
		{"post 3", "body 3"},
		{"post 4", "body 4"},
	}

	go func() {

		i := 5

		twoSeconds := time.NewTicker(time.Second * 2)
		threeSeconds := time.NewTicker(time.Second * 3)

		defer func() { twoSeconds.Stop(); threeSeconds.Stop() }()

		for {
			select {
			case <-socket.Done():
				return
			// case <-twoSeconds.C:
			// 	t.viewed++
			// 	socket.Changes(t).Do()
			case <-threeSeconds.C:
				t.posts = append(t.posts, post{fmt.Sprintf("title: %d", i), "some body"})
				i++
				socket.Changes(t).Do()
			}
		}
	}()

	go func() {
		i := 0
		for {

			buf := make([]byte, 128)

			n, _ := os.Stdin.Read(buf)

			t.posts[i].body = string(buf[:n])
			socket.Changes(t).Do()
			i++
		}

	}()

	socket.Changes(t).Do()
}

func (t *Simple4) HandleEvent(event pulp.Event, socket pulp.Socket) {}

func (t Simple4) Render() pulp.HTML {
	return pulp.L(`
	ticks: <span> {{t.viewed}} </span>

	posts: 
	{{ for i, post := range t.posts }}
		<span> {{ i }} </span> <h3> {{post.title}} </h3> <p> {{post.body}} </p>
	{{end}}

`)
}

func (Simple4) Name() string { return "Simple4" }
