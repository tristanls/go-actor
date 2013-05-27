package main

import "fmt"
import "github.com/tristanls/go-actor"

func Change(context actor.Context, msg actor.Message) {
  context.Self <- msg
  context.Become(Print)
}

func Print(context actor.Context, msg actor.Message) {
  for _, param := range msg.Params {
    fmt.Println(param.(string))
  }
}

func main() {
  config := actor.Configuration()
  change := config.Create(Change)
  change <- config.CreateMessage("foo")
  <-config.Done
}