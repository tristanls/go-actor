package main

import "fmt"
import actor "github.com/tristanls/go-actor"

func Change(self actor.Reference, become actor.Become, msg actor.Message) {
  self <- msg
  become(Print)
}

func Print(self actor.Reference, become actor.Become, msg actor.Message) {
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