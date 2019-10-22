package keys

const (
	// PrometheusPathAnnotationKey is the Boot's created service's prometheus path annotation key
	PrometheusPathAnnotationKey = "prometheus.io/path"
	// PrometheusPathAnnotationValue is the Boot's created service's prometheus path annotation value
	PrometheusPathAnnotationValue = "/metrics"

	// PrometheusPortAnnotationKey is the Boot's created service's prometheus port annotation key
	PrometheusPortAnnotationKey = "prometheus.io/port"

	// PrometheusSchemeAnnotationKey is the Boot's created service's prometheus scheme annotation key
	PrometheusSchemeAnnotationKey = "prometheus.io/scheme"
	// PrometheusSchemeAnnotationValue is the Boot's created service's prometheus scheme annotation value
	PrometheusSchemeAnnotationValue = "http"

	// PrometheusScrapeAnnotationKey is the Boot's created service's prometheus scrape annotation key
	PrometheusScrapeAnnotationKey = "prometheus.io/scrape"
	// PrometheusScrapeAnnotationValue is the Boot's created service's prometheus scrape annotation value
	PrometheusScrapeAnnotationValue = "true"

	// EnvAnnotationKey is the annotation key for storing when changed env
	EnvAnnotationKey = "app.logancloud.com/env"
	// EnvAnnotationValue is default value for for env
	EnvAnnotationValue = "generated"
	// BootEnvsAnnotationKey is the annotation key for storing previous envs
	BootEnvsAnnotationKey = "app.logancloud.com/boot-envs"
	// BootPvcsAnnotationKey is the annotation key for storing previous pvcs
	BootPvcsAnnotationKey = "app.logancloud.com/boot-pvcs"
	// BootDeployPvcsAnnotationKey is the annotation key for storing previous deploy pvcs
	BootDeployPvcsAnnotationKey = "app.logancloud.com/boot-deploy-pvcs"
	// BootImagesAnnotationKey is the annotation key for storing previous images
	BootImagesAnnotationKey = "app.logancloud.com/boot-images"
	// BootRestartedAtAnnotationKey is the annotation key for recording restarted time
	BootRestartedAtAnnotationKey = "app.logancloud.com/restartedAt"

	// DeployAnnotationKey is the annotation key for storing boot's current Deployment name
	DeployAnnotationKey = "app.logancloud.com/deploy"
	// ServicesAnnotationKey is the annotation key for storing boot's current services name list
	ServicesAnnotationKey = "app.logancloud.com/services"
	// AppTypeAnnotationKey is the annotation key for storing boot's type
	AppTypeAnnotationKey = "app.logancloud.com/type"
	// AppTypeAnnotationDeploy is the annotation value for Deployment
	AppTypeAnnotationDeploy = "deploy"

	// StatusAvailableAnnotationKey is the annotation key for storing boot's current pods
	StatusAvailableAnnotationKey = "app.logancloud.com/status.available"
	// StatusDesiredAnnotationKey is the annotation key for storing boot's desired pods
	StatusDesiredAnnotationKey = "app.logancloud.com/status.desired"
	// StatusModificationTimeAnnotationKey is the annotation key for storing boot's type
	StatusModificationTimeAnnotationKey = "app.logancloud.com/status.lastUpdateTimeStamp"

	// BootRevisionIdAnnotationKey is the annotation key for boot revision's ID
	BootRevisionIdAnnotationKey = "app.logancloud.com/revision"
	// BootRevisionHashAnnotationKey is the annotation key for boot revision's boot hash
	BootRevisionHashAnnotationKey = "app.logancloud.com/hash"
	// BootRevisionPhaseAnnotationKey is the annotation key for boot revision's phase
	BootRevisionPhaseAnnotationKey = "app.logancloud.com/phase"
	// BootRevisionDiffAnnotationKey is the annotation key for boot revision's the differences from the previous version
	BootRevisionDiffAnnotationKey = "app.logancloud.com/diff"
	// BootRevisionRetryAnnotationKey is the annotation key for boot revision's fail retry times
	BootRevisionRetryAnnotationKey = "app.logancloud.com/retry"
)
