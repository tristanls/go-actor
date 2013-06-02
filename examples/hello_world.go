  package main
  import "fmt"
  import "github.com/tristanls/go-actor"
  // create an actor behavior
  func Print(context actor.Context, msg actor.Message) {
    for _, param := range msg {
      fmt.Println(param.(string))
    }
    // starting go routines within an actor behavior is *NOT SAFE*
    // create actors instead, that's what they're for :)
  }
  // create a new actor configuration
  func main() {
    config := actor.Configuration()
    config.Trace = true // trace message deliveries
    // create a new actor
    printer := config.Create(Print)
    // send a message to an actor
    printer <- actor.Message{"hello world"}
    // wait for actor configuration to finish
    config.Wait()
  }