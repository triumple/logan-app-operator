package javaboot

import (
	"github.com/logancloud/logan-app-operator/pkg/apis"
	javabootv1 "github.com/logancloud/logan-app-operator/pkg/apis/app/v1"
	"github.com/logancloud/logan-app-operator/pkg/logan/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/net/context"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"os"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"strconv"
	"strings"
	"time"
)

var cfg *rest.Config
var t *envtest.Environment
var c client.Client

var stop chan struct{}

const timeout = time.Second * 10

var _ = Describe("JavaBoot", func() {
	var recFn reconcile.Reconciler
	var requests chan reconcile.Request

	BeforeEach(func() {
		//Every case using clean k8s environment, avoid conflict each other.
		logf.SetLogger(logf.ZapLoggerTo(os.Stderr, true)) //Debug Output
		err := config.NewConfigFromString(getConfig())
		//err := config.InitByFile("../../../configs/config.yaml")
		if err != nil {
			log.Error(err, "")
		}

		t = &envtest.Environment{
			CRDDirectoryPaths: []string{filepath.Join("..", "..", "..", "deploy", "crds")},
		}

		err = apis.AddToScheme(scheme.Scheme)
		if err != nil {
			log.Error(err, "")
		}

		if cfg, err = t.Start(); err != nil {
			log.Error(err, "")
		}

		mgr, err := manager.New(cfg, manager.Options{})
		Expect(err).NotTo(HaveOccurred())
		c = mgr.GetClient()

		inner := newReconciler(mgr)

		//requests用于跟踪reconcile是否执行的请求
		requests = make(chan reconcile.Request)
		// delegates to inner and writes the request to requests after Reconcile is finished.
		recFn = reconcile.Func(func(req reconcile.Request) (reconcile.Result, error) {
			result, err := inner.Reconcile(req)
			requests <- req
			return result, err
		})

		err = add(mgr, recFn)
		Expect(err).NotTo(HaveOccurred())

		log.Info("Starting Manager")
		stop = make(chan struct{})
		go func() {
			Expect(mgr.Start(stop)).NotTo(HaveOccurred())
			log.Info("Stoped Manager")
		}()

	})

	AfterEach(func() {
		close(stop)
		log.Info("Stoping Manager")
		err := t.Stop()
		if err != nil {
			log.Error(err, "")
		}
	})

	Describe("can be created deployment and service by javaboot", func() {
		It("testing create", func() {
			res := genResource()

			replicas := int32(1)
			javaboot := &javabootv1.JavaBoot{
				ObjectMeta: metav1.ObjectMeta{Name: res.bootKey.Name, Namespace: res.bootKey.Namespace},
				Spec: javabootv1.BootSpec{
					Replicas: &replicas,
					Image:    "logan-startkit-boot",
					Version:  "1.2.1",
				},
			}

			createBoot(javaboot)
			defer c.Delete(context.TODO(), javaboot)
			wait(requests, res.expectedRequest, 2)
			boot := getBoot(res.bootKey)
			log.Info("get boot", "boot", boot)

			deploy := getDeploy(res.depKey)
			// Delete the Deployment and expect Reconcile to be called for Deployment deletion
			Expect(c.Delete(context.TODO(), deploy)).NotTo(HaveOccurred())
			wait(requests, res.expectedRequest, 2)
		})
	})

	Describe("can update deployment by javaboot ", func() {
		Context("testing update replicas", func() {
			It("testing update replicas", func() {
				res := genResource()

				replicas := int32(3)
				javaboot := &javabootv1.JavaBoot{
					ObjectMeta: metav1.ObjectMeta{Name: res.bootKey.Name, Namespace: res.bootKey.Namespace},
					Spec: javabootv1.BootSpec{
						Replicas: &replicas,
						Port:     8080,
					},
				}

				createBoot(javaboot)
				defer c.Delete(context.TODO(), javaboot)
				wait(requests, res.expectedRequest, 3)

				deploy := getDeploy(res.depKey)

				// check replicas
				Expect(deploy.Spec.Replicas).Should(Equal(&replicas))
				// get boot
				boot := getBoot(res.bootKey)

				//update replicas
				newReplicas := int32(3)
				boot.Spec.Replicas = &newReplicas
				updateBoot(boot)

				//wait(requests, res.expectedRequest, 1)
				//check replicas
				updateDeploy := getDeploy(res.depKey)
				Expect(updateDeploy.Spec.Replicas).Should(Equal(&newReplicas))

			})
		})

		Context("update version", func() {
			It("testing update version", func() {
				res := genResource()
				replicas := int32(3)
				javaboot := &javabootv1.JavaBoot{
					ObjectMeta: metav1.ObjectMeta{Name: res.bootKey.Name, Namespace: res.bootKey.Namespace},
					Spec: javabootv1.BootSpec{
						Replicas: &replicas,
						Port:     8080,
						Image:    "logan-startkit-boot",
						Version:  "V1.0.0",
					},
				}

				createBoot(javaboot)
				defer c.Delete(context.TODO(), javaboot)
				wait(requests, res.expectedRequest, 3)

				deploy := getDeploy(res.depKey)
				// check version
				image := config.JavaConfig.AppSpec.Settings.Registry + "/" + javaboot.Spec.Image + ":" + javaboot.Spec.Version
				Expect(deploy.Spec.Template.Spec.Containers[0].Image).Should(Equal(image))

				// get boot
				boot := getBoot(res.bootKey)

				//update version
				boot.Spec.Version = "V1.0.1"
				updateBoot(boot)
				wait(requests, res.expectedRequest, 2)

				//check image
				updateDeploy := getDeploy(res.depKey)

				updateImages := config.JavaConfig.AppSpec.Settings.Registry + "/" + boot.Spec.Image + ":" + boot.Spec.Version
				Expect(updateDeploy.Spec.Template.Spec.Containers[0].Image).Should(Equal(updateImages))
			})

			It("testing update image", func() {
				res := genResource()
				replicas := int32(3)
				javaboot := &javabootv1.JavaBoot{
					ObjectMeta: metav1.ObjectMeta{Name: res.bootKey.Name, Namespace: res.bootKey.Namespace},
					Spec: javabootv1.BootSpec{
						Replicas: &replicas,
						Port:     8080,
						Image:    "logan-startkit-boot",
						Version:  "V1.0.0",
					},
				}
				createBoot(javaboot)
				defer c.Delete(context.TODO(), javaboot)
				wait(requests, res.expectedRequest, 3)

				deploy := getDeploy(res.depKey)

				// check image
				image := config.JavaConfig.AppSpec.Settings.Registry + "/" + javaboot.Spec.Image + ":" + javaboot.Spec.Version
				Expect(deploy.Spec.Template.Spec.Containers[0].Image).Should(Equal(image))

				// get boot
				boot := getBoot(res.bootKey)

				//update image
				boot.Spec.Image = "logan-startkit-boot_new"
				updateBoot(boot)

				wait(requests, res.expectedRequest, 1)

				//check image
				updateDeploy := getDeploy(res.depKey)

				updateImages := config.JavaConfig.AppSpec.Settings.Registry + "/" + boot.Spec.Image + ":" + boot.Spec.Version
				Expect(updateDeploy.Spec.Template.Spec.Containers[0].Image).Should(Equal(updateImages))

			})
		})

		Context("testing update port", func() {
			It("testing update port", func() {
				res := genResource()
				replicas := int32(3)
				javaboot := &javabootv1.JavaBoot{
					ObjectMeta: metav1.ObjectMeta{Name: res.bootKey.Name, Namespace: res.bootKey.Namespace},
					Spec: javabootv1.BootSpec{
						Replicas: &replicas,
						Port:     8080,
					},
				}

				createBoot(javaboot)
				defer c.Delete(context.TODO(), javaboot)
				wait(requests, res.expectedRequest, 3)

				deploy := getDeploy(res.depKey)
				svc := getService(res.serviceKey)

				// check ports
				Expect(svc.Spec.Ports[0].Port).Should(Equal(javaboot.Spec.Port))
				Expect(svc.Annotations["prometheus.io/port"]).Should(Equal(strconv.Itoa(8080)))
				Expect(deploy.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort).Should(Equal(javaboot.Spec.Port))

				// get boot
				boot := getBoot(res.bootKey)

				//update port
				boot.Spec.Port = 8081
				updateBoot(boot)

				wait(requests, res.expectedRequest, 2)

				//check port
				updateDeploy := getDeploy(res.depKey)
				updatesvc := getService(res.serviceKey)
				Expect(updatesvc.Spec.Ports[0].Port).Should(Equal(boot.Spec.Port))
				Expect(updatesvc.Annotations["prometheus.io/port"]).Should(Equal("8081"))
				Expect(updateDeploy.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort).Should(Equal(boot.Spec.Port))
				Expect(updateDeploy.Spec.Template.Spec.Containers[0].ReadinessProbe.HTTPGet.Port.IntVal).Should(Equal(boot.Spec.Port))
				Expect(updateDeploy.Spec.Template.Spec.Containers[0].LivenessProbe.HTTPGet.Port.IntVal).Should(Equal(boot.Spec.Port))

			})
		})

		Context("testing update resources", func() {
			It("testing scale up cpu and memory", func() {
				res := genResource()
				replicas := int32(3)

				resources := &corev1.ResourceRequirements{
					Limits:   map[corev1.ResourceName]resource.Quantity{},
					Requests: map[corev1.ResourceName]resource.Quantity{},
				}

				resources.Limits[corev1.ResourceMemory] = *resource.NewMilliQuantity(2048, resource.BinarySI)
				resources.Limits[corev1.ResourceCPU] = *resource.NewQuantity(2, resource.DecimalSI)

				resources.Requests[corev1.ResourceMemory] = *resource.NewMilliQuantity(1024, resource.BinarySI)
				resources.Requests[corev1.ResourceCPU] = *resource.NewQuantity(1, resource.DecimalSI)

				javaboot := &javabootv1.JavaBoot{
					ObjectMeta: metav1.ObjectMeta{Name: res.bootKey.Name, Namespace: res.bootKey.Namespace},
					Spec: javabootv1.BootSpec{
						Replicas:  &replicas,
						Port:      8080,
						Resources: *resources,
					},
				}

				createBoot(javaboot)
				defer c.Delete(context.TODO(), javaboot)
				wait(requests, res.expectedRequest, 3)

				deploy := getDeploy(res.depKey)

				// check resource
				Expect(deploy.Spec.Template.Spec.Containers[0].Resources.Limits.Memory()).Should(Equal(resources.Limits.Memory()))
				Expect(deploy.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu()).Should(Equal(resources.Limits.Cpu()))

				Expect(deploy.Spec.Template.Spec.Containers[0].Resources.Requests.Memory()).Should(Equal(resources.Requests.Memory()))
				Expect(deploy.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu()).Should(Equal(resources.Requests.Cpu()))

				// get boot
				boot := getBoot(res.bootKey)

				//update resource

				updateResources := &corev1.ResourceRequirements{
					Limits:   map[corev1.ResourceName]resource.Quantity{},
					Requests: map[corev1.ResourceName]resource.Quantity{},
				}

				updateResources.Limits[corev1.ResourceMemory] = *resource.NewMilliQuantity(2048*2, resource.BinarySI)
				updateResources.Limits[corev1.ResourceCPU] = *resource.NewQuantity(2*2, resource.DecimalSI)

				updateResources.Requests[corev1.ResourceMemory] = *resource.NewMilliQuantity(1024*2, resource.BinarySI)
				updateResources.Requests[corev1.ResourceCPU] = *resource.NewQuantity(1*2, resource.DecimalSI)

				boot.Spec.Resources = *updateResources

				updateBoot(boot)
				wait(requests, res.expectedRequest, 1)

				//check resource
				updateDeploy := getDeploy(res.depKey)
				Expect(updateDeploy.Spec.Template.Spec.Containers[0].Resources.Limits.Memory()).Should(Equal(updateResources.Limits.Memory()))
				Expect(updateDeploy.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu()).Should(Equal(updateResources.Limits.Cpu()))

				Expect(updateDeploy.Spec.Template.Spec.Containers[0].Resources.Requests.Memory()).Should(Equal(updateResources.Requests.Memory()))
				Expect(updateDeploy.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu()).Should(Equal(updateResources.Requests.Cpu()))

			})

			It("testing scale down cpu and memory", func() {
				res := genResource()
				replicas := int32(3)

				resources := &corev1.ResourceRequirements{
					Limits:   map[corev1.ResourceName]resource.Quantity{},
					Requests: map[corev1.ResourceName]resource.Quantity{},
				}

				resources.Limits[corev1.ResourceMemory] = *resource.NewMilliQuantity(2048*2, resource.BinarySI)
				resources.Limits[corev1.ResourceCPU] = *resource.NewQuantity(2*2, resource.DecimalSI)

				resources.Requests[corev1.ResourceMemory] = *resource.NewMilliQuantity(1024*2, resource.BinarySI)
				resources.Requests[corev1.ResourceCPU] = *resource.NewQuantity(1*2, resource.DecimalSI)

				javaboot := &javabootv1.JavaBoot{
					ObjectMeta: metav1.ObjectMeta{Name: res.bootKey.Name, Namespace: res.bootKey.Namespace},
					Spec: javabootv1.BootSpec{
						Replicas:  &replicas,
						Port:      8080,
						Resources: *resources,
					},
				}

				createBoot(javaboot)
				defer c.Delete(context.TODO(), javaboot)
				wait(requests, res.expectedRequest, 3)

				deploy := getDeploy(res.depKey)

				// check resource
				Expect(deploy.Spec.Template.Spec.Containers[0].Resources.Limits.Memory()).Should(Equal(resources.Limits.Memory()))
				Expect(deploy.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu()).Should(Equal(resources.Limits.Cpu()))

				Expect(deploy.Spec.Template.Spec.Containers[0].Resources.Requests.Memory()).Should(Equal(resources.Requests.Memory()))
				Expect(deploy.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu()).Should(Equal(resources.Requests.Cpu()))

				// get boot
				boot := getBoot(res.bootKey)

				//update resource

				updateResources := &corev1.ResourceRequirements{
					Limits:   map[corev1.ResourceName]resource.Quantity{},
					Requests: map[corev1.ResourceName]resource.Quantity{},
				}

				updateResources.Limits[corev1.ResourceMemory] = *resource.NewMilliQuantity(2048, resource.BinarySI)
				updateResources.Limits[corev1.ResourceCPU] = *resource.NewQuantity(2, resource.DecimalSI)

				updateResources.Requests[corev1.ResourceMemory] = *resource.NewMilliQuantity(1024, resource.BinarySI)
				updateResources.Requests[corev1.ResourceCPU] = *resource.NewQuantity(1, resource.DecimalSI)

				boot.Spec.Resources = *updateResources

				updateBoot(boot)
				wait(requests, res.expectedRequest, 1)

				//check resource
				updateDeploy := getDeploy(res.depKey)
				Expect(updateDeploy.Spec.Template.Spec.Containers[0].Resources.Limits.Memory()).Should(Equal(updateResources.Limits.Memory()))
				Expect(updateDeploy.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu()).Should(Equal(updateResources.Limits.Cpu()))

				Expect(updateDeploy.Spec.Template.Spec.Containers[0].Resources.Requests.Memory()).Should(Equal(updateResources.Requests.Memory()))
				Expect(updateDeploy.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu()).Should(Equal(updateResources.Requests.Cpu()))

			})

			It("testing  cpu and memory Limits lager than Requests", func() {
				res := genResource()
				replicas := int32(3)

				resources := &corev1.ResourceRequirements{
					Limits:   map[corev1.ResourceName]resource.Quantity{},
					Requests: map[corev1.ResourceName]resource.Quantity{},
				}

				resources.Limits[corev1.ResourceMemory] = *resource.NewMilliQuantity(2048, resource.BinarySI)
				resources.Limits[corev1.ResourceCPU] = *resource.NewQuantity(2, resource.DecimalSI)

				resources.Requests[corev1.ResourceMemory] = *resource.NewMilliQuantity(1024, resource.BinarySI)
				resources.Requests[corev1.ResourceCPU] = *resource.NewQuantity(1, resource.DecimalSI)

				javaboot := &javabootv1.JavaBoot{
					ObjectMeta: metav1.ObjectMeta{Name: res.bootKey.Name, Namespace: res.bootKey.Namespace},
					Spec: javabootv1.BootSpec{
						Replicas:  &replicas,
						Port:      8080,
						Resources: *resources,
					},
				}

				createBoot(javaboot)
				defer c.Delete(context.TODO(), javaboot)
				wait(requests, res.expectedRequest, 3)

				deploy := getDeploy(res.depKey)

				// check resource
				Expect(deploy.Spec.Template.Spec.Containers[0].Resources.Limits.Memory()).Should(Equal(resources.Limits.Memory()))
				Expect(deploy.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu()).Should(Equal(resources.Limits.Cpu()))

				Expect(deploy.Spec.Template.Spec.Containers[0].Resources.Requests.Memory()).Should(Equal(resources.Requests.Memory()))
				Expect(deploy.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu()).Should(Equal(resources.Requests.Cpu()))

				// get boot
				boot := getBoot(res.bootKey)

				//update resource

				updateResources := &corev1.ResourceRequirements{
					Limits:   map[corev1.ResourceName]resource.Quantity{},
					Requests: map[corev1.ResourceName]resource.Quantity{},
				}

				updateResources.Limits[corev1.ResourceMemory] = *resource.NewMilliQuantity(2048*2, resource.BinarySI)
				updateResources.Limits[corev1.ResourceCPU] = *resource.NewQuantity(2*2, resource.DecimalSI)

				updateResources.Requests[corev1.ResourceMemory] = *resource.NewMilliQuantity(1024*6, resource.BinarySI)
				updateResources.Requests[corev1.ResourceCPU] = *resource.NewQuantity(1*6, resource.DecimalSI)

				boot.Spec.Resources = *updateResources

				updateBoot(boot)
				wait(requests, res.expectedRequest, 1)

				//check resource
				updateDeploy := getDeploy(res.depKey)

				Expect(updateDeploy.Spec.Template.Spec.Containers[0].Resources.Limits.Memory()).Should(Equal(updateResources.Limits.Memory()))
				Expect(updateDeploy.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu()).Should(Equal(updateResources.Limits.Cpu()))

				// Requests should equal or less than Limits
				Expect(updateDeploy.Spec.Template.Spec.Containers[0].Resources.Requests.Memory()).Should(Equal(updateResources.Limits.Memory()))
				Expect(updateDeploy.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu()).Should(Equal(updateResources.Limits.Cpu()))

			})
		})

		Context("testing nodeSelector", func() {
			It("testing update nodeSelector", func() {
				res := genResource()
				replicas := int32(3)
				javaboot := &javabootv1.JavaBoot{
					ObjectMeta: metav1.ObjectMeta{Name: res.bootKey.Name, Namespace: res.bootKey.Namespace},
					Spec: javabootv1.BootSpec{
						Replicas:     &replicas,
						Port:         8080,
						NodeSelector: map[string]string{"env": "dev", "app": "myAPPName"},
					},
				}

				createBoot(javaboot)
				defer c.Delete(context.TODO(), javaboot)
				wait(requests, res.expectedRequest, 3)

				deploy := getDeploy(res.depKey)

				// check NodeSelector
				Expect(deploy.Spec.Template.Spec.NodeSelector).Should(Equal(javaboot.Spec.NodeSelector))

				// get boot
				boot := getBoot(res.bootKey)

				//update NodeSelector
				boot.Spec.NodeSelector = map[string]string{"env": "test", "app": "myAPPName2", "new": "new_label"}
				updateBoot(boot)

				wait(requests, res.expectedRequest, 1)

				//check NodeSelector
				updateDeploy := getDeploy(res.depKey)
				Expect(updateDeploy.Spec.Template.Spec.NodeSelector).Should(Equal(boot.Spec.NodeSelector))

			})
		})

		Context("testing health", func() {
			It("testing update health", func() {
				res := genResource()
				replicas := int32(3)
				health := "/health"
				javaboot := &javabootv1.JavaBoot{
					ObjectMeta: metav1.ObjectMeta{Name: res.bootKey.Name, Namespace: res.bootKey.Namespace},
					Spec: javabootv1.BootSpec{
						Replicas: &replicas,
						Port:     8080,
						Health:   &health,
					},
				}

				createBoot(javaboot)
				defer c.Delete(context.TODO(), javaboot)
				wait(requests, res.expectedRequest, 3)

				deploy := getDeploy(res.depKey)

				// check health
				//Expect(deploy.Spec.Template.Spec.NodeSelector).Should(Equal(javaboot.Spec.NodeSelector))
				Expect(deploy.Spec.Template.Spec.Containers[0].LivenessProbe.HTTPGet.Path).Should(Equal(health))
				Expect(deploy.Spec.Template.Spec.Containers[0].ReadinessProbe.HTTPGet.Path).Should(Equal(health))

				// get boot
				boot := getBoot(res.bootKey)

				//update health
				health2 := "/health2"
				boot.Spec.Health = &health2
				updateBoot(boot)

				wait(requests, res.expectedRequest, 1)

				//check health
				updateDeploy := getDeploy(res.depKey)

				Expect(updateDeploy.Spec.Template.Spec.Containers[0].LivenessProbe.HTTPGet.Path).Should(Equal(health2))
				Expect(updateDeploy.Spec.Template.Spec.Containers[0].ReadinessProbe.HTTPGet.Path).Should(Equal(health2))

			})
		})

		Context("testing prometheusScrape", func() {
			It("testing update prometheusScrape", func() {
				res := genResource()
				replicas := int32(3)

				javaboot := &javabootv1.JavaBoot{
					ObjectMeta: metav1.ObjectMeta{Name: res.bootKey.Name, Namespace: res.bootKey.Namespace},
					Spec: javabootv1.BootSpec{
						Replicas:   &replicas,
						Port:       8080,
						Prometheus: "true",
					},
				}

				createBoot(javaboot)
				defer c.Delete(context.TODO(), javaboot)
				wait(requests, res.expectedRequest, 3)

				svr := getService(res.serviceKey)
				Expect(len(svr.Annotations)).Should(Equal(4))
				// get boot
				boot := getBoot(res.bootKey)

				//update
				boot.Spec.Prometheus = "false"
				updateBoot(boot)

				wait(requests, res.expectedRequest, 1)

				//check health
				updateSvr := getService(res.serviceKey)
				Expect(len(updateSvr.Annotations)).Should(Equal(0))
			})
		})

		Context("testing env", func() {
			It("testing update env simple", func() {
				res := genResource()
				//logan.OperatorUID="local"
				replicas := int32(3)
				javaboot := &javabootv1.JavaBoot{
					ObjectMeta: metav1.ObjectMeta{Name: res.bootKey.Name, Namespace: res.bootKey.Namespace},
					Spec: javabootv1.BootSpec{
						Replicas: &replicas,
						Port:     8080,
						Env: []corev1.EnvVar{
							{Name: "key1", Value: "value1"},
							{Name: "key2", Value: "value2"},
							{Name: "myApp", Value: "${APP}"},
							{Name: "myEnv", Value: "${ENV}"},
						},
					},
				}

				createBoot(javaboot)
				defer c.Delete(context.TODO(), javaboot)
				wait(requests, res.expectedRequest, 3)

				deploy := getDeploy(res.depKey)

				// check env
				for _, i := range javaboot.Spec.Env {
					if strings.EqualFold(i.Name, "myAPP") {
						i.Value = res.bootKey.Name
					}
					if strings.EqualFold(i.Name, "myEnv") {
						i.Value = "test"
					}

					for _, j := range deploy.Spec.Template.Spec.Containers[0].Env {
						if strings.EqualFold(i.Name, j.Name) {
							Expect(i).Should(Equal(j))
						}
					}
				}

				// get boot
				boot := getBoot(res.bootKey)

				//update env
				boot.Spec.Env = []corev1.EnvVar{
					{Name: "key1", Value: "value2"},
					{Name: "key5", Value: "value1"},
					{Name: "myApp", Value: "${APP}"},
					{Name: "myEnv", Value: "${ENV}"},
				}

				updateBoot(boot)

				wait(requests, res.expectedRequest, 2)
				//check env
				updateDeploy := getDeploy(res.depKey)

				for _, i := range boot.Spec.Env {
					if strings.EqualFold(i.Name, "myAPP") {
						i.Value = res.bootKey.Name
					}
					if strings.EqualFold(i.Name, "myEnv") {
						i.Value = "test"
					}
					for _, j := range updateDeploy.Spec.Template.Spec.Containers[0].Env {
						if strings.EqualFold(i.Name, j.Name) {
							Expect(i).Should(Equal(j))
						}
					}
				}

			})

			It("testing update env on runtime", func() {
				res := genResource()

				replicas := int32(3)
				javaboot := &javabootv1.JavaBoot{
					ObjectMeta: metav1.ObjectMeta{Name: res.bootKey.Name, Namespace: res.bootKey.Namespace},
					Spec: javabootv1.BootSpec{
						Replicas: &replicas,
						Port:     8080,
					},
				}

				createBoot(javaboot)
				defer c.Delete(context.TODO(), javaboot)
				wait(requests, res.expectedRequest, 3)
				deploy := getDeploy(res.depKey)

				// check replicas
				Expect(deploy.Spec.Replicas).Should(Equal(&replicas))

				// get boot
				boot := getBoot(res.bootKey)

				log.Info("myenv", "env", config.JavaConfig.AppSpec.Env)
				config.JavaConfig.AppSpec.Env[0].Value = "false"
				//update replicas
				updateBoot(boot)

				wait(requests, res.expectedRequest, 1)

				//check replicas
				updateDeploy := getDeploy(res.depKey)

				//still old env
				Expect(updateDeploy.Spec.Template.Spec.Containers[0].Env[0].Value).Should(Equal("true"))

			})

		})

		Context("testing update ownerReferences", func() {
			It("testing set ownerReferences", func() {
				res := genResource()
				replicas := int32(3)
				javaboot := &javabootv1.JavaBoot{
					ObjectMeta: metav1.ObjectMeta{Name: res.bootKey.Name, Namespace: res.bootKey.Namespace},
					Spec: javabootv1.BootSpec{
						Replicas: &replicas,
						Port:     8080,
					},
				}

				createBoot(javaboot)
				defer c.Delete(context.TODO(), javaboot)
				wait(requests, res.expectedRequest, 3)

				deploy := getDeploy(res.depKey)
				svc := getService(res.serviceKey)

				// check
				Expect(deploy.Spec.Replicas).Should(Equal(&replicas))
				Expect(deploy.ObjectMeta.OwnerReferences[0].Kind).Should(Equal("JavaBoot"))
				Expect(deploy.ObjectMeta.OwnerReferences[0].APIVersion).Should(Equal("app.logancloud.com/v1"))
				Expect(*deploy.ObjectMeta.OwnerReferences[0].BlockOwnerDeletion).Should(Equal(true))
				Expect(svc.ObjectMeta.OwnerReferences[0].Kind).Should(Equal("JavaBoot"))
				Expect(svc.ObjectMeta.OwnerReferences[0].APIVersion).Should(Equal("app.logancloud.com/v1"))
				Expect(*svc.ObjectMeta.OwnerReferences[0].BlockOwnerDeletion).Should(Equal(true))

				// get boot
				boot := getBoot(res.bootKey)

				//update
				newReplicas := int32(3)
				boot.Spec.Replicas = &newReplicas
				updateBoot(boot)

				//wait(requests, res.expectedRequest, 1)

				//check
				updateDeploy := getDeploy(res.depKey)
				updateSvc := getService(res.serviceKey)
				Expect(updateDeploy.Spec.Replicas).Should(Equal(&newReplicas))
				Expect(updateDeploy.ObjectMeta.OwnerReferences[0].Kind).Should(Equal("JavaBoot"))
				Expect(updateDeploy.ObjectMeta.OwnerReferences[0].APIVersion).Should(Equal("app.logancloud.com/v1"))
				Expect(*updateDeploy.ObjectMeta.OwnerReferences[0].BlockOwnerDeletion).Should(Equal(true))
				Expect(updateSvc.ObjectMeta.OwnerReferences[0].Kind).Should(Equal("JavaBoot"))
				Expect(updateSvc.ObjectMeta.OwnerReferences[0].APIVersion).Should(Equal("app.logancloud.com/v1"))
				Expect(*updateSvc.ObjectMeta.OwnerReferences[0].BlockOwnerDeletion).Should(Equal(true))

			})
		})
	})

	Describe("can not update deployment by deployment", func() {
		Context("can not update deployment replicas by deployment", func() {
			It("can not update deployment replicas", func() {
				res := genResource()

				replicas := int32(3)
				javaboot := &javabootv1.JavaBoot{
					ObjectMeta: metav1.ObjectMeta{Name: res.bootKey.Name, Namespace: res.bootKey.Namespace},
					Spec: javabootv1.BootSpec{
						Replicas: &replicas,
						Port:     8080,
					},
				}

				createBoot(javaboot)
				defer c.Delete(context.TODO(), javaboot)
				wait(requests, res.expectedRequest, 3)

				deploy := getDeploy(res.depKey)

				// check replicas
				Expect(deploy.Spec.Replicas).Should(Equal(&replicas))

				newReplicas := int32(1)
				deploy.Spec.Replicas = &newReplicas

				updateDeploy(deploy)

				wait(requests, res.expectedRequest, 1)

				//check replicas
				updateDeploy := getDeploy(res.depKey)
				Expect(updateDeploy.Spec.Replicas).Should(Equal(&replicas))
			})
		})

		Context("can not update deployment version by deployment", func() {
			It("can not update deployment version", func() {
				res := genResource()

				replicas := int32(3)
				javaboot := &javabootv1.JavaBoot{
					ObjectMeta: metav1.ObjectMeta{Name: res.bootKey.Name, Namespace: res.bootKey.Namespace},
					Spec: javabootv1.BootSpec{
						Replicas: &replicas,
						Port:     8080,
						Image:    "logan-startkit-boot",
						Version:  "V1.0.0",
					},
				}

				createBoot(javaboot)
				defer c.Delete(context.TODO(), javaboot)
				wait(requests, res.expectedRequest, 3)

				deploy := getDeploy(res.depKey)
				// check version
				image := config.JavaConfig.AppSpec.Settings.Registry + "/" + javaboot.Spec.Image + ":" + javaboot.Spec.Version
				Expect(deploy.Spec.Template.Spec.Containers[0].Image).Should(Equal(image))

				deploy.Spec.Template.Spec.Containers[0].Image = "myImages"
				updateDeploy(deploy)

				wait(requests, res.expectedRequest, 1)

				//check replicas
				updateDeploy := getDeploy(res.depKey)

				updateImages := config.JavaConfig.AppSpec.Settings.Registry + "/" + javaboot.Spec.Image + ":" + javaboot.Spec.Version
				Expect(updateDeploy.Spec.Template.Spec.Containers[0].Image).Should(Equal(updateImages))
			})
		})

		Context("can not update deployment port by deployment", func() {
			It("can not update deployment port", func() {
				res := genResource()
				replicas := int32(3)
				javaboot := &javabootv1.JavaBoot{
					ObjectMeta: metav1.ObjectMeta{Name: res.bootKey.Name, Namespace: res.bootKey.Namespace},
					Spec: javabootv1.BootSpec{
						Replicas: &replicas,
						Port:     8080,
					},
				}

				createBoot(javaboot)
				defer c.Delete(context.TODO(), javaboot)
				wait(requests, res.expectedRequest, 3)

				deploy := getDeploy(res.depKey)
				svc := getService(res.serviceKey)

				// check ports
				Expect(svc.Spec.Ports[0].Port).Should(Equal(javaboot.Spec.Port))
				Expect(svc.Annotations["prometheus.io/port"]).Should(Equal(strconv.Itoa(8080)))
				Expect(deploy.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort).Should(Equal(javaboot.Spec.Port))

				deploy.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort = 8081
				deploy.Spec.Template.Spec.Containers[0].ReadinessProbe.HTTPGet.Port.IntVal = 8081
				deploy.Spec.Template.Spec.Containers[0].LivenessProbe.HTTPGet.Port.IntVal = 8081
				svc.Spec.Ports[0].Port = 8081
				svc.Annotations["prometheus.io/port"] = "8081"
				updateDeploy(deploy)
				updateService(svc)

				wait(requests, res.expectedRequest, 2)

				updateDeploy := getDeploy(res.depKey)
				updatesvc := getService(res.serviceKey)
				Expect(updatesvc.Spec.Ports[0].Port).Should(Equal(javaboot.Spec.Port))
				Expect(updatesvc.Annotations["prometheus.io/port"]).Should(Equal("8080"))
				Expect(updateDeploy.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort).Should(Equal(javaboot.Spec.Port))
				Expect(updateDeploy.Spec.Template.Spec.Containers[0].ReadinessProbe.HTTPGet.Port.IntVal).Should(Equal(javaboot.Spec.Port))
				Expect(updateDeploy.Spec.Template.Spec.Containers[0].LivenessProbe.HTTPGet.Port.IntVal).Should(Equal(javaboot.Spec.Port))
			})
		})

		Context("can not update resources by deployment", func() {
			It("can not scale up cpu and memory by deployment", func() {
				res := genResource()
				replicas := int32(3)

				resources := &corev1.ResourceRequirements{
					Limits:   map[corev1.ResourceName]resource.Quantity{},
					Requests: map[corev1.ResourceName]resource.Quantity{},
				}

				resources.Limits[corev1.ResourceMemory] = *resource.NewMilliQuantity(2048, resource.BinarySI)
				resources.Limits[corev1.ResourceCPU] = *resource.NewQuantity(2, resource.DecimalSI)

				resources.Requests[corev1.ResourceMemory] = *resource.NewMilliQuantity(1024, resource.BinarySI)
				resources.Requests[corev1.ResourceCPU] = *resource.NewQuantity(1, resource.DecimalSI)

				javaboot := &javabootv1.JavaBoot{
					ObjectMeta: metav1.ObjectMeta{Name: res.bootKey.Name, Namespace: res.bootKey.Namespace},
					Spec: javabootv1.BootSpec{
						Replicas:  &replicas,
						Port:      8080,
						Resources: *resources,
					},
				}

				createBoot(javaboot)
				defer c.Delete(context.TODO(), javaboot)
				wait(requests, res.expectedRequest, 3)

				deploy := getDeploy(res.depKey)

				// check resource
				Expect(deploy.Spec.Template.Spec.Containers[0].Resources.Limits.Memory()).Should(Equal(resources.Limits.Memory()))
				Expect(deploy.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu()).Should(Equal(resources.Limits.Cpu()))

				Expect(deploy.Spec.Template.Spec.Containers[0].Resources.Requests.Memory()).Should(Equal(resources.Requests.Memory()))
				Expect(deploy.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu()).Should(Equal(resources.Requests.Cpu()))

				deploy.Spec.Template.Spec.Containers[0].Resources.Limits[corev1.ResourceMemory] = *resource.NewMilliQuantity(2048*2, resource.BinarySI)
				deploy.Spec.Template.Spec.Containers[0].Resources.Limits[corev1.ResourceCPU] = *resource.NewQuantity(2*2, resource.DecimalSI)
				deploy.Spec.Template.Spec.Containers[0].Resources.Requests[corev1.ResourceMemory] = *resource.NewMilliQuantity(1024*2, resource.BinarySI)
				deploy.Spec.Template.Spec.Containers[0].Resources.Requests[corev1.ResourceCPU] = *resource.NewQuantity(1*2, resource.DecimalSI)

				updateDeploy(deploy)
				wait(requests, res.expectedRequest, 1)

				//check resource
				updateDeploy := getDeploy(res.depKey)
				Expect(updateDeploy.Spec.Template.Spec.Containers[0].Resources.Limits.Memory()).Should(Equal(resources.Limits.Memory()))
				Expect(updateDeploy.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu()).Should(Equal(resources.Limits.Cpu()))

				Expect(updateDeploy.Spec.Template.Spec.Containers[0].Resources.Requests.Memory()).Should(Equal(resources.Requests.Memory()))
				Expect(updateDeploy.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu()).Should(Equal(resources.Requests.Cpu()))

			})

			It("can not scale down cpu and memory by deployment", func() {
				res := genResource()
				replicas := int32(3)

				resources := &corev1.ResourceRequirements{
					Limits:   map[corev1.ResourceName]resource.Quantity{},
					Requests: map[corev1.ResourceName]resource.Quantity{},
				}

				resources.Limits[corev1.ResourceMemory] = *resource.NewMilliQuantity(2048*2, resource.BinarySI)
				resources.Limits[corev1.ResourceCPU] = *resource.NewQuantity(2*2, resource.DecimalSI)

				resources.Requests[corev1.ResourceMemory] = *resource.NewMilliQuantity(1024*2, resource.BinarySI)
				resources.Requests[corev1.ResourceCPU] = *resource.NewQuantity(1*2, resource.DecimalSI)

				javaboot := &javabootv1.JavaBoot{
					ObjectMeta: metav1.ObjectMeta{Name: res.bootKey.Name, Namespace: res.bootKey.Namespace},
					Spec: javabootv1.BootSpec{
						Replicas:  &replicas,
						Port:      8080,
						Resources: *resources,
					},
				}

				createBoot(javaboot)
				defer c.Delete(context.TODO(), javaboot)
				wait(requests, res.expectedRequest, 3)

				deploy := getDeploy(res.depKey)

				// check resource
				Expect(deploy.Spec.Template.Spec.Containers[0].Resources.Limits.Memory()).Should(Equal(resources.Limits.Memory()))
				Expect(deploy.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu()).Should(Equal(resources.Limits.Cpu()))

				Expect(deploy.Spec.Template.Spec.Containers[0].Resources.Requests.Memory()).Should(Equal(resources.Requests.Memory()))
				Expect(deploy.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu()).Should(Equal(resources.Requests.Cpu()))

				//update resource

				deploy.Spec.Template.Spec.Containers[0].Resources.Limits[corev1.ResourceMemory] = *resource.NewMilliQuantity(2048, resource.BinarySI)
				deploy.Spec.Template.Spec.Containers[0].Resources.Limits[corev1.ResourceCPU] = *resource.NewQuantity(2, resource.DecimalSI)

				deploy.Spec.Template.Spec.Containers[0].Resources.Requests[corev1.ResourceMemory] = *resource.NewMilliQuantity(1024, resource.BinarySI)
				deploy.Spec.Template.Spec.Containers[0].Resources.Requests[corev1.ResourceCPU] = *resource.NewQuantity(1, resource.DecimalSI)

				updateDeploy(deploy)
				wait(requests, res.expectedRequest, 1)

				//check resource
				updateDeploy := getDeploy(res.depKey)
				Expect(updateDeploy.Spec.Template.Spec.Containers[0].Resources.Limits.Memory()).Should(Equal(resources.Limits.Memory()))
				Expect(updateDeploy.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu()).Should(Equal(resources.Limits.Cpu()))

				Expect(updateDeploy.Spec.Template.Spec.Containers[0].Resources.Requests.Memory()).Should(Equal(resources.Requests.Memory()))
				Expect(updateDeploy.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu()).Should(Equal(resources.Requests.Cpu()))

			})
		})

		Context("can not update nodeSelector by deployment", func() {
			It("can not update nodeSelector", func() {
				res := genResource()
				replicas := int32(3)
				javaboot := &javabootv1.JavaBoot{
					ObjectMeta: metav1.ObjectMeta{Name: res.bootKey.Name, Namespace: res.bootKey.Namespace},
					Spec: javabootv1.BootSpec{
						Replicas:     &replicas,
						Port:         8080,
						NodeSelector: map[string]string{"env": "test", "app": "myAPPName"},
					},
				}

				createBoot(javaboot)
				defer c.Delete(context.TODO(), javaboot)
				wait(requests, res.expectedRequest, 3)

				deploy := getDeploy(res.depKey)

				// check NodeSelector
				Expect(deploy.Spec.Template.Spec.NodeSelector).Should(Equal(javaboot.Spec.NodeSelector))

				//update NodeSelector
				deploy.Spec.Template.Spec.NodeSelector = map[string]string{"env": "test", "app": "myAPPName2", "new": "new_label"}
				updateDeploy(deploy)

				wait(requests, res.expectedRequest, 1)

				//check NodeSelector
				updateDeploy := getDeploy(res.depKey)
				Expect(updateDeploy.Spec.Template.Spec.NodeSelector).Should(Equal(javaboot.Spec.NodeSelector))

			})
		})

		Context("can not update  health by deployment", func() {
			It("can not update health", func() {
				res := genResource()
				replicas := int32(3)
				health := "/health"
				javaboot := &javabootv1.JavaBoot{
					ObjectMeta: metav1.ObjectMeta{Name: res.bootKey.Name, Namespace: res.bootKey.Namespace},
					Spec: javabootv1.BootSpec{
						Replicas: &replicas,
						Port:     8080,
						Health:   &health,
					},
				}

				createBoot(javaboot)
				defer c.Delete(context.TODO(), javaboot)
				wait(requests, res.expectedRequest, 3)

				deploy := getDeploy(res.depKey)

				// check health
				Expect(deploy.Spec.Template.Spec.Containers[0].LivenessProbe.HTTPGet.Path).Should(Equal(health))
				Expect(deploy.Spec.Template.Spec.Containers[0].ReadinessProbe.HTTPGet.Path).Should(Equal(health))

				// get boot

				//update health
				deploy.Spec.Template.Spec.Containers[0].LivenessProbe.HTTPGet.Path = "/health2"
				deploy.Spec.Template.Spec.Containers[0].ReadinessProbe.HTTPGet.Path = "/health2"

				updateDeploy(deploy)

				wait(requests, res.expectedRequest, 1)

				//check health
				updateDeploy := getDeploy(res.depKey)

				Expect(updateDeploy.Spec.Template.Spec.Containers[0].LivenessProbe.HTTPGet.Path).Should(Equal(health))
				Expect(updateDeploy.Spec.Template.Spec.Containers[0].ReadinessProbe.HTTPGet.Path).Should(Equal(health))

			})
		})

		Context("can not update env by deployment", func() {
			It("can not update env simple", func() {
				res := genResource()
				//logan.OperatorUID="local"
				replicas := int32(3)
				javaboot := &javabootv1.JavaBoot{
					ObjectMeta: metav1.ObjectMeta{Name: res.bootKey.Name, Namespace: res.bootKey.Namespace},
					Spec: javabootv1.BootSpec{
						Replicas: &replicas,
						Port:     8080,
						Env: []corev1.EnvVar{
							{Name: "key1", Value: "value1"},
							{Name: "key2", Value: "value2"},
							{Name: "myApp", Value: "${APP}"},
							{Name: "myEnv", Value: "${ENV}"},
						},
					},
				}

				createBoot(javaboot)
				defer c.Delete(context.TODO(), javaboot)
				wait(requests, res.expectedRequest, 3)
				deploy := getDeploy(res.depKey)

				// check env
				for _, i := range javaboot.Spec.Env {
					if strings.EqualFold(i.Name, "myAPP") {
						i.Value = res.bootKey.Name
					}
					if strings.EqualFold(i.Name, "myEnv") {
						i.Value = "test"
					}

					for _, j := range deploy.Spec.Template.Spec.Containers[0].Env {
						if strings.EqualFold(i.Name, j.Name) {
							Expect(i).Should(Equal(j))
						}
					}
				}

				deploy.Spec.Template.Spec.Containers[0].Env = []corev1.EnvVar{
					{Name: "key1", Value: "value2"},
					{Name: "key5", Value: "value1"},
					{Name: "myApp", Value: "${APP}"},
					{Name: "myEnv", Value: "${ENV}"},
				}

				updateDeploy(deploy)

				wait(requests, res.expectedRequest, 1)
				//check env
				updateDeploy := getDeploy(res.depKey)

				for _, i := range javaboot.Spec.Env {
					if strings.EqualFold(i.Name, "myAPP") {
						i.Value = res.bootKey.Name
					}
					if strings.EqualFold(i.Name, "myEnv") {
						i.Value = "test"
					}
					for _, j := range updateDeploy.Spec.Template.Spec.Containers[0].Env {
						if strings.EqualFold(i.Name, j.Name) {
							Expect(i).Should(Equal(j))
						}
					}
				}

			})

		})

	})
})

