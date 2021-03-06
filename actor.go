// Go language actor framework.
//
// Stability: 1 - Experimental
// (see: https://github.com/tristanls/stability-index#stability-1---experimental)
//
//    package main
//
//    import "fmt"
//    import "github.com/tristanls/go-actor"
//
//    // create an actor behavior
//    func Print(context actor.Context, msg actor.Message) {
//      for _, param := range msg {
//        fmt.Println(param.(string))
//      }  
//      // starting go routines within an actor behavior is *NOT SAFE*
//      // create actors instead, that's what they're for :)
//    }
//
//    // create a new actor configuration
//    func main() {
//      config := actor.Configuration()
//      config.Trace = true // trace message deliveries
//      
//      // create a new actor
//      printer := config.Create(Print)
//     
//      // send a message to an actor
//      printer <- actor.Message{"hello world"}
//
//      // wait for actor configuration to finish
//      config.Wait()
//    }
package actor

import (
  "fmt"
  "sync"
  "time"
)

// Become is a function that takes the behavior to handle the next message
type Become func(Behavior)

// Behavior describes how an actor will respond to a message. Actors are created 
// with behaviors. Each behavior takes a context and a message to respond to. 
// These behaviors are invoked by the library when executing an actor
// configuration.
type Behavior func(Context, Message)

// Message is just a slice of data
type Message []interface{}

// Reference to an actor is a channel that accepts messages
type Reference chan<- Message

// ActorConfiguration holds the state of an actor configuration
// To wait for all actors to finish call ActorConfiguration.Wait(), for example:
//    config := actor.Configuration()
//    // send some messages to actors
//    config.Wait() // wait for all actors to finish
// "finish" in this case means that all actors finished execution and there are
// no more messages in flight
type ActorConfiguration struct {
  Trace bool
  waitGroup sync.WaitGroup
}

// Create creates a new actor as part of the actor configuration
func (configuration *ActorConfiguration) Create(behavior Behavior) Reference {
  instrumentedReference := make(chan Message)
  reference := make(chan Message)
  go actorBehavior(configuration, behavior, reference, instrumentedReference)
  go func() {
    buffer := []Message{}
    receiveLoop:    
    for {
      if len(buffer) == 0 {
        message, ok := <-instrumentedReference
        if !ok { break }
        configuration.waitGroup.Add(1)
        if configuration.Trace {
          fmt.Println(instrumentedReference, "queued", message)
        }
        buffer = append(buffer, message)
      }

      select {
      case message, ok := <-instrumentedReference:
        if !ok { break receiveLoop }
        configuration.waitGroup.Add(1)
        if configuration.Trace {
          fmt.Println(instrumentedReference, "queued", message)
        }
        buffer = append(buffer, message)
      case reference <- buffer[0]:
        buffer = buffer[1:]
      }
    }
    for _, message := range buffer {
      reference <- message
    }
  }()
  if configuration.Trace {
    fmt.Println(instrumentedReference, "created")
  }
  return instrumentedReference
}

// Wait blocks until the actor configuration finishes executing
func (configuration *ActorConfiguration) Wait() {
  configuration.waitGroup.Wait()
}

// Context is passed to an actor behavior when it is invoked upon a message receive
type Context struct {
	Become Become
	Create func(Behavior) Reference
	Self Reference
}

// Configuration creates a new actor configuration
func Configuration() ActorConfiguration {
  return ActorConfiguration{}
}

// actorBehavior implements execution of actor behaviors, become semantics,
// and making sure that the configuration doesn't exit while messages are still
// in flight.
func actorBehavior(configuration *ActorConfiguration, behavior Behavior, reference chan Message, instrumentedReference chan Message) {
  var become func(Behavior)
  become = func(nextBehavior Behavior) {
    become = func(Behavior){} // become can only be called once
    go actorBehavior(configuration, nextBehavior, reference, instrumentedReference)
    behavior = nil
  }
  for {
    if behavior != nil {
      message := <-reference
      if configuration.Trace {
        fmt.Println(instrumentedReference, "<-", message)
      }
      context := Context{Become: become, Create: configuration.Create, Self: instrumentedReference}
      behavior(context, message)
      // there is a race condition between last message delivery in behavior
      // and Done() being called below, where the message is pulled off of a
      // channel, Done() executes, config finishes, and then only Add(1) 
      // is called :/ ... by then, main() exits
      // the current fix is to wait a set amount of time before
      // reporting finishing the behavior (behavior already finished, only
      // the report is delayed to allow for other messages to be delivered)
      // perhaps this parameter should be configurable?
      // it also need only be a tiny amount of time, just enough to allow
      // the channels sent to in the behavior above to increment the 
      // wait group counter
      go func() {
        time.Sleep(100 * time.Millisecond)
        configuration.waitGroup.Done() // one message has been processed
      }()
      if configuration.Trace {
        fmt.Println(instrumentedReference, "completed", message)
      }
    } else {
      return // this actor behavior has been replaced.. go away
    }
  }
}