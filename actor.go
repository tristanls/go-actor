package actor

import "fmt"

type Become func(Behavior)

type Behavior func(Reference, Become, Message)

type Message struct {
  Params []interface{}
}

type Reference chan<- Message
func (reference Reference) Send(msg Message) {
  reference <- msg
}

type actorConfiguration struct {
  Done chan bool
  countChan chan<- int
  messageCount int
}
func (configuration *actorConfiguration) Create(behavior Behavior) Reference {
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

func (configuration *actorConfiguration) CreateMessage(params ...interface{}) Message {
  message := Message{Params: make([]interface{}, len(params), len(params))}
  for index, param := range params {
    message.Params[index] = param
  }
  return message
}

func Configuration() actorConfiguration {
  countChan := make(chan int)
  configuration := actorConfiguration{Done: make(chan bool), countChan: countChan, messageCount: 0}
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

func CreateMessage(params ...interface{}) Message {
  message := Message{Params: make([]interface{}, len(params), len(params))}
  for index, param := range params {
    message.Params[index] = param
  }
  return message
}

func actorBehavior(configuration actorConfiguration, behavior Behavior, reference chan Message) {
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