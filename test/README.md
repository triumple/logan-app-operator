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

