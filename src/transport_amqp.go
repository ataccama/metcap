package metcap

import (
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

type AMQPTransport struct {
	InputConn       *amqp.Connection
	OutputConn      *amqp.Connection
	InputChannel    *amqp.Channel
	OutputChannel   *amqp.Channel
	Size            int
	Workers         int
	Exchange        string
	Queue           string
	ListenerEnabled bool
	WriterEnabled   bool
	Input           chan *Metric
	Output          chan *Metric
	ExitChan        chan bool
	ExitFlag        *Flag
	Wg              *sync.WaitGroup
	Logger          *Logger
	Stats           *AMQPTransportStats
}

// NewAMQPTransport
func NewAMQPTransport(c *TransportConfig, listenerEnabled bool, writerEnabled bool, exitFlag *Flag, logger *Logger) (*AMQPTransport, error) {
	// connection

	if c.AMQPTag == "" {
		c.AMQPTag = "default"
	}

	if c.BufferSize == 0 {
		c.BufferSize = 1000
	}

	var (
		inputConn     *amqp.Connection
		inputChannel  *amqp.Channel
		outputConn    *amqp.Connection
		outputChannel *amqp.Channel
		err           error
	)

	queue := "metcap:" + c.AMQPTag
	exchange := "metcap:" + c.AMQPTag
	key := "metcap:" + c.AMQPTag

	if listenerEnabled {
		inputConn, inputChannel, err = amqpInit(c)
		if err != nil {
			return nil, &TransportError{"amqp", err}
		}

		err = inputChannel.ExchangeDeclare(
			exchange, // exchange name
			"direct", // exchange type
			true,     // durable?
			false,    // auto-delete?
			false,    // internal?
			false,    // no-wait?
			nil,      // arguments
		)
		if err != nil {
			return nil, &TransportError{"amqp", err}
		}
		_, err = inputChannel.QueueDeclare(
			queue, // queue name
			true,  // durable?
			false, // auto-delete?
			false, // exclusive?
			false, // no-wait?
			nil,   // arguments
		)
		if err != nil {
			return nil, &TransportError{"amqp", err}
		}

		err = inputChannel.QueueBind(
			queue,    // queue name
			key,      // key name
			exchange, // exchange name
			false,    // no-wait?
			nil,      // arguments
		)
		if err != nil {
			return nil, &TransportError{"amqp", err}
		}
	}

	if writerEnabled {
		outputConn, outputChannel, err = amqpInit(c)
		if err != nil {
			return nil, &TransportError{"amqp", err}
		}
	}

	return &AMQPTransport{
		InputConn:       inputConn,
		OutputConn:      outputConn,
		InputChannel:    inputChannel,
		OutputChannel:   outputChannel,
		Size:            c.BufferSize,
		Workers:         c.AMQPWorkers,
		Exchange:        exchange,
		Queue:           queue,
		ListenerEnabled: listenerEnabled,
		WriterEnabled:   writerEnabled,
		Input:           make(chan *Metric, c.BufferSize),
		Output:          make(chan *Metric, c.BufferSize),
		ExitChan:        make(chan bool, 1),
		ExitFlag:        exitFlag,
		Wg:              &sync.WaitGroup{},
		Logger:          logger,
		Stats:           NewAMQPTransportStats(),
	}, nil
}

func amqpInit(c *TransportConfig) (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.DialConfig(c.AMQPURL, amqp.Config{
		Dial: func(network, addr string) (net.Conn, error) {
			return net.DialTimeout(network, addr, time.Duration(c.AMQPTimeout)*time.Second)
		},
	})
	if err != nil {
		return nil, nil, &TransportError{"amqp", err}
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, nil, &TransportError{"amqp", err}
	}

	return conn, channel, nil
}

func (t *AMQPTransport) publish(m *Metric) error {
	return t.InputChannel.Publish(
		t.Exchange, // exchange
		t.Exchange, // routing key
		false,      // mandatory?
		false,      // immediate?
		amqp.Publishing{ // message definition
			Headers:         amqp.Table{},          // AMQP message headers
			ContentType:     "application/msgpack", // content type
			ContentEncoding: "UTF-8",               // encoding
			Body:            m.Serialize(),         // serialized metric data
			DeliveryMode:    amqp.Transient,        // AMQP message delivery mode
			Priority:        0,                     // AMQP message priority
		},
	)
}

func (t *AMQPTransport) Start() {

	if t.ListenerEnabled {
		for producerCount := 1; producerCount <= t.Workers; producerCount++ {
			go func(i int) {
				t.Wg.Add(1)
				defer t.Wg.Done()
				for {
					select {
					case m := <-t.Input:
						err := t.publish(m)
						if err != nil {
							t.Logger.Errorf("[amqp] Failed to publish metric: %v", err)
						}
					case <-t.ExitChan:
						time.Sleep(1 * time.Second)
						for m := range t.Input {
							err := t.publish(m)
							if err != nil {
								t.Logger.Errorf("[amqp] Failed to publish metric: %v", err)
							}
						}
						return
					}
				}
			}(producerCount)
		}
	}

	if t.WriterEnabled {
		for consumerCount := 1; consumerCount <= t.Workers; consumerCount++ {
			go func(i int) {
				t.Wg.Add(1)
				defer t.Wg.Done()
				delivery, err := t.OutputChannel.Consume(
					t.Exchange, // queue name
					t.Exchange+":writer:"+strconv.Itoa(i), // consumer tag
					false, // autoAck? (auto acknowledge delivery)
					false, // exclusive? (there are multiple consumers)
					false, // no-local?
					true,  // no-wait?
					nil,   // arguments
				)
				if err != nil {
					t.Logger.Errorf("[amqp] Failed to setup delivery channel: %v", err)
				}
				for {
					select {
					case message := <-delivery:
						metric, err := DeserializeMetric(string(message.Body))
						if err != nil {
							message.Nack(false, false)
							t.Logger.Errorf("[amqp] Failed to deserialize metric: %v", err)
						} else {
							t.Output <- &metric
							message.Ack(false)
						}
					case <-t.ExitChan:
						for message := range delivery { // drain delivery channel
							metric, err := DeserializeMetric(string(message.Body))
							if err != nil {
								message.Nack(false, false)
								t.Logger.Errorf("[amqp] Failed to deserialize metric: %v", err)
							} else {
								t.Output <- &metric
								message.Ack(false)
							}
						}
						return
					}
				}
			}(consumerCount)
		}
	}

	go func() {
		goroutines := 0
		if t.ListenerEnabled {
			goroutines += t.Workers
		}
		if t.WriterEnabled {
			goroutines += t.Workers
		}

		for {
			switch {
			case t.ExitFlag.Get():
				t.OutputChannel.Close()
				for i := 0; i < goroutines; i++ {
					t.ExitChan <- true
				}
				t.Wg.Wait()
				return
			default:
				time.Sleep(10 * time.Millisecond)
			}
		}
	}()
}

func (t *AMQPTransport) Stop() {
	t.Wg.Wait()
	if t.ListenerEnabled {
		// close(t.Input)
		t.InputChannel.Close()
		t.InputConn.Close()
	}
	if t.WriterEnabled {
		// close(t.Output)
		t.OutputChannel.Close()
		t.OutputConn.Close()
	}
}

func (t *AMQPTransport) StopOutput() {

}

func (t *AMQPTransport) StopInput() {

}

func (t *AMQPTransport) LogReport() {

}

func (t *AMQPTransport) InputChan() chan<- *Metric {
	return t.Input
}

func (t *AMQPTransport) OutputChan() <-chan *Metric {
	return t.Output
}

func (t *AMQPTransport) InputChanLen() int {
	return len(t.Input)
}

func (t *AMQPTransport) OutputChanLen() int {
	return len(t.Output)
}

type AMQPTransportStats struct {
	MessagesInQueue     *StatsGauge
	InputChannelLength  *StatsGauge
	OutputChannelLength *StatsGauge
}

func NewAMQPTransportStats() *AMQPTransportStats {
	return &AMQPTransportStats{
		MessagesInQueue:     NewStatsGauge(),
		InputChannelLength:  NewStatsGauge(),
		OutputChannelLength: NewStatsGauge(),
	}
}

func (s *AMQPTransportStats) Reset() {}

func (s *AMQPTransportStats) Report() {}
