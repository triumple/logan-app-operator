package mutation

import (
	"context"
	"fmt"
	v1 "github.com/logancloud/logan-app-operator/pkg/apis/app/v1"
	"github.com/logancloud/logan-app-operator/pkg/controller/javaboot"
	"github.com/logancloud/logan-app-operator/pkg/logan/operator"
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

func (mHandler *BootMutator) Handle(ctx context.Context, req types.Request) types.Response {
	boot := &v1.JavaBoot{}

	err := mHandler.decoder.Decode(req, boot)
	if err != nil {
		logger.Error(err, "Decoding request error")
		return admission.ErrorResponse(http.StatusBadRequest, err)
	}
	bootCopy := boot.DeepCopy()
	if operator.Ignore(req.AdmissionRequest.Namespace) {
		return admission.PatchResponse(boot, bootCopy)
	}

	ns, found := os.LookupEnv("MUTATION_DEFAULTER")

	if found && ns == "true" {
		err = mHandler.mutateBoot(ctx, bootCopy)
		if err != nil {
			logger.Error(err, "mutate error")
			return admission.ErrorResponse(http.StatusInternalServerError, err)
		}

		logger.Info(fmt.Sprintf("Boot Defaulters: [%s/%s]",
			req.AdmissionRequest.Namespace, req.AdmissionRequest.Name),
			"operation", req.AdmissionRequest.Operation)
	}

	logger.Info(fmt.Sprintf("Successfully Mutating Boot: [%s/%s]",
		req.AdmissionRequest.Namespace, req.AdmissionRequest.Name),
		"operation", req.AdmissionRequest.Operation)

	return admission.PatchResponse(boot, bootCopy)
}

// mutateBoot add an annotation into Boot
func (mHandler *BootMutator) mutateBoot(ctx context.Context, javaBoot *v1.JavaBoot) error {
	c := mHandler.client
	scheme := mHandler.Schema
	recorder := mHandler.Recorder
	handler := javaboot.InitHandler(javaBoot, scheme, c, logger, recorder)

	changed := handler.DefaultValue()

	//Update the Boot's default Value
	if changed {
		reason := "Updating Boot with Defaulters"
		logger.Info(reason)
		handler.EventNormal(reason, javaBoot.Name)
	}

	javaBoot.Annotations[operator.MutationAnnotationKey] = "true"

	return nil
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
