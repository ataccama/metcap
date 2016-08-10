package metcap

import (
  "fmt"
  "sync"
  "net"
  "strconv"
  "bufio"
)

type Listener struct {
  Name    *string
  Socket  net.Listener
  Config  *ListenerConfig
  Wg      *sync.WaitGroup
  Buffer  *Buffer
}

func NewListener(name string, c ListenerConfig, b *Buffer, wg *sync.WaitGroup) Listener {
  wg.Add(1)
  sock, err := net.Listen("tcp", ":" + strconv.Itoa(c.Port))
  if err != nil {
    panic(err)
  }
  return Listener{
    Name: &name,
    Socket: sock,
    Config: &c,
    Wg: wg,
    Buffer: b}
}

func (l Listener) Run() {
  defer l.Stop()
  for {
    connection, err := l.Socket.Accept()
    if err == nil {
      go l.handleConnection(connection)
    } else {
      fmt.Println("ERROR: Can't accept connection:" + err.Error())
    }
  }
}

func (l Listener) Stop() {
  l.Socket.Close()
  l.Wg.Done()
}

func (l Listener) handleConnection(conn net.Conn) {
  defer conn.Close()
  scn := bufio.NewScanner(conn)
  for scn.Scan() {
    line := scn.Text()
    metric, err := NewMetricFromLine(line, l.Config.Codec, &[]string{})
    if err != nil {
      fmt.Println("ERROR: Can't parse metric data:" + err.Error())
    }
    err = l.Buffer.Push(&metric)
    if err != nil {
      fmt.Println("ERROR: Can't push metric into Redis buffer:" + err.Error())
    }
  }
}
