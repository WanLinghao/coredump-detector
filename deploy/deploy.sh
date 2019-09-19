#!/bin/bash
# This script will generate a yaml file that contains all the components of k8s core dump feature.
# 1. It will use local directory as backend storage, '/tmp/coredump-backend-storage' for storing core files
#    '/tmp/etcd-backend' for storing etcd data.
# 2. You could replace those backend storage with your own one, but be careful that it supports only filesystem storage now.
# 3. It would also generate ca bundles automatically, you could also use your own certificates.
# 4. Run the script like './deploy.sh [namespace] [name] [dir path]
#    [namespace] is the namespace that all the components would run in
#    [name] is the service name for aggregation api layer.
#    [dir path] is the output path, the script will create one if the dir is not exist
# 5. For more detail, please see https://kubernetes.io/docs/tasks/access-kubernetes-api/setup-extension-api-server/

extract_pem(){
     cat $1|base64|awk 'NF {sub(/\r/, ""); printf "%s",$0;}'
}

replace_str(){
    ori_str=$1
    new_str=$2
    file=$3

    echo $ori_str
    echo $new_str
    echo $file
    sed -i "s#$ori_str#$new_str#g" $file
}
if [ $# -ne 3 ];then
    echo "bad paramaters, you should offer namespace, name, filepath"
    exit 1
fi
namespace=$1
name=$2
dir=$3

if [ ! -d "$dir" ]; then
    mkdir $dir
    if [ $? -ne 0 ];then
        echo "create $dir failed"
        exit 1
    fi
fi

openssl req -x509 -newkey rsa:2048 -keyout $dir/apiserver_ca.key  -out $dir/apiserver_ca.crt -days 365 -nodes -subj /C=un/ST=st/L=l/O=o/OU=ou/CN=$name-certificate-authority
openssl req -out $dir/apiserver.csr -new -newkey rsa:2048 -nodes -keyout $dir/apiserver.key -subj /C=un/ST=st/L=l/O=o/OU=ou/CN=$name.$namespace.svc
openssl x509 -req -days 365 -in $dir/apiserver.csr -CA $dir/apiserver_ca.crt -CAkey $dir/apiserver_ca.key -CAcreateserial -out $dir/apiserver.crt

cp ./coredump_template.yaml $dir/coredump.yaml

api_svc_pem=$(extract_pem $dir/apiserver_ca.crt)
echo $api_svc_pem
replace_str __APISERVICE_CA_BUNDLE__ $api_svc_pem $dir/coredump.yaml

svc_crt=$(extract_pem $dir/apiserver.crt)
replace_str __SERVICE_TLS_CRT__ $svc_crt $dir/coredump.yaml

svc_key=$(extract_pem $dir/apiserver.key)
replace_str __SERVICE_TLS_KEY__ $svc_key $dir/coredump.yaml

replace_str __NAMESPACE__ $namespace $dir/coredump.yaml
replace_str __SERVICE_NAME__ $name $dir/coredump.yaml
