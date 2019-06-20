package mutation

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/appscode/jsonpatch"
	v1 "github.com/logancloud/logan-app-operator/pkg/apis/app/v1"
	"github.com/logancloud/logan-app-operator/pkg/controller/javaboot"
	"github.com/logancloud/logan-app-operator/pkg/controller/nodejsboot"
	"github.com/logancloud/logan-app-operator/pkg/controller/phpboot"
	"github.com/logancloud/logan-app-operator/pkg/controller/pythonboot"
	"github.com/logancloud/logan-app-operator/pkg/controller/webboot"
	"github.com/logancloud/logan-app-operator/pkg/logan/operator"
	"github.com/logancloud/logan-app-operator/pkg/logan/webhook"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"net/http"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission/types"
)

// Now BootMutator only add an annotation to the Boot.

// BootMutator is a Handler that implements interfaces: admission.Handler, inject.Client and inject.Decoder
type BootMutator struct {
	client   client.Client
	decoder  types.Decoder
	Schema   *runtime.Scheme
	Recorder record.EventRecorder
}

var logger = logf.Log.WithName("logan_webhook_mutation")

var _ admission.Handler = &BootMutator{}

// Handle is the actual logic that will be called by every webhook request
func (mHandler *BootMutator) Handle(ctx context.Context, req types.Request) types.Response {
	if operator.Ignore(req.AdmissionRequest.Namespace) {
		return admission.PatchResponse(&v1.Boot{}, &v1.Boot{})
	}

	patchResponse, err := mHandler.mutateBoot(ctx, req)
	if err != nil {
		logger.Error(err, "mutate error")
		return admission.ErrorResponse(http.StatusInternalServerError, err)
	}

	logger.V(2).Info("patch", "result", patchResponse)

	return patchResponse
}

// mutateBoot mutate the Boot
func (mHandler *BootMutator) mutateBoot(ctx context.Context, req types.Request) (types.Response, error) {
	c := mHandler.client
	scheme := mHandler.Schema
	recorder := mHandler.Recorder

	bootType := req.AdmissionRequest.Kind.Kind

	if bootType == webhook.ApiTypeJava {
		javaBoot, err := webhook.DecodeJavaBoot(req, mHandler.decoder)
		if err != nil {
			logger.Error(err, "Decoding boot error.")
		}
		bootCopy := javaBoot.DeepCopy()

		handler := javaboot.InitHandler(bootCopy, scheme, c, logger, recorder)

		mutationDefault(handler, req, bootCopy.Name)
		mutationBoot(&bootCopy.ObjectMeta, req)

		marshaledBoot, err := json.Marshal(bootCopy)
		if err != nil {
			return admission.ErrorResponse(http.StatusInternalServerError, err), err
		}
		return PatchResponseFromRaw(req.AdmissionRequest.Object.Raw, marshaledBoot), nil
	} else if bootType == webhook.ApiTypePhp {
		phpBoot, err := webhook.DecodePhpBoot(req, mHandler.decoder)
		if err != nil {
			logger.Error(err, "Decoding boot error.")
		}
		bootCopy := phpBoot.DeepCopy()

		handler := phpboot.InitHandler(bootCopy, scheme, c, logger, recorder)

		mutationDefault(handler, req, bootCopy.Name)
		mutationBoot(&bootCopy.ObjectMeta, req)

		marshaledBoot, err := json.Marshal(bootCopy)
		if err != nil {
			return admission.ErrorResponse(http.StatusInternalServerError, err), err
		}
		return PatchResponseFromRaw(req.AdmissionRequest.Object.Raw, marshaledBoot), nil
	} else if bootType == webhook.ApiTypePython {
		pythonBoot, err := webhook.DecodePythonBoot(req, mHandler.decoder)
		if err != nil {
			logger.Error(err, "Decoding boot error.")
		}
		bootCopy := pythonBoot.DeepCopy()

		handler := pythonboot.InitHandler(bootCopy, scheme, c, logger, recorder)

		mutationDefault(handler, req, bootCopy.Name)
		mutationBoot(&bootCopy.ObjectMeta, req)

		marshaledBoot, err := json.Marshal(bootCopy)
		if err != nil {
			return admission.ErrorResponse(http.StatusInternalServerError, err), err
		}
		return PatchResponseFromRaw(req.AdmissionRequest.Object.Raw, marshaledBoot), nil
	} else if bootType == webhook.ApiTypeNodeJS {
		nodejsBoot, err := webhook.DecodeNodeJSBoot(req, mHandler.decoder)
		if err != nil {
			logger.Error(err, "Decoding boot error.")
		}
		bootCopy := nodejsBoot.DeepCopy()

		handler := nodejsboot.InitHandler(bootCopy, scheme, c, logger, recorder)

		mutationDefault(handler, req, bootCopy.Name)
		mutationBoot(&bootCopy.ObjectMeta, req)

		marshaledBoot, err := json.Marshal(bootCopy)
		if err != nil {
			return admission.ErrorResponse(http.StatusInternalServerError, err), err
		}
		return PatchResponseFromRaw(req.AdmissionRequest.Object.Raw, marshaledBoot), nil
	} else if bootType == webhook.ApiTypeWeb {
		webBoot, err := webhook.DecodeWebBoot(req, mHandler.decoder)
		if err != nil {
			logger.Error(err, "Decoding boot error.")
		}
		bootCopy := webBoot.DeepCopy()

		handler := webboot.InitHandler(bootCopy, scheme, c, logger, recorder)

		mutationDefault(handler, req, bootCopy.Name)
		mutationBoot(&bootCopy.ObjectMeta, req)

		marshaledBoot, err := json.Marshal(bootCopy)
		if err != nil {
			return admission.ErrorResponse(http.StatusInternalServerError, err), err
		}
		return PatchResponseFromRaw(req.AdmissionRequest.Object.Raw, marshaledBoot), nil
	}

	return types.Response{}, nil
}

