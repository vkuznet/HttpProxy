#!/bin/bash
nohup ./http_proxy 2>&1 1>& http_proxy.log < /dev/null &
