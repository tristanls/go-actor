  package main
  import "fmt"
  import "github.com/tristanls/go-actor"
  // create an actor behavior
  func Print(self actor.Reference, become actor.Become, msg actor.Message) {
    for _, param := range msg.Params {
      fmt.Println(param.(string))
    }
  }
  // create a new actor configuration
  func main() {
    config := actor.Configuration()
    // create a new actor
    printer := config.Create(Print)
    // send a message to an actor
    // both ways of creating a message are identical
    printer <- config.CreateMessage("hello world")
    printer <- actor.CreateMessage("hello world")
    printer <- actor.Message{Params: []interface{}{"hello world"}}
    // wait for actor configuration to finish
    <-config.Done
  }