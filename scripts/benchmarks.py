#!/usr/bin/env python
import os
import sys
import random
import time
import argparse

template = """\\{\\"apiVersion\\":\\"app.logancloud.com/v1\\",\\"kind\\":\\"JavaBoot\\",\\"metadata\\":\\{\\"name\\":\\"demo-boot-java%s\\"\\},\\"spec\\":\\{\\"image\\":\\"logan/logan-startkit-java\\",\\"version\\":\\"latest\\",\\"replicas\\":0\\}\\}"""

def init_boot(size, group):
    for i in range(size):
        boot = template % str(i+group)
        cmd = "echo %s | oc create -f -" % boot
        print cmd
        os.system(cmd)


def replace_boot(size, group):
    for i in range(size):
        boot = template % str(i+group)
        cmd = "echo %s | oc replace -f -" % boot
        print cmd
        os.system(cmd)


def chaos_replace_boot(size):
    for i in range(20):
        num = random.randint(0,size)
        boot = template % str(num)
        cmd = "echo %s | oc replace -f -" % boot
        print cmd
        os.system(cmd)


def delete_boot(size, group):
    for i in range(size):
        cmd = "oc delete javaboot demo-boot-java%s" % str(i+group)
        print cmd
        os.system(cmd)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description='simple boot benchmarks' )
    parser.add_argument('--group', type=int, default = 10,help='boot size')
    parser.add_argument('--size', type=int, default = 100,help='boot size')
    parser.add_argument('--op',type = str,required = True, help='op type: init replace chaos del')
    args = parser.parse_args()

    op, size, group = args.op, args.size, args.group
    time_start=time.time()

    if op == "init":
        init_boot(size, group)

    if op == "replace":
        replace_boot(size, group)

    if op == "chaos":
            chaos_replace_boot(size)

    if op == "del":
        delete_boot(size, group)

    time_end=time.time()
    print('totally cost',time_end-time_start)