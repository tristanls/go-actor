package main

import "fmt"
import "github.com/tristanls/go-actor"

func Replier(context actor.Context, msg actor.Message) {
  switch len(msg) {
  case 2:
    customer, content := msg[0].(actor.Reference), msg[1].(string)
    if content == "ping" {
      msg := actor.Message{"reply"}
      customer <- msg
    }
  }
}

func Requester(context actor.Context, msg actor.Message) {
  switch len(msg) {
  case 1:
    command := msg[0].(string)
    switch command {
    case "start":
      replier := context.Create(Replier)
      msg := actor.Message{context.Self, "ping"}
      replier <- msg
    case "reply":
      fmt.Println("requester got reply")
    }
  }
}

func main() {
  config := actor.Configuration()
  config.Trace = true

  requester := config.Create(Requester)
  requester <- actor.Message{"start"}

  config.Wait()
}