package validation

import (
	"context"
	"fmt"
	appv1 "github.com/logancloud/logan-app-operator/pkg/apis/app/v1"
	v1 "github.com/logancloud/logan-app-operator/pkg/apis/app/v1"
	"github.com/logancloud/logan-app-operator/pkg/logan/config"
	"github.com/logancloud/logan-app-operator/pkg/logan/operator"
	"github.com/logancloud/logan-app-operator/pkg/logan/util"
	"github.com/logancloud/logan-app-operator/pkg/logan/webhook"
	admssionv1beta1 "k8s.io/api/admission/v1beta1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission/types"
)

// BootValidator is a Handler that implements interfaces: admission.Handler, inject.Client and inject.Decoder
type BootValidator struct {
	client  client.Client
	decoder types.Decoder
}

var _ admission.Handler = &BootValidator{}

// Handle is the actual logic that will be called by every webhook request
func (vHandler *BootValidator) Handle(ctx context.Context, req types.Request) types.Response {
	if operator.Ignore(req.AdmissionRequest.Namespace) {
		return admission.ValidationResponse(true, "")
	}

	msg, valid, err := vHandler.Validate(req)
	if err != nil {
		logger.Error(err, msg)
		return admission.ErrorResponse(http.StatusBadRequest, err)
	}

	if !valid {
		return webhook.ValidationResponse(false, http.StatusBadRequest, msg)
	}

	return admission.ValidationResponse(true, "")
}

var _ inject.Client = &BootValidator{}

// InjectClient will inject client into BootValidator
func (vHandler *BootValidator) InjectClient(c client.Client) error {
	vHandler.client = c
	return nil
}

var _ inject.Decoder = &BootValidator{}

// InjectDecoder will inject decoder into BootValidator
func (vHandler *BootValidator) InjectDecoder(d types.Decoder) error {
	vHandler.decoder = d
	return nil
}

// Validate will do the validating for request boot.
// Returns
//   msg: Error message
//   valid: true if valid, otherwise false
//   error: decoding error, otherwise nil
func (vHandler *BootValidator) Validate(req types.Request) (string, bool, error) {
	operation := req.AdmissionRequest.Operation

	boot, err := webhook.DecodeBoot(req, vHandler.decoder)
	if err != nil {
		return "Decoding request error", false, err
	}

	if boot == nil {
		logger.Info("Can not recognize the bootType", "bootType", req.AdmissionRequest.Kind.Kind)
		return "Can not decoding boot", false, nil
	}

	// Only Check Boot's names when creating.
	if operation == admssionv1beta1.Create {
		msg, valid := vHandler.BootNameExist(boot)

		if !valid {
			logger.Info(msg)
			return msg, false, nil
		}
	}

	// Check Boot's envs when creating or updating.
	if operation == admssionv1beta1.Create || operation == admssionv1beta1.Update {
		msg, valid := vHandler.CheckEnvKeys(boot, operation)

		if !valid {
			logger.Info(msg)
			return msg, false, nil
		}
	}

	logger.Info("Validation Boot valid: ",
		"name", boot.Name, "namespace", boot.Namespace, "operation", operation)

	return "", true, nil
}

// BootNameExist check if name is exist.
// Returns
//    msg: error message
//    valid: If exist false, otherwise false
func (vHandler *BootValidator) BootNameExist(boot *v1.Boot) (string, bool) {
	c := vHandler.client

	namespaceName := k8stypes.NamespacedName{
		Namespace: boot.Namespace,
		Name:      boot.Name,
	}

	err := c.Get(context.TODO(), namespaceName, &appv1.JavaBoot{})
	if err == nil {
		return fmt.Sprintf("Boot's name %s exists in type %s", namespaceName, webhook.ApiTypeJava), false
	}

	err = c.Get(context.TODO(), namespaceName, &appv1.PhpBoot{})
	if err == nil {
		return fmt.Sprintf("Boot's name %s exists in type %s", namespaceName, webhook.ApiTypePhp), false
	}

	err = c.Get(context.TODO(), namespaceName, &appv1.PythonBoot{})
	if err == nil {
		return fmt.Sprintf("Boot's name %s exists in type %s", namespaceName, webhook.ApiTypePython), false
	}

	err = c.Get(context.TODO(), namespaceName, &appv1.NodeJSBoot{})
	if err == nil {
		return fmt.Sprintf("Boot's name %s exists in type %s", namespaceName, webhook.ApiTypeNodeJS), false
	}

	err = c.Get(context.TODO(), namespaceName, &appv1.WebBoot{})
	if err == nil {
		return fmt.Sprintf("Boot's name %s exists in type %s", namespaceName, webhook.ApiTypeWeb), false
	}

	return "", true
}

