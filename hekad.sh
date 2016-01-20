#!/bin/bash
function usage {
    cat << EOF
usage:
./hekad.sh start
./hekad.sh status
./hekad.sh stop
./hekad.sh restart
EOF
    exit 0
}

#ps -ef | grep hekad
#  248  kill 22382
#  249  nohup hekad -config=/etc/hekad/hekad.toml &
STDOUT=/opt/heka/build/heka/hekad.log

function pidofproc {
    if [ $# -ne 3 ]; then
        echo "Expected three arguments, e.g. $0 -p pidfile daemon-name"
    fi

    pid=`pgrep -f $3`
    local pidfile=`cat $2`

    if [ "x$pidfile" == "x" ]; then
        return 1
    fi

    if [ "x$pid" != "x" -a "$pidfile" == "$pid" ]; then
        return 0
    fi

    return 1
}

function killproc {
    if [ $# -ne 3 ]; then
        echo "Expected three arguments, e.g. $0 -p pidfile signal"
    fi

    pid=`cat $2`

    kill -s $3 $pid
}


function log_failure_msg {
    echo "$@" "[ FAILED ]"
}

function log_success_msg {
    echo "$@" "[ OK ]"
}
name=hekad

daemon=/usr/bin/$name

pidfile=/var/run/hekad.pid

function start {

        if [ -e $pidfile ]; then
                log_failure_msg "$name process is running"

        else

                log_success_msg "Starting the process" "$name"
                nohup /opt/heka/build/heka/bin/hekad -config=/opt/heka/build/heka/bin/hekad.toml &>$STDOUT &
                echo $! > $pidfile
                log_success_msg "$name process was started"

        fi
}

function stop {

        if [ -e $pidfile ]
                then
                stoppingpid=`cat $pidfile`
                if kill $stoppingpid && /bin/rm -rf $pidfile
                then
                        log_success_msg "$name process was stopped"

                else
                    log_failure_msg "$name failed to stop service"
                fi

        else
            log_failure_msg "$name process is not running"
        fi

}

function restart {
        stop
        sleep 1
        start
}

function status {
  if [ -e $pidfile ]
        then
        runningpid=`cat $pidfile`
    echo "hekad is running as "$runningpid""
  else
    echo "hekad is not running"
  fi
}

case "$1" in
  start)
        start
        ;;
  stop)
        stop
        ;;
  restart)
        restart
        ;;
  status)
        status
        ;;
  *)
        usage
        exit 1
esac
