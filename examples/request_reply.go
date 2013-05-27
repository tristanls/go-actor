package main

import "fmt"
import "github.com/tristanls/go-actor"

func Replier(context actor.Context, msg actor.Message) {
  switch len(msg.Params) {
  case 2:
    customer, content := msg.Params[0].(actor.Reference), msg.Params[1].(string)
    if content == "ping" {
      customer <- actor.CreateMessage("reply")
    }
  }
}

func Requester(context actor.Context, msg actor.Message) {
  switch len(msg.Params) {
  case 1:
    command := msg.Params[0].(string)
    switch command {
    case "start":
      replier := context.Create(Replier)
      replier <- actor.CreateMessage(context.Self, "ping")
    case "reply":
      fmt.Println("got reply")
    }
  }
}

func main() {
  config := actor.Configuration()

  requester := config.Create(Requester)
  requester <- actor.CreateMessage("start")

  <-config.Done
}