// CheckEnvKeys check the boot's env keys.
// Returns
//    msg: error message
//    valid: If valid false, otherwise false
func (vHandler *BootValidator) CheckEnvKeys(boot *v1.Boot, operation admssionv1beta1.Operation) (string, bool) {
	configSpec := operator.GetConfigSpec(boot)
	if configSpec == nil {
		logger.Info("AppSpec is nil, valid is true.")
		return "", true
	}

	//Creating: should not contains the key in global settings.
	if operation == admssionv1beta1.Create {
		for _, cfgEnv := range configSpec.Env {
			cfgEnvName := cfgEnv.Name
			// Decode the ${APP}, ${ENV} context
			cfgEnvValue, _ := operator.Decode(boot, cfgEnv.Value)

			for _, env := range boot.Spec.Env {
				if env.Name == cfgEnvName {
					if env.Value != cfgEnvValue {
						return fmt.Sprintf("Boot's added Env [%s=%s] not allowed with settings [%s=%s]",
							env.Name, env.Value, cfgEnvName, cfgEnvValue), false
					}
				}
			}
		}

		return "", true
	}

	if operation == admssionv1beta1.Update {
		return checkEnvUpdate(configSpec, boot)
	}

	return "", true
}

func checkEnvUpdate(configSpec *config.AppSpec, boot *v1.Boot) (string, bool) {
	bootMetaEnvs, err := operator.DecodeAnnotationEnvs(boot)
	if err != nil {
		logger.Error(err, "Boot's annotation env decode error")
		return "", true
	}

	if bootMetaEnvs == nil {
		// First update(By Controller), or manually deleting the annotation's env.
		logger.Info("Boot's annotation env decode empty",
			"namespace", boot.Namespace, "name", boot.Name)
		return "", true
	}

	deleted, added, modified := util.Difference2(bootMetaEnvs, boot.Spec.Env)

	logger.V(2).Info("Validating Boot", "deleted", deleted,
		"added", added, "modified", modified)

	for _, cfgEnv := range configSpec.Env {
		cfgEnvName := cfgEnv.Name
		// Decode the ${APP}, ${ENV} context
		cfgEnvValue, _ := operator.Decode(boot, cfgEnv.Value)

		// 1. Manual Delete key of Env: If key exists in global settings, valid is false.
		for _, env := range deleted {
			if env.Name == cfgEnvName {
				return fmt.Sprintf("Boot's deleted Env [%s=%s] not allowed with settings [%s=%s]",
					env.Name, env.Value, cfgEnvName, cfgEnvValue), false
			}
		}

		// 2. Manual Add key of Env: If key exists in global settings, and value not equal, valid is false.
		for _, env := range added {
			if env.Name == cfgEnvName {
				if env.Value != cfgEnvValue {
					return fmt.Sprintf("Boot's added Env [%s=%s] not allowed with settings [%s=%s]",
						env.Name, env.Value, cfgEnvName, cfgEnvValue), false
				}
			}
		}

		// 3. Manual Modify value of Env: If key exists in global settings, and value not equal, valid is false.
		for _, env := range modified {
			if env.Name == cfgEnvName {
				if env.Value != cfgEnvValue {
					return fmt.Sprintf("Boot's edit Env [%s=%s] not allowed with settings [%s=%s]",
						env.Name, env.Value, cfgEnvName, cfgEnvValue), false
				}
			}
		}
	}

	return "", true
}
