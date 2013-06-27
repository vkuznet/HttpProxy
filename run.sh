#!/bin/bash
me=HttpProxy
cmd=$PWD/$me
port=":9998"
verbose=0
log=http_proxy.log
version=1.0.0

# Start the service.
start()
{
  echo "starting $me"
  nohup $cmd -port=$port -verbose=$verbose 2>&1 1>& $log < /dev/null &
}

# Stop the service.
stop()
{
  pid=`ps | grep $me | grep -v grep | awk '{print $1}'`
  if  [ -n "$pid" ]; then
    echo "stopping $me pid=$pid"
    kill -9 $pid
  fi
}

# Check if the server is running.
status()
{
  pid=`ps | grep $me | grep -v grep | awk '{print $1}'`
  if  [ -n "$pid" ]; then
    echo "$me is running, pid=$pid"
  else
    echo "$me is not running"
  fi
}
help()
{
  $cmd -help
}
showLog()
{
  tail -f $log
}

# Main routine, perform action requested on command line.
case ${1:-status} in
  start | restart )
    stop
    start
    status
    ;;

  status )
    status
    ;;

  stop )
    stop
    status
    ;;

  help )
    help
    ;;

  log )
    showLog
    ;;

  version )
    echo "$me version $version"
    ;;

  * )
    echo "$0: unknown action '$1', please try '$0 help' or documentation." 1>&2
    exit 1
    ;;
esac
