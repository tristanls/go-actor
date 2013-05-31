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
//      for _, param := range msg.Params {
//        fmt.Println(param.(string))
//      }  
//      // starting go routines within an actor behavior is *NOT SAFE*
//      // create actors instead, that's what they're for :)
//    }
//
//    // create a new actor configuration
//    func main() {
//      config := actor.Configuration()
//      
//      // create a new actor
//      printer := config.Create(Print)
//     
//      // send a message to an actor
//      // both ways of creating a message are identical
//      printer <- config.CreateMessage("hello world")
//      printer <- actor.CreateMessage("hello world")
//      printer <- actor.Message{Params: []interface{}{"hello world"}}
//
//      // wait for actor configuration to finish
//      config.Wait()
//    }
package actor

import (
  "fmt"
  "sync"
)

// Become is a function that takes the behavior to handle the next message
type Become func(Behavior)

// Behavior describes how an actor will respond to a message. Actors are created 
// with behaviors. Each behavior takes a context and a message to respond to. 
// These behaviors are invoked by the library when executing an actor
// configuration.
type Behavior func(Context, Message)

// Message is just a slice of data
type Message struct {
  Params []interface{}
}

// Reference to an actor is a channel that accepts messages
type Reference chan<- Message

// Send is here for completeness, use `Reference <- Message` notation instead
func (reference Reference) Send(msg Message) {
  reference <- msg
}

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
    for message := range instrumentedReference {
      configuration.waitGroup.Add(1)
      reference <- message
    }
  }()
  return instrumentedReference
}

// CreateMessage is syntactic sugar to create a Message
func (configuration *ActorConfiguration) CreateMessage(params ...interface{}) Message {
  message := Message{Params: make([]interface{}, len(params), len(params))}
  for index, param := range params {
    message.Params[index] = param
  }
  return message
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

// CreateMessage is syntactic sugar to create a Message
func CreateMessage(params ...interface{}) Message {
  message := Message{Params: make([]interface{}, len(params), len(params))}
  for index, param := range params {
    message.Params[index] = param
  }
  return message
}

// actorBehavior implements execution of actor behaviors, become semantics,
// and making sure that the configuration doesn't exit while messages are still
// in flight.
// func actorBehavior(configuration *ActorConfiguration, behavior Behavior, reference chan Message) {
//   var become func(Behavior)
//   become = func(nextBehavior Behavior) {
//     become = func(Behavior){} // become can only be called once
//     go actorBehavior(configuration, nextBehavior, reference)
//     behavior = nil
//   }
//   for {
//     if behavior != nil {
//       messageCounter := make(chan Message)
//       self := make(chan Message)
//       go func() {
//         for msg := range messageCounter {
//           configuration.countChan <- 1
//           self <- msg
//         }
//         close(self)
//       }()
//       msg := <-reference
//       // fmt.Println("trace: <-", msg.Params)
//       context := Context{Become: become, Create: configuration.Create, Self: messageCounter}
//       behavior(context, msg)
//       close(messageCounter)
//       go func() {
//         for selfMsg := range self {
//           // fmt.Println("trace: self <-", selfMsg.Params)
//           reference <- selfMsg
//         }
//         configuration.countChan <- -1
//       }()
//     } else {
//       return
//     }
//   }
// }
func actorBehavior(configuration *ActorConfiguration, behavior Behavior, reference chan Message, instrumentedReference chan Message) {
  var become func(Behavior)
  become = func(nextBehavior Behavior) {
    become = func(Behavior){} // become can only be called once
    go actorBehavior(configuration, nextBehavior, reference, instrumentedReference)
    behavior = nil
  }
  for {
    if behavior != nil {
      msg := <-reference
      if configuration.Trace {
        fmt.Println(instrumentedReference, "<-", msg)
      }
      context := Context{Become: become, Create: configuration.Create, Self: instrumentedReference}
      behavior(context, msg)
      configuration.waitGroup.Done() // one message has been delivered
    } else {
      return // this actor behavior has been replace.. go away
    }
  }
}