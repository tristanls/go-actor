package main

import "fmt"
import "github.com/tristanls/go-actor"

func Replier(context actor.Context, msg actor.Message) {
  switch len(msg.Params) {
  case 2:
    customer, content := msg.Params[0].(actor.Reference), msg.Params[1].(string)
    if content == "ping" {
      msg := actor.CreateMessage("reply")
      customer <- msg
      fmt.Println("replier sent", msg)
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
      msg := actor.CreateMessage(context.Self, "ping")
      replier <- msg
      fmt.Println("requester sent", msg)
    case "reply":
      fmt.Println("requester got reply")
    }
  }
}

func main() {
  config := actor.Configuration()
  config.Trace = true

  requester := config.Create(Requester)
  requester <- actor.CreateMessage("start")

  config.Wait()
}