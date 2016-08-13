package metcap

import (
  "sync"
  "net"
  "strconv"
  "bufio"
  "os"
)

type Listener struct {
  Name            string
  Socket          net.Listener
  Config          ListenerConfig
  Wg              *sync.WaitGroup
  Buffer          *Buffer
  GraphiteMutator *[]string
  Logger          *Logger
}

func NewListener(name string, c ListenerConfig, b *Buffer, wg *sync.WaitGroup, logger *Logger) Listener {
  wg.Add(1)

  logger.Infof("Starting listener '%s' [%s://0.0.0.0:%d/%s]", name, c.Protocol, c.Port, c.Codec)
  sock, err := net.Listen(c.Protocol, "0.0.0.0:" + strconv.Itoa(c.Port))
  if err != nil {
    logger.Alertf("Couldn't start listener '%s': %v", name, err)
  }

  var mut []string
  if c.Codec == "graphite" {
    logger.Debug("Detected graphite codec, loading mutator config")
    mut_file, err := os.Open(c.MutatorFile)
    if err != nil {
        logger.Alertf("Couldn't open mutator config: %v", err)
    } else {
      scn := bufio.NewScanner(mut_file)
      for scn.Scan() {
        mut = append(mut, scn.Text())
      }
      logger.Debug("Loaded mutator rules")
    }
  }

  return Listener{
    Name: name,
    Socket: sock,
    Config: c,
    Wg: wg,
    Buffer: b,
    GraphiteMutator: &mut,
    Logger: logger}
}

func (l *Listener) Run() {
  l.Logger.Infof("Starting to accept connections on '%s' listener", l.Name)
  defer l.Stop()
  for {
    c, err := l.Socket.Accept()
    if err == nil {
      l.Logger.Debugf("Accepted connection on '%s' from %s", l.Name, c.RemoteAddr().String())
      go l.handleConnection(c)
    } else {
      l.Logger.Errorf("Can't accept connection: %v", err)
    }
  }
}

func (l *Listener) Stop() {
  l.Socket.Close()
  l.Wg.Done()
}

func (l *Listener) handleConnection(conn net.Conn) {
  defer conn.Close()
  scn := bufio.NewScanner(conn)
  for scn.Scan() {
    line := scn.Text()
    metric, err := NewMetricFromLine(line, l.Config.Codec, l.GraphiteMutator)
    if err == nil {
      err = l.Buffer.Push(&metric)
      if err != nil {
        l.Logger.Errorf("Can't push metric into Redis buffer: %v", err)
      }
    } else {
      l.Logger.Errorf("%v", err)
    }
  }
}
