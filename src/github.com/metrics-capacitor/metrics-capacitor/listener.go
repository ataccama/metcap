package metcap

import (
  "fmt"
  "sync"
  "net"
  "strconv"
)

type Listener struct {
  Name    *string
  Socket  net.Listener
  Config  *ListenerConfig
  Wg      *sync.WaitGroup
  Buffer  *Buffer
}

func NewListener(name *string, c *ListenerConfig, b *Buffer, wg *sync.WaitGroup) *Listener {
  wg.Add(1)
  sock, err := net.Listen("tcp", ":" + strconv.Itoa(c.Port))
  if err != nil {
    panic(err)
  }
  return &Listener{
    Name: name,
    Socket: sock,
    Config: c,
    Wg: wg,
    Buffer: b}
}

func (l *Listener) Run() {
  for {
    conn, err := l.Socket.Accept()
    if err != nil {
      fmt.Println("Can't accept connection: " + err.Error())
    }
    go l.handle(conn)
  }
}

func (l *Listener) Stop() {
  l.Socket.Close()
  l.Wg.Done()
}

func (l *Listener) handle(conn net.Conn) {
  sockBuf := make([]byte, 0, 65535)
  _, err := conn.Read(sockBuf)
  if err != nil {
    fmt.Println("Can't copy the data from socket into the buffer: ", err.Error())
  }
  //
  // TODO
  //
  conn.Close()

}
