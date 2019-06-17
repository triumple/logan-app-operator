## Cluster configuration(Openshift)
Openshift 3.9 and 3.11 both beed the configuration.

1. Modify cluster's master config file: /etc/origin/master/master-config.yaml

After the folloing config, add 用webhook function.
```yaml
admissionConfig：
   pluginConfig:
```
Such as：
```yaml
# mutation begin
    MutatingAdmissionWebhook:
      configuration:
        apiVersion: apiserver.config.k8s.io/v1alpha1
        kind: WebhookAdmission
        kubeConfigFile: /dev/null
    ValidatingAdmissionWebhook:
      configuration:
        apiVersion: apiserver.config.k8s.io/v1alpha1
        kind: WebhookAdmission
        kubeConfigFile: /dev/null
# mutation end
    openshift.io/ImagePolicy:
      configuration:
        apiVersion: v1
        executionRules:
```

2. Restart openshift apiserver
```yaml
# systemctl restart origin-master-api
```