var counter int = 0

type TestResource struct {
	expectedRequest reconcile.Request
	depKey          types.NamespacedName
	bootKey         types.NamespacedName
	serviceKey      types.NamespacedName
}

func genResource() *TestResource {
	res := &TestResource{}
	res.bootKey = types.NamespacedName{
		Name:      "foo" + strconv.Itoa(counter),
		Namespace: "default" + strconv.Itoa(counter)}

	res.depKey = types.NamespacedName{
		Name:      res.bootKey.Name,
		Namespace: res.bootKey.Namespace}

	res.serviceKey = res.depKey
	res.expectedRequest = reconcile.Request{NamespacedName: res.bootKey}
	counter++
	return res
}

func wait(requests chan reconcile.Request, req reconcile.Request, cnt int) {
	for i := 0; i < cnt; i++ {
		log.Info("wait", "time", i)
		Eventually(requests, timeout).Should(Receive(Equal(req)))
	}
}

func createBoot(javaboot *javabootv1.JavaBoot) {
	// Create the JavaBoot object and expect the Reconcile and Deployment to be created
	err := c.Create(context.TODO(), javaboot)
	// The javaboot object may not be a valid object because it might be missing some required fields.
	// Please modify the javaboot object by adding required fields and then remove the following if statement.
	if apierrors.IsInvalid(err) {
		log.Error(err, "failed to create object, got an invalid object error: %v")
		return
	}
	Expect(err).NotTo(HaveOccurred())
}

