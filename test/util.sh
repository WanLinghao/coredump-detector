#!/bin/bash

# 'generate_random_str 10' will generate a random str with fix length 10
# all be lower case
generate_random_str(){
   echo $(head /dev/urandom | tr -dc a-z0-9 | head -c $1 ; echo '')
}

# 'generate_random_num 5 10' will generate a random number between 5-10
generate_random_num(){
    min=$1
    max=$(($2-$min+1))
    num=$(date +%s%N)
    echo $(($num%$max+$min))
}

# 'generate_random_str_with_random_length 5 10' will generate a random str with random length between 5-10
# all be lower case
generate_random_str_with_random_length(){
    length=$(generate_random_num $1 $2)
    echo $(generate_random_str $length)
}

test_connect(){
    output_message=$(kubectl version)
    kube::test::if_has_not_string "${output_message}" "Unable to connect to the server: Forbidden" > /dev/null 2>&1
    if [ $? -ne 0 ];then
        echo_red "Error: can't connect to api-server, please check your KUBECONFIG setting" 1>&2
        exit 1
    fi
}

test_authority(){
    if [ $# -eq 2 ];then
        output_message=$(kubectl auth can-i $1 $2 2>&1)
        kube::test::if_has_string "${output_message}" "yes"
        if [ $? -ne 0 ];then
            echo_red "Error: the test needs access to $1 $2" 1>&2
            exit 1
        fi
    elif [ $# -eq 3 ];then
        output_message=$(kubectl auth can-i $1 $2 --subresource=$3 2>&1)
        kube::test::if_has_string "${output_message}" "yes"
        if [ $? -ne 0 ];then
            echo_red "Error: the test needs access to $1 $2/$3" 1>&2
            exit 1
        fi
    else
        echo_red "Error: unexpeceted paramaters for test_authority: $@" 1>&2
        exit 1
    fi
}

create_via_file(){
    yaml_file=$1
    kubectl create -f $yaml_file
    ret=$?
#    rm $yaml_file
    if [ $ret -eq 0 ];then
        echo_green "Create resource by $yaml_file succeed"
        return 0
    else
        echo_red "Error: Create resource by $yaml_file failed"
        return 1
    fi
}

# red, green, reset defined in ./lib/test.sh
echo_green(){
    echo -n "${green}"
    echo $1
    echo -n "${reset}"
}

echo_red(){
    echo -n "${red}"
    echo $1
    echo -n "${reset}"
}
