package webhook

import (
	"github.com/go-logr/logr"
	v1 "github.com/logancloud/logan-app-operator/pkg/apis/app/v1"
	"github.com/logancloud/logan-app-operator/pkg/logan"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission/builder"
	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"sigs.k8s.io/controller-runtime/pkg/manager"

	bootmutation "github.com/logancloud/logan-app-operator/pkg/logan/webhook/mutation"
	bootvalidation "github.com/logancloud/logan-app-operator/pkg/logan/webhook/validation"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
)

const (
	OperatorName = "logan-app-operator"

	//Webhook Server
	Port        = 8443
	ServerName  = "logan-app-webhook-server"
	CertDir     = "/etc/webhook/cert"
	ServiceName = "logan-app-webhook"

	MutationName    = "mutation.app.logancloud.com"
	MutationCfgName = "logan-app-webhook-mutation"

	ValidationName    = "validation.app.logancloud.com"
	ValidationCfgName = "logan-app-webhook-validation"
)

func RegisterWebhook(mgr manager.Manager, log logr.Logger, operatorNs string) {
	// 1. Create a webhook(mutation)
	mutationHandler := &bootmutation.BootMutator{
		Schema:   mgr.GetScheme(),
		Recorder: mgr.GetRecorder("logan-webhook-mutation"),
	}
	mutationWh, err := builder.NewWebhookBuilder().
		Name(MutationName).
		Mutating().
		Operations(admissionregistrationv1beta1.Create, admissionregistrationv1beta1.Update).
		ForType(&v1.JavaBoot{}).
		Handlers(mutationHandler).
		WithManager(mgr).
		Build()
	if err != nil {
		log.Error(err, "Creating mutation webhook error")
	}

	// 2. Create a webhook(validation)
	rules := admissionregistrationv1beta1.RuleWithOperations{
		Operations: []admissionregistrationv1beta1.OperationType{
			admissionregistrationv1beta1.Create,
			admissionregistrationv1beta1.Update},
		Rule: admissionregistrationv1beta1.Rule{
			APIGroups:   []string{"app.logancloud.com"},
			APIVersions: []string{"v1"},
			Resources:   []string{"javaboots", "phpboots", "pythonboots", "nodejsboots", "webboots"}},
	}
	validationHandler := &bootvalidation.BootValidator{}
	validationWh, err := builder.NewWebhookBuilder().
		Name(ValidationName).
		Validating().
		Operations(admissionregistrationv1beta1.Create, admissionregistrationv1beta1.Update).
		//ForType(&v1.JavaBoot{}).
		Rules(rules).
		Handlers(validationHandler).
		WithManager(mgr).
		Build()
	if err != nil {
		log.Error(err, "Creating validation webhook error")
	}

	// Create a server(mutation)
	whServer, err := webhook.NewServer(ServerName, mgr, webhook.ServerOptions{
		Port:             Port,
		CertDir:          CertDir,
		BootstrapOptions: getBootstrapOption(operatorNs, log),
	})
	if err != nil {
		log.Error(err, "Creating webhook server error")
	}

	err = whServer.Register(mutationWh, validationWh)
	if err != nil {
		log.Error(err, "Registering webhook error")
	}
}

func getBootstrapOption(operatorNs string, log logr.Logger) *webhook.BootstrapOptions {
	operatorName := OperatorName
	svcName := ServiceName
	mutationName := MutationCfgName
	validationName := ValidationCfgName

	if logan.OperDev == "dev" || logan.OperDev == "auto" {
		operatorName = operatorName + "-" + logan.OperDev
		svcName = svcName + "-" + logan.OperDev
		mutationName = mutationName + "-" + logan.OperDev
		validationName = validationName + "-" + logan.OperDev
	}

	log.Info("Register Webhook info",
		"operatorName", operatorName, "svcName", svcName,
		"mutationName", mutationName, "validationName", validationName)

	return &webhook.BootstrapOptions{
		MutatingWebhookConfigName:   mutationName,
		ValidatingWebhookConfigName: validationName,

		Service: &webhook.Service{
			Namespace: operatorNs,
			Name:      svcName,
			Selectors: map[string]string{
				"name": operatorName,
			},
		},
	}
}
