#!/bin/bash
rand(){
    min=$1
    max=$(($2-$min+1))
    num=$(date +%s%N)
    echo $(($num%$max+$min))
}
if [ $# -eq 0 ]
then
    echo "sleep interaval must be specified!"
    exit 1
fi
sleep_interval=$1 
echo "looping..."
while true
do
    rnd=$(rand 1 3)
    if [ $rnd -eq 1 ]
    then
      echo "run 'dzero' to trigger crash"
      $(./dzero)
    elif [ $rnd -eq 2 ]
    then
      echo "run 'abort' to trigger crash"
      $(./abort)
    else
      echo "do nothing this round"
    fi
    sleep $sleep_interval
done