// PatchResponseFromRaw takes 2 byte arrays and returns a new response with json patch.
// The original object should be passed in as raw bytes to avoid the roundtripping problem
// described in https://github.com/kubernetes-sigs/kubebuilder/issues/510.
// PR https://github.com/kubernetes-sigs/controller-runtime/pull/256
func PatchResponseFromRaw(original, current []byte) types.Response {
	patches, err := jsonpatch.CreatePatch(original, current)
	if err != nil {
		return admission.ErrorResponse(http.StatusInternalServerError, err)
	}
	return types.Response{
		Patches: patches,
		Response: &admissionv1beta1.AdmissionResponse{
			Allowed:   true,
			PatchType: func() *admissionv1beta1.PatchType { pt := admissionv1beta1.PatchTypeJSONPatch; return &pt }(),
		},
	}
}

func mutationDefault(handler *operator.BootHandler, req types.Request, bootName string) {
	mutationDefaulter := false
	ns, found := os.LookupEnv("MUTATION_DEFAULTER")
	if found && ns == "true" {
		mutationDefaulter = true
	}

	if mutationDefaulter {
		changed := handler.DefaultValue()

		//Update the Boot's default Value
		if changed {
			reason := "Updating Boot with Defaulters"
			logger.Info(fmt.Sprintf("%s: [%s/%s]",
				reason, req.AdmissionRequest.Namespace, req.AdmissionRequest.Name),
				"operation", req.AdmissionRequest.Operation)
			handler.EventNormal(reason, bootName)
		}
	}
}

func mutationBoot(metaData *metav1.ObjectMeta, req types.Request) {
	if metaData == nil {
		return
	}

	operation := req.AdmissionRequest.Operation

	if operation == admissionv1beta1.Update {
		metaAnnotation := metaData.Annotations
		if metaAnnotation == nil {
			metaAnnotation = make(map[string]string)
			metaData.Annotations = metaAnnotation
		}

		metaAnnotation[operator.StatusModificationTimeAnnotationKey] = operator.GetCurrentTimestamp()
	}
}

var _ inject.Client = &BootMutator{}

// InjectClient will inject client into BootMutator
func (mHandler *BootMutator) InjectClient(c client.Client) error {
	mHandler.client = c
	return nil
}

var _ inject.Decoder = &BootMutator{}

// InjectDecoder will inject decoder into BootMutator
func (mHandler *BootMutator) InjectDecoder(d types.Decoder) error {
	mHandler.decoder = d
	return nil
}
