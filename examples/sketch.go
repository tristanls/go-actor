package main

import "fmt"
import "github.com/tristanls/go-actor"

func Change(context actor.Context, msg actor.Message) {
  context.Self <- msg
  context.Self <- msg
  context.Self <- msg
  context.Self <- msg
  context.Become(Print)
}

func Print(context actor.Context, msg actor.Message) {
  for _, param := range msg {
    fmt.Println(param.(string))
  }
}

func main() {
  config := actor.Configuration()
  config.Trace = true
  change := config.Create(Change)
  change <- actor.Message{"foo"}
  config.Wait()
}