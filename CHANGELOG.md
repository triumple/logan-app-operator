# logan-app-operator Release Notes

## Version 0.6.0 - 10/31/2019

* Boot supports history revision
* Upgrade operator-sdk v0.10.1, use go modules
* Support restart deployment by annotation
* Optimize event output info
* Bugfix: operator’s vol config update should not restart deployment
* Bugfix: env ${PORT} replace error 

## Version 0.5.0 - 9/25/2019

* Boot support pvc 
* Boot support nodePort service 
* Raise go report grade 
* Add custom metrics for reconcile time and errors 
* Add [Serial] and [Slow] label to e2e test cases
* Bugfix: PodAntiAffinity label selector
* Bugfix: travis continue to tag version after fail

## Version 0.4.0 - 8/29/2019

* More e2e test case.
* Add prometheus annotation support for operator's service when started.

## Version 0.3.0 - 7/30/2019

* Improve travis dingding notification
* Add timezone support, and default is Asia/Shanghai. 
* Add config reloader through webhook
* Update sidecar service with runtime ports
* Change created pod labels to : bootName and bootType
* Add CRD verification for Boots
* Change PythonBoot's failureThreshold, from 10 to 15
* InitContainers support oenv config
* Concurrent improvements: set controller MaxConcurrentReconciles = cpu_num * 2
* BDD test environment improvement

## Version 0.2.0 - 6/27/2019

* Upgrade Operator SDK to 0.8.1.
* Add webhook supoort for validation and mutation.
* Add config profile support.
* Support new types: WebBoot.
* Add more fields for all Boot's: command, sessionAffinity, prometheus.
* Fix some bugs and add some docs, examples.

## Version 0.1.0 - 5/30/2019

* Initial logan-app-operator release.
