#!/bin/bash
source "./kube.sh"
source "./util.sh"

IMAGE="wlhtorresowen\/fujitsu-coredump-integration-testing:v0.1"
DEPLOYMENT_TEMPLATE="deployment_template.yaml"
COREDUMP_TEMPLATE="coredumpendpoint_template.yaml"
LABEL="status=coredump-test"

test_authority_all(){
    test_authority create deployment
    test_authority create namespace
    test_authority create coredumpendpoints
    test_authority get coredumpendpoints
    test_authority get coredumpendpoints dump
    test_authority get pods log
}

# 'create_deployment default 10'
#  means creating a deployment with replicaset 10 in namespace default
create_deployment(){
    namespace=$1
    pods_num=$2

    deployment_name=$(generate_random_str_with_random_length 5 10)
    yaml_file=$(echo "/tmp/dep_"$deployment_name".yaml")
    cp $DEPLOYMENT_TEMPLATE $yaml_file

    # replace namespace
    sed -i "s/__NAMESPACE__/${namespace}/g" $yaml_file

    # replace deployment name
    sed -i "s/__DEPLOYMENT_NAME__/${deployment_name}/g" $yaml_file

    # replace image name
    sed -i "s/__CRASH_TRIGGER_CONTAINER__/${IMAGE}/g" $yaml_file

    # replace container name 
    container_name_1=$(generate_random_str_with_random_length 5 10)
    container_name_2=$(generate_random_str_with_random_length 5 10)
    sed -i "s/__CONTAINER_NAME_1__/${container_name_1}/g" $yaml_file
    sed -i "s/__CONTAINER_NAME_2__/${container_name_2}/g" $yaml_file

    # replace job period duration for container 
    period_duration_1=$(generate_random_num 1 20)
    period_duration_2=$(generate_random_num 1 20)
    sed -i "s/__PERIOD_1__/${period_duration_1}/g" $yaml_file
    sed -i "s/__PERIOD_2__/${period_duration_2}/g" $yaml_file

    # replace pods num
    pods_num=$(generate_random_num 1 10)
    sed -i "s/__PODS_NUM__/${pods_num}/g" $yaml_file

    create_via_file $yaml_file
}

# build_coredump $pod_max
build_coredump(){
    current_pod_sum=0
    pod_max=$1
    while [ $current_pod_sum -lt $pod_max ]
    do
        namespace=$(generate_random_str_with_random_length 5 10)
        kubectl create ns $namespace
        # failure reason varied, we just ingore it and move to next round
        if [ $? -ne 0 ];then
            echo_red "Warning: faile to create namespace $namespace"
            continue
        fi

        # pre condition: the namespace must be exist
        kube::test::wait_object_assert "namespaces/${namespace}" '{{.metadata.name}}' "${namespace}"
        kubectl label ns $namespace $LABEL
        if [ $? -ne 0 ];then
            echo_red "Warning: faile to label namespace $namespace, delete it"
            kubectl delete ns $namespace
            continue
        fi

        for i in $(seq 1 3)  
        do
            pods_num=$(generate_random_num 1 4)
            create_deployment $namespace $pods_num 
            if [ $? -eq 0 ];then
                let current_pod_sum+=$pods_num
            fi
        done
    done
}

test_create_coredumpendpoint(){
    namespace=$1
    pod=$2

    yaml_file=$(echo "/tmp/cde_"$pod".yaml")
    cp $COREDUMP_TEMPLATE $yaml_file
    sed -i "s/__NAMESPACE__/${namespace}/g" $yaml_file
    sed -i "s/__NAME__/${pod}/g" $yaml_file
    create_via_file $yaml_file
}

# test_download_core_files namespace pod container
test_download_core_files(){
    namespace=$1
    pod=$2
    container=$3

    # test if we have core files generated in that container
    output_message=$(kubectl logs $pod -n $namespace $container)
    kube::test::if_has_string "${output_message}" "core dumped"  > /dev/null 2>&1
    if [ $? -ne 0 ];then
        return 0
    fi

    
    raw_link=$(echo "/apis/coredump.fujitsu.com/v1alpha1/namespaces/${namespace}/coredumpendpoints/${pod}/dump?${container}")
    #kubectl get --raw=/apis/coredump.fujitsu.com/v1alpha1/namespaces/default/coredumpendpoints/etcd-0/dump
    kubectl get --raw=$raw_link > /dev/null 2>&1
}

test_basic(){
    for namespace in $(kubectl get ns -o custom-columns=:.metadata.name --no-headers -l $LABEL);do
        echo "handle $namespace"
        for pod in $(kubectl get pod -n $namespace -o custom-columns=:.metadata.name --no-headers);do
            status=$(kubectl get pod $pod -n $namespace -o jsonpath="{.status.phase}")
            if [ $status != "Running" ];then
                continue
            fi
            test_create_coredumpendpoint $namespace $pod
            if [ $? -ne 0 ];then
                echo_red "Warning:create coredumpendpoint failed for $namespace/$pod"
                continue
            fi
            for container in $(kubectl get pod $pod -n $namespace -o jsonpath={.spec.containers[*].name});do
                test_download_core_files $namespace $pod $container
                if [ $? -ne 0 ];then
                    echo_red "Error: unable to download core files for container $container in $namespace/$pod"
                else
                    echo_green "Download core files for container $container in $namespace/$pod successed"
                fi
            done
        done
    done
}

clean_namespace(){
    kubectl delete ns -l $LABEL
}

if [ $# -gt 0 ];then
echo "Use $1 as kube config file to access k8s cluster"
export KUBECONFIG=$1
else
echo "Use default settings as kube config file to access k8s cluster"
fi
test_connect
test_authority_all

if [ $# -gt 1 ];then
echo "Set test pod sum to $2"
pod_max=$2
else
pod_max=20
echo "Set test pod sum to default value: $pod_max"
fi

build_coredump $pod_max
echo "*****************NOW THE TEST DEPLOYMENTS HAS BEEN LAUNCHED************"
echo "Please enter 'y' when most pods are ready, enter other to exit the test and do clean job"
read answer
if [ $answer == "y" ];then
    echo "*****************BASIC TEST BEGIN************"
    test_basic
else
    echo "*****************ww************"
    clean_namespace
fi
