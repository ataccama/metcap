package metcap

import (
	// "fmt"
	"fmt"
	"github.com/streadway/amqp"
	"net"
	"strconv"
	"sync"
	"time"
)

type AMQPTransport struct {
	Conn            *amqp.Connection
	Channel         *amqp.Channel
	Size            int
	Consumers       int
	Producers       int
	Exchange        string
	ListenerEnabled bool
	WriterEnabled   bool
	Listener        chan *Metric
	Writer          chan *Metric
	ExitChan        chan bool
	ExitFlag        *Flag
	Wg              *sync.WaitGroup
}

// NewAMQPTransport
func NewAMQPTransport(c *TransportConfig, listenerEnabled bool, writerEnabled bool, exitFlag *Flag) *AMQPTransport {
	// connection

	conn, err := amqp.DialConfig(c.AMQPURL, amqp.Config{
		Dial: func(network, addr string) (net.Conn, error) {
			return net.DialTimeout(network, addr, time.Duration(c.AMQPTimeout)*time.Second)
		},
	})

	// conn, err := amqp.Dial(c.AMQPURL)
	if err != nil {
		panic(err)
	}

	// channel setup
	channel, err := conn.Channel()
	if err != nil {
		panic(err)
	}

	if c.AMQPTag == "" {
		c.AMQPTag = "default"
	}

	if c.BufferSize == 0 {
		c.BufferSize = 1000
	}

	// exchange setup
	err = channel.ExchangeDeclare(
		"metcap:"+c.AMQPTag, // name
		"direct",            // type
		true,                // durable
		false,               // auto-delete
		false,               // internal
		false,               // noWait
		nil,                 // arguments
	)
	if err != nil {
		panic(err)
	}

	_, err = channel.QueueDeclare(
		"metcap:"+c.AMQPTag,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}

	err = channel.QueueBind(
		"metcap:"+c.AMQPTag,
		"metcap:"+c.AMQPTag,
		"metcap:"+c.AMQPTag,
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}

	return &AMQPTransport{
		Conn:            conn,
		Channel:         channel,
		Size:            c.BufferSize,
		Consumers:       c.AMQPConsumers,
		Producers:       c.AMQPProducers,
		Exchange:        "metcap:" + c.AMQPTag,
		ListenerEnabled: listenerEnabled,
		WriterEnabled:   writerEnabled,
		Listener:        make(chan *Metric, c.BufferSize),
		Writer:          make(chan *Metric, c.BufferSize),
		ExitChan:        make(chan bool, 1),
		ExitFlag:        exitFlag,
		Wg:              &sync.WaitGroup{},
	}
}

func (t *AMQPTransport) Start() {

	if t.ListenerEnabled {
		for producerCount := 1; producerCount <= t.Producers; producerCount++ {
			go func(i int) {
				t.Wg.Add(1)
				defer t.Wg.Done()
				for {
					select {
					case m := <-t.Listener:
						err := t.Channel.Publish(
							t.Exchange,
							"",
							false,
							false,
							amqp.Publishing{
								Headers:         amqp.Table{},
								ContentType:     "application/msgpack",
								ContentEncoding: "UTF-8",
								Body:            m.Serialize(),
								DeliveryMode:    amqp.Transient,
								Priority:        0,
							},
						)
						if err != nil {
							panic(err)
						}
					case <-t.ExitChan:
						return
					}
				}
			}(producerCount)
		}
	}

	if t.WriterEnabled {
		for consumerCount := 1; consumerCount <= t.Consumers; consumerCount++ {
			go func(i int) {
				t.Wg.Add(1)
				defer t.Wg.Done()
				delivery, err := t.Channel.Consume(
					t.Exchange,
					t.Exchange+":writer:"+strconv.Itoa(i),
					false,
					false,
					false,
					false,
					nil,
				)
				if err != nil {
					panic(err)
				}
				for {
					select {
					case m := <-delivery:
						fmt.Println(m)
						metric, err := DeserializeMetric(string(m.Body))
						if err != nil {
							m.Nack(false, false)
							fmt.Println(err)
							continue
						}
						t.Writer <- &metric
						m.Ack(false)
					case <-t.ExitChan:
						return
					}
				}
			}(consumerCount)
		}
	}

	go func() {
		goroutines := 0
		if t.ListenerEnabled {
			goroutines = goroutines + t.Producers
		}
		if t.WriterEnabled {
			goroutines = goroutines + t.Consumers
		}

		for {
			switch {
			case t.ExitFlag.Get():
				for i := 0; i < goroutines; i++ {
					t.ExitChan <- true
				}
				return
			default:
				time.Sleep(10 * time.Millisecond)
			}
		}
	}()
}

func (t *AMQPTransport) Stop() {
	t.Wg.Wait()
	t.Channel.Close()
	t.Conn.Close()
}

func (t *AMQPTransport) ListenerChan() chan<- *Metric {
	return t.Listener
}

func (t *AMQPTransport) WriterChan() <-chan *Metric {
	return t.Writer
}
