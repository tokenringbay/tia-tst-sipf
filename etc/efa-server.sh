#! /bin/sh
# /etc/init.d/efa-server
#
# EFA server init script.
#

### BEGIN INIT INFO
# Provides:          efa-server
# Required-Start:    $remote_fs $syslog
# Required-Stop:     $remote_fs $syslog
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# Short-Description: Start EFA server at boot time.
# Description:       Enable EFA server.
### END INIT INFO

PID=/var/run/efa-server.pid
USER=nobody
BIN=/opt/efa/efa-server

# Carry out specific functions when asked to by the system
case "$1" in
  start)
    echo "Starting EFA server."
    start-stop-daemon --start --pidfile $PID --make-pidfile --user $USER --background --exec $BIN
    ;;
  stop)
    if [ -f $PID ]; then
      echo "Stopping EFA server.";
      start-stop-daemon --stop --pidfile $PID
    else
      echo "EFA server is not running.";
    fi
    ;;
  restart)
    echo "Restarting EFA server."
    start-stop-daemon --stop --pidfile $PID
    start-stop-daemon --start --pidfile $PID --make-pidfile --user $USER --background --exec $BIN
    ;;
  status)
    if [ -f $PID ]; then
      echo "EFA server is running.";
    else
      echo "EFA server is not running.";
      exit 3
    fi
    ;;
  *)
    echo "Usage: /etc/init.d/efa-server {start|stop|status|restart}"
    exit 1
    ;;
esac

exit 0