#!/bin/bash
#
# metcap    MetCap metrics processing engine
#
# chkconfig: 345 70 30
# description: MetCap metrics processing engine
# processname: metcap

# Source function library.
. /etc/init.d/functions

RETVAL=0
prog="metcap"
LOCKFILE=/var/lock/subsys/$prog
PIDFILE=/var/tmp/metcap.pid
METCAP_USER=metcap
METCAP_BINARY=/usr/bin/metcap

source /etc/sysconfig/metcap

start() {
  echo -n "Starting $prog: "
  [ -e $PIDFILE ] && echo "PID file exists: ${PIDFILE}" && return 1
  daemon --user $METCAP_USER $METCAP_BINARY
  RETVAL=$?
  PID=`pgrep /usr/bin/metcap`
  [[ $RETVAL -eq 0 ] && [ "${PID}" == "" ]] && ( touch $LOCKFILE; echo $PID > $PIDFILE )
  echo
  return $RETVAL
}

stop() {
  echo -n "Shutting down $prog: "
  $PID=`cat ${PIDFILE}`
  kill -TERM $PID
  for i in `seq 1 300`; do
    if [ -e /proc/$PID ]; then
      sleep 1
    else
      rm -f $PIDFILE $LOCKFILE
      echo
      return 0
    fi
  done
  echo
  return 1
}

status() {
  echo -n "Checking $prog status: "
  PID=`cat $PIDFILE`
  test -e /proc/$PID
  RETVAL=$?
  return $RETVAL
}

case "$1" in
  start)
    start
    ;;
  stop)
    stop
    ;;
  status)
    status
    ;;
  restart)
    stop
    start
    ;;
  *)
    echo "Usage: $prog {start|stop|status|restart}"
    exit 1
    ;;
esac
exit $RETVAL
