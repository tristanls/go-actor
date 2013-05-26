// Go language actor framework.
//
// Create an actor behavior:
//    func Print(self actor.Reference, become actor.Become, msg actor.Message) {
//      for _, param := range msg.Params {
//        fmt.Println(param.(string))
//      }  
//    }
//
// Create a new actor configuration:
//    func main() {
//      config := actor.Configuration()
//      ...
//
// Create a new actor:
//      ...
//      printer := config.Create(Print)
//      ...
//
// Send a message to an actor:
//      ...
//      printer <- config.CreateMessage("print me")
//      printer <- actor.CreateMessage("print me")
//      ...
//
// Wait for actor configuration to finish:
//      ...
//      <-config.Done
//    }
package actor

import "fmt"

// Process the next message using the Behavior given
type Become func(Behavior)

// Actors are created with behaviors. Each behavior takes a reference to "self",
// a function to "become" another behavior, and a message to respond to.
type Behavior func(Reference, Become, Message)

// Each message is just a slice of data
type Message struct {
  Params []interface{}
}

// Actor references is how messages are sent to actors
type Reference chan<- Message

// Send is here for completeness, use `Reference <- Message` notation instead
func (reference Reference) Send(msg Message) {
  reference <- msg
}

// Block on ActorConfiguration.Done to wait for all actors to finish, for example:
//    config := actor.Configuration()
//    // send some messages to actors
//    <-config.Done
type ActorConfiguration struct {
  Done chan bool
  countChan chan<- int
  messageCount int
}

// Create a new actor as part of the actor configuration
func (configuration *ActorConfiguration) Create(behavior Behavior) Reference {
  messageCounter := make(chan Message)
  reference := make(chan Message)
  go actorBehavior(*configuration, behavior, reference)
  go func() {
    msg := <-messageCounter
    configuration.countChan <- 1
    reference <- msg
  }()
  return messageCounter
}

// CreateMessage is syntactic sugar to create a Message
func (configuration *ActorConfiguration) CreateMessage(params ...interface{}) Message {
  message := Message{Params: make([]interface{}, len(params), len(params))}
  for index, param := range params {
    message.Params[index] = param
  }
  return message
}

// Creates a new actor configuration
func Configuration() ActorConfiguration {
  countChan := make(chan int)
  configuration := ActorConfiguration{Done: make(chan bool), countChan: countChan, messageCount: 0}
  go func() {
    for {
      i := <-countChan
      if (i == 1) {
        configuration.messageCount++
        fmt.Println("trace: messages in flight", configuration.messageCount)
      } else {
        configuration.messageCount--
        fmt.Println("trace: messages in flight", configuration.messageCount)
        if (configuration.messageCount == 0) {
          configuration.Done <- true
          return
        }
      }
    }
  }()
  return configuration
}

// CreateMessage is syntactic sugar to create a Message
func CreateMessage(params ...interface{}) Message {
  message := Message{Params: make([]interface{}, len(params), len(params))}
  for index, param := range params {
    message.Params[index] = param
  }
  return message
}

func actorBehavior(configuration ActorConfiguration, behavior Behavior, reference chan Message) {
  var become func(Behavior)
  become = func(nextBehavior Behavior) {
    become = func(Behavior){} // become can only be called once
    go actorBehavior(configuration, nextBehavior, reference)
    behavior = nil
  }
  for {
    if behavior != nil {
      messageCounter := make(chan Message)
      self := make(chan Message)
      go func() {
        for msg := range messageCounter {
          configuration.countChan <- 1
          self <- msg
        }
        close(self)
      }()
      msg := <-reference
      fmt.Println("trace: <-", msg.Params)
      behavior(messageCounter, become, msg)
      close(messageCounter)
      go func() {
        for selfMsg := range self {
          fmt.Println("trace: self <-", selfMsg.Params)
          reference <- selfMsg
        }
        configuration.countChan <- -1
      }()
    } else {
      return
    }
  }
}