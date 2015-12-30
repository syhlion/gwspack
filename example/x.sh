#!/bin/bash
examplePID=`ps -ef | grep example | grep -v grep | awk '{print $2}'`

if [ -n "$examplePID" ] ; then
 echo try stop examle pid $examplePID
 kill $examplePID
fi
rm all.log
./example  >> all.log 2>&1 &

echo example run pid ${!}