func getBoot(bootKey types.NamespacedName) *javabootv1.JavaBoot {
	time.Sleep(2 * time.Second)
	boot := &javabootv1.JavaBoot{}
	Eventually(func() error {
		return c.Get(context.TODO(), bootKey, boot)
	}, timeout).
		Should(Succeed())
	log.Info("get boot ok!")
	return boot
}

func getDeploy(depKey types.NamespacedName) *appsv1.Deployment {
	time.Sleep(2 * time.Second)
	deploy := &appsv1.Deployment{}

	Eventually(func() error {
		return c.Get(context.TODO(), depKey, deploy)
	}, timeout).
		Should(Succeed())
	log.Info("get deploy ok!")
	return deploy
}

func getService(serviceKey types.NamespacedName) *corev1.Service {
	time.Sleep(2 * time.Second)
	appSvcFound := &corev1.Service{}
	Eventually(func() error {
		return c.Get(context.TODO(), serviceKey, appSvcFound)
	}, timeout).
		Should(Succeed())
	log.Info("get service ok!")
	return appSvcFound
}

func updateBoot(boot *javabootv1.JavaBoot) {
	err := c.Update(context.TODO(), boot)
	if apierrors.IsInvalid(err) {
		log.Error(err, "failed to update object, got an invalid object error: %v")
		return
	}
	Expect(err).NotTo(HaveOccurred())
	log.Info("update boot ok!")
}

