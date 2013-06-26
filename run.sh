#!/bin/bash
pid=`ps | grep HttpProxy | grep -v grep | awk '{print $1}'`
if  [ -n "$pid" ]; then
    echo "Kill previous HttpProxy, pid=$pid"
    kill -9 $pid
fi
nohup ./HttpProxy 2>&1 1>& http_proxy.log < /dev/null &
