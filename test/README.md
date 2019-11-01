## E2E on (Minikube)

####What is E2E

We define a E2E struct to execute the e2e flow
```golang
type E2E struct {
	Build          func() // build boot
	Check          func() // check boot \ deploy \ service is ok
	Update         func() // update the boot \ deploy \ service 
	Recheck        func() // recheck boot \ deploy \ service is ok
	BuildAndCheck  func() // execute the build and check at the same time
	UpdateAndCheck func() // execute the update and recheck at the same time
}
```

example:

```golang
// test case name
It("testing update replicas", func() {
    (&(operatorFramework.E2E{
		Build: func() {
			// create a java boot
			operatorFramework.CreateBoot(javaBoot)
		},
		Check: func() {
			// check if the book is ok
			deploy := operatorFramework.GetDeployment(bootKey)
			Expect(deploy.Spec.Replicas).Should(Equal(javaBoot.Spec.Replicas))
		},
		UpdateAndCheck: func() {
			// update  the boot and recheck it
			newReplicas := int32(3)
			boot := operatorFramework.GetBoot(bootKey)
			boot.Spec.Replicas = &newReplicas
			operatorFramework.UpdateBoot(boot)

			updateDeploy := operatorFramework.GetDeployment(bootKey)
			Expect(updateDeploy.Spec.Replicas).Should(Equal(&newReplicas))
		},
	})).Run()
})
```

####How to run
```yaml
#execute all test case
make test-e2e
```

ENV on IDE and debug

If we want to debug one case on IDE, we can set a environment,
```yaml
GINKGO_FOCUS=test case name

#more like
ginkgo -r test -focus=test case name
```

If we have a lot of test cases, we need to wait the k8s to process request.
set a environment like this, default is 1 second
```yaml
WAIT_TIME=1
```

####Run e2e on local linux/mac host
```yaml
#execute all test case
make test-e2e-local
```

Before running e2e on local linux host, we need to install following tools: 
```yaml
# install go, version should be 1.12.5 or above.

export MINIKUBE_VERSION=v1.2.0
export KUBERNETES_VERSION=v1.12.0
# install kubectl
    curl -Lo kubectl https://storage.googleapis.com/kubernetes-release/release/$KUBERNETES_VERSION/bin/linux/amd64/kubectl && \
        chmod +x kubectl &&  \
        sudo mv kubectl /usr/local/bin/
        sudo cp /usr/local/bin/kubectl $GOPATH/bin/
# install minikube
    curl -Lo minikube https://storage.googleapis.com/minikube/releases/$MINIKUBE_VERSION/minikube-linux-amd64 && \
        chmod +x minikube && \
        sudo mv minikube /usr/local/bin/
        sudo cp /usr/local/bin/minikube $GOPATH/bin/
# install oc
    curl -Lo /tmp/oc.tar.gz https://github.com/openshift/origin/releases/download/v3.11.0/openshift-origin-client-tools-v3.11.0-0cbc58b-linux-64bit.tar.gz  && \
        tar -zxvf /tmp/oc.tar.gz -C /tmp &&  \
        sudo mv /tmp/openshift-origin-client-tools-v3.11.0-0cbc58b-linux-64bit/oc /usr/local/bin/
        sudo cp /usr/local/bin/oc $GOPATH/bin/
# install operator-sdk
    wget https://github.com/operator-framework/operator-sdk/releases/download/v0.10.1/operator-sdk-v0.10.1-x86_64-linux-gnu
    mv operator-sdk-v0.10.1-x86_64-linux-gnu $GOPATH/bin/operator-sdk
    chmod +x $GOPATH/bin/operator-sdk
# install ginkgo
    go get -u github.com/onsi/ginkgo/ginkgo
```

Before running e2e on local mac host, we need to install following tools: 
```yaml
# install go, version should be 1.12.5 or above.
# install hg, virtualBox. 

export MINIKUBE_VERSION=v1.2.0
export KUBERNETES_VERSION=v1.12.0
# install kubectl
     curl -Lo kubectl https://storage.googleapis.com/kubernetes-release/release/$KUBERNETES_VERSION/bin/darwin/amd64/kubectl && \
        chmod +x kubectl &&  \
        sudo mv kubectl /usr/local/bin/
        sudo cp /usr/local/bin/kubectl $GOPATH/bin/
# install minikube
    curl -Lo minikube https://storage.googleapis.com/minikube/releases/$MINIKUBE_VERSION/minikube-darwin-amd64 && \
        chmod +x minikube && \
        sudo mv minikube /usr/local/bin/
        sudo cp /usr/local/bin/minikube $GOPATH/bin/
# install oc
    curl -Lo oc.zip https://github.com/openshift/origin/releases/download/v3.11.0/openshift-origin-client-tools-v3.11.0-0cbc58b-mac.zip  && \
        unzip oc.zip &&  \
        sudo mv oc /usr/local/bin/
        sudo cp /usr/local/bin/oc $GOPATH/bin/
# install operator-sdk
    wget https://github.com/operator-framework/operator-sdk/releases/download/v0.10.1/operator-sdk-v0.10.1-x86_64-apple-darwin
    mv operator-sdk-v0.10.1-x86_64-apple-darwin $GOPATH/bin/operator-sdk
    chmod +x $GOPATH/bin/operator-sdk
# install ginkgo
    go get -u github.com/onsi/ginkgo/ginkgo
```

If we have a lot of test cases, we need to wait the k8s to process request.
set a environment like this, default is 1 second
```yaml
export WAIT_TIME=1
```

If we have cases labeled [Slow], we need to wait longer time the k8s to process request.
set a environment like this, default is 5 second
```yaml
export SLOW_WAIT_TIME=5
```

If we need to skip all test case, set a environment like this, default is unset.
```yaml
export SKIP_TEST="yes"
```

If we want to delete minikube manually, run the command like this.
```yaml
./scripts/delete-minikube.sh
```

If virtual box on Mac ip address is not 192.168.99.0/24, set a environment like this,
```yaml
export INSECURE_REGISTRY="192.168.33.0/24"
```

if we occur error like "./minikube/client.crt not found", rerun minikube like this.
```yaml
minikube start
minikube delete
```