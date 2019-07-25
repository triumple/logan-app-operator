# Minikube Release Notes

## Version 0.3.0 - 7/2019

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
