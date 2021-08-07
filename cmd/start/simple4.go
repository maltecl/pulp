package main

import (
	"amigo"
	"fmt"
	"os"
	"time"
)

var _ amigo.LiveComponent = &Simple4{}

type post struct {
	title, body string
}

type Simple4 struct {
	viewed int

	posts []post
}

func (t *Simple4) Mount(socket amigo.Socket) {

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

func (t *Simple4) HandleEvent(event amigo.Event, socket amigo.Socket) {}

func (t Simple4) Render() amigo.StaticDynamic {

	arg0 := amigo.For{
		Statics:      []string{"<h3>", "</h3> <p>", "</p>"},
		ManyDynamics: make([]amigo.Dynamics, len(t.posts)),
		DiffStrategy: amigo.Append,
	}

	for i, post := range t.posts {
		arg0.ManyDynamics[i] = amigo.Dynamics{post.title, post.body}
	}

	return amigo.NewStaticDynamic(
		`ticks: <span> {} </span>
		
		posts:
			{}
		`,
		t.viewed,
		arg0,
	)
}

func (Simple4) Name() string { return "Simple4" }
