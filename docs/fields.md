# How to add field
- Such as: add the ``Command`` for all Boot types.

## 
1. pkg/apis/app/v1/boot_types.go
- ```BootSpec```:

2. Run
```shell
operator-sdk generate k8s
```

3. Add command's definition to CRD's definition
- deploy/crds/app_v1_javaboot_crd.yaml
- deploy/crds/app_v1_phpboot_crd.yaml
- deploy/crds/app_v1_pythonboot_crd.yaml
- deploy/crds/app_v1_nodejsboot_crd.yaml
- deploy/crds/app_v1_webboot_crd.yaml

4. Add Business logic
- logan/operator/boot_handler.go