func updateDeploy(deploy *appsv1.Deployment) {
	err := c.Update(context.TODO(), deploy)
	if apierrors.IsInvalid(err) {
		log.Error(err, "failed to update object, got an invalid object error: %v")
		return
	}
	Expect(err).NotTo(HaveOccurred())
	log.Info("update deploy ok!")
}

func updateService(svr *corev1.Service) {
	err := c.Update(context.TODO(), svr)
	if apierrors.IsInvalid(err) {
		log.Error(err, "failed to update object, got an invalid object error: %v")
		return
	}
	Expect(err).NotTo(HaveOccurred())
	log.Info("update service ok!")
}

func getConfig() string {
	configText := `
java:
  settings:
    registry: "registry.logan.local"
  oEnvs:
    app:
      test:
        env:
          - name: LOGAN_ZIPKIN_KAFKA_BOOTSTRAP-SERVERS
            value: "127.0.0.1:9092"
        subDomain: "test.logan.local"
  app:
    port: 8080
    replicas: 1
    health: /health
    env:
      - name: SPRING_ZIPKIN_ENABLED
        value: "true"
      - name: SPRING_ZIPKIN_KAFKA_TOPIC
        value: "logan-tracing"
      - name: SERVER_PORT
        value: "8080"
    resources:
      limits:
        cpu: "2"
        memory: "2Gi"
      requests:
        cpu: "10m"
        memory: "1Gi"
    subDomain: "exp.logan.local"
`
	return configText
}
