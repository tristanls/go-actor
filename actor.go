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
//    func Print(self actor.Reference, become actor.Become, msg actor.Message) {
//      for _, param := range msg.Params {
//        fmt.Println(param.(string))
//      }  
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
//      <-config.Done
//    }
package actor

// Become is a function that takes the behavior to handle the next message
type Become func(Behavior)

// Behavior describes how an actor will respond to a message.
// Actors are created with behaviors. Each behavior takes a reference to "self",
// a function to "become" another behavior, and a message to respond to.
// These behaviors are invoked by the library when executing an actor
// configuration.
type Behavior func(Reference, Become, Message)

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

// Block on ActorConfiguration.Done to wait for all actors to finish, for example:
//    config := actor.Configuration()
//    // send some messages to actors
//    <-config.Done
type ActorConfiguration struct {
  Done chan bool
  countChan chan<- int
  messageCount int
}

// Create creats a new actor as part of the actor configuration
func (configuration *ActorConfiguration) Create(behavior Behavior) Reference {
  messageCounter := make(chan Message)
  reference := make(chan Message)
  go actorBehavior(*configuration, behavior, reference)
  go func() {
    for msg := range messageCounter {
    	configuration.countChan <- 1
    	reference <- msg
    }
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

// start starts the configuration message counter that signals when there are
// no more messages in flight and the configuration has stopped executing
func (configuration *ActorConfiguration) start(countChan chan int) {
  go func() {
    for {
      i := <-countChan
      switch i {
      case 1:
      	configuration.messageCount++
      	// fmt.Println("trace: messages in flight", configuration.messageCount)
      default:
        configuration.messageCount--
        // fmt.Println("trace: messages in flight", configuration.messageCount)
        if (configuration.messageCount == 0) {
        	configuration.Done <- true
        	return
        }
      }
    }
  }()
}

// Configuration creates a new actor configuration
func Configuration() ActorConfiguration {
  countChan := make(chan int)
  configuration := ActorConfiguration{Done: make(chan bool), countChan: countChan, messageCount: 0}
  configuration.start(countChan)
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

// actorBehavior implements execution of actor behaviors, become semantics,
// and making sure that the configuration doesn't exit while messages are still
// in flight.
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
      // fmt.Println("trace: <-", msg.Params)
      behavior(messageCounter, become, msg)
      close(messageCounter)
      go func() {
        for selfMsg := range self {
          // fmt.Println("trace: self <-", selfMsg.Params)
          reference <- selfMsg
        }
        configuration.countChan <- -1
      }()
    } else {
      return
    }
  }
}