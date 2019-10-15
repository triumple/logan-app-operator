package webhook

import (
	"github.com/go-logr/logr"
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
	operatorName = "logan-app-operator"

	//Webhook Server
	port        = 8443
	serverName  = "logan-app-webhook-server"
	certDir     = "/etc/webhook/cert"
	serviceName = "logan-app-webhook"

	mutationName    = "mutation.app.logancloud.com"
	mutationCfgName = "logan-app-webhook-mutation"

	validationName       = "validation.app.logancloud.com"
	validationConfigName = "config.validation.app.logancloud.com"

	validationCfgName = "logan-app-webhook-validation"
)

// RegisterWebhook will register webhook for mutation and validation
func RegisterWebhook(mgr manager.Manager, log logr.Logger, operatorNs string) {
	rules := admissionregistrationv1beta1.RuleWithOperations{
		Operations: []admissionregistrationv1beta1.OperationType{
			admissionregistrationv1beta1.Create,
			admissionregistrationv1beta1.Update,
			//admissionregistrationv1beta1.Delete,
		},
		Rule: admissionregistrationv1beta1.Rule{
			APIGroups:   []string{"app.logancloud.com"},
			APIVersions: []string{"v1"},
			Resources:   []string{"javaboots", "phpboots", "pythonboots", "nodejsboots", "webboots"}},
	}

	// 1. Create a webhook(boot mutation)
	mutationHandler := &bootmutation.BootMutator{
		Schema:   mgr.GetScheme(),
		Recorder: mgr.GetRecorder("logan-webhook-mutation"),
	}
	mutationWh, err := builder.NewWebhookBuilder().
		Name(mutationName).
		Mutating().
		Rules(rules).
		Handlers(mutationHandler).
		WithManager(mgr).
		Build()
	if err != nil {
		log.Error(err, "Creating boot mutation webhook error")
	}

	// 2. Create a webhook(boot validation)
	validationHandler := &bootvalidation.BootValidator{
		Schema:   mgr.GetScheme(),
		Recorder: mgr.GetRecorder("logan-webhook-validation"),
	}
	validationWh, err := builder.NewWebhookBuilder().
		Name(validationName).
		Validating().
		Rules(rules).
		Handlers(validationHandler).
		WithManager(mgr).
		Build()
	if err != nil {
		log.Error(err, "Creating boot validation webhook error")
	}

	// 3. Create a webhook(config validation)

	configRules := admissionregistrationv1beta1.RuleWithOperations{
		Operations: []admissionregistrationv1beta1.OperationType{
			admissionregistrationv1beta1.Create,
			admissionregistrationv1beta1.Update,
			admissionregistrationv1beta1.Delete},
		Rule: admissionregistrationv1beta1.Rule{
			APIGroups:   []string{""},
			APIVersions: []string{"v1"},
			Resources:   []string{"configmaps"}},
	}

	validationConfigHandler := &bootvalidation.ConfigValidator{
		OperatorNamespace: operatorNs,
	}
	validationConfigWh, err := builder.NewWebhookBuilder().
		Name(validationConfigName).
		Validating().
		Rules(configRules).
		Handlers(validationConfigHandler).
		WithManager(mgr).
		Build()
	if err != nil {
		log.Error(err, "Creating config validation webhook error")
	}

	// Create a server
	whServer, err := webhook.NewServer(serverName, mgr, webhook.ServerOptions{
		Port:             port,
		CertDir:          certDir,
		BootstrapOptions: getBootstrapOption(operatorNs, log),
	})
	if err != nil {
		log.Error(err, "Creating webhook server error")
	}

	err = whServer.Register(mutationWh, validationWh, validationConfigWh)
	if err != nil {
		log.Error(err, "Registering webhook error")
	}
}

func getBootstrapOption(operatorNs string, log logr.Logger) *webhook.BootstrapOptions {
	operatorName := operatorName
	svcName := serviceName
	mutationName := mutationCfgName
	validationName := validationCfgName

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
