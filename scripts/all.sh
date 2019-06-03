#!/usr/bin/env bash
for i in {1..100}
do
    echo "Doing " $i
	oc delete java default-javaboot; oc create -f deploy/crds/test.yaml
    sleep 1s
done