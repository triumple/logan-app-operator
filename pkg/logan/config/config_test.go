package config

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	coreV1 "k8s.io/api/core/v1"
)

var _ = Describe("Config", func() {

	Context("With empty config content", func() {
		var err error

		It("Test empty config is valid", func() {
			err = NewConfigFromString("")
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("With default config file", func() {
		It("Test parsing config", func() {
			text := `
java:
  oEnvs:
    app:
      dev:
        env:
          - name: TEST_ENV1
            value: "-Denv=${ENV}"
  app:
    port: 8080
    replicas: 1
    health: /health
    env:
      - name: SPRING_PROFILES_ACTIVE
        value: "${ENV}"
      - name: SERVER_PORT
        value: "8080"
`
			err := NewConfigFromString(text)
			Expect(err).NotTo(HaveOccurred())

			Expect(JavaConfig.AppSpec.Port).To(BeEquivalentTo(8080))
			Expect(JavaConfig.AppSpec.Replicas).To(BeEquivalentTo(1))
			Expect(JavaConfig.AppSpec.Health).To(Equal("/health"))

			Expect(JavaConfig.AppSpec).NotTo(BeNil())
		})
	})

	Context("Test app config", func() {

		It("Test app config with default config", func() {
			text := `
java:
  settings:
    registry: "registry.logan.local"
  oEnvs:
    app:
`
			err := NewConfigFromString(text)
			Expect(err).NotTo(HaveOccurred())

			Expect(JavaConfig.AppSpec.Port).To(BeEquivalentTo(8080))
			Expect(JavaConfig.AppSpec.Replicas).To(BeEquivalentTo(1))
			Expect(JavaConfig.AppSpec.Health).To(Equal("/health"))
			Expect(JavaConfig.SidecarContainers).Should(BeNil())
		})

		It("Test app config with oenv config", func() {
			text := `
java:
  settings:
    registry: "registry.logan.local"
  oEnvs:
    app:
      test:
        port: 8082
        replicas: 2
        health: /health2
        env:
            # Podpreset
            - name: SPRING_ZIPKIN_ENABLED2
              value: "true"
        nodeSelector:
          logan/env: test
        resources:
          limits:
            cpu: "2"
            memory: "2Gi"
          requests:
            cpu: "1"
            memory: "1Gi"
        subDomain: "2exp.logan.local"
`
			err := NewConfigFromString(text)
			Expect(err).NotTo(HaveOccurred())

			Expect(JavaConfig.AppSpec.Port).To(BeEquivalentTo(8082))
			Expect(JavaConfig.AppSpec.Replicas).To(BeEquivalentTo(2))
			Expect(JavaConfig.AppSpec.Health).To(Equal("/health2"))
			Expect(JavaConfig.AppSpec.SubDomain).To(Equal("2exp.logan.local"))

			Expect(JavaConfig.AppSpec.Env[0].Name).To(Equal("SPRING_ZIPKIN_ENABLED2"))
			Expect(JavaConfig.AppSpec.Env[0].Value).To(Equal("true"))

			myNodeSelector := map[string]string{"logan/env": "test"}
			Expect(JavaConfig.AppSpec.NodeSelector).Should(Equal(myNodeSelector))

			Expect(JavaConfig.AppSpec.Resources.Limits.Cpu().Value()).To(Equal(int64(2)))
			Expect(JavaConfig.AppSpec.Resources.Limits.Memory().Value()).To(Equal(int64(2048 * 1024 * 1024)))
			Expect(JavaConfig.AppSpec.Resources.Requests.Cpu().Value()).To(Equal(int64(1)))
			Expect(JavaConfig.AppSpec.Resources.Requests.Memory().Value()).To(Equal(int64(1024 * 1024 * 1024)))
		})

		It("Test app config with app config", func() {
			text := `
java:
  settings:
    registry: "registry.logan.local"
  oEnvs:
    app:
      test:
  app:
    port: 8083
    replicas: 3
    health: /health3
    env:
      # Podpreset
      - name: SPRING_ZIPKIN_ENABLED
        value: "true"
    nodeSelector:
      logan/env: test
    resources:
      limits:
        cpu: "2"
        memory: "2Gi"
      requests:
        cpu: "1"
        memory: "1Gi"
    subDomain: "3exp.logan.local"
`
			err := NewConfigFromString(text)
			Expect(err).NotTo(HaveOccurred())
			Expect(JavaConfig.AppSpec.Port).To(BeEquivalentTo(8083))
			Expect(JavaConfig.AppSpec.Replicas).To(BeEquivalentTo(3))
			Expect(JavaConfig.AppSpec.Health).To(Equal("/health3"))
			Expect(JavaConfig.AppSpec.SubDomain).To(Equal("3exp.logan.local"))

			Expect(JavaConfig.AppSpec.Env[0].Name).To(Equal("SPRING_ZIPKIN_ENABLED"))
			Expect(JavaConfig.AppSpec.Env[0].Value).To(Equal("true"))

			myNodeSelector := map[string]string{"logan/env": "test"}
			Expect(JavaConfig.AppSpec.NodeSelector).Should(Equal(myNodeSelector))

			Expect(JavaConfig.AppSpec.Resources.Limits.Cpu().Value()).To(Equal(int64(2)))
			Expect(JavaConfig.AppSpec.Resources.Limits.Memory().Value()).To(Equal(int64(2048 * 1024 * 1024)))
			Expect(JavaConfig.AppSpec.Resources.Requests.Cpu().Value()).To(Equal(int64(1)))
			Expect(JavaConfig.AppSpec.Resources.Requests.Memory().Value()).To(Equal(int64(1024 * 1024 * 1024)))
		})

		It("Test app config order", func() {
			text := `
java:
  settings:
    registry: "registry.logan.local"
  oEnvs:
    app:
      test:
        port: 8082
        replicas: 2
        health: /health2
        env:
          # Podpreset
          - name: SPRING_ZIPKIN_ENABLED2
            value: "false"
        nodeSelector:
          logan/envA: A
          logan/envB: B
        resources:
          limits:
            cpu: "4"
            memory: "4Gi"
          requests:
            cpu: "3"
            memory: "3Gi"
        subDomain: "2exp.logan.local"
  app:
    port: 8083
    replicas: 3
    health: /health3
    env:
      # Podpreset
      - name: SPRING_ZIPKIN_ENABLED
        value: "true"
    nodeSelector:
      logan/envA: NewA
      logan/envC: C
    resources:
      limits:
        cpu: "2"
        memory: "2Gi"
      requests:
        cpu: "1"
        memory: "1Gi"
    subDomain: "3exp.logan.local"
`
			err := NewConfigFromString(text)
			Expect(err).NotTo(HaveOccurred())

			Expect(JavaConfig.AppSpec.Port).To(BeEquivalentTo(8082))
			Expect(JavaConfig.AppSpec.Replicas).To(BeEquivalentTo(2))
			Expect(JavaConfig.AppSpec.Health).To(Equal("/health2"))
			Expect(JavaConfig.AppSpec.SubDomain).To(Equal("2exp.logan.local"))

			myNodeSelector := map[string]string{"logan/envA": "A", "logan/envB": "B", "logan/envC": "C"}
			Expect(JavaConfig.AppSpec.NodeSelector).Should(Equal(myNodeSelector))

			Expect(JavaConfig.AppSpec.Resources.Limits.Cpu().Value()).To(Equal(int64(4)))
			Expect(JavaConfig.AppSpec.Resources.Limits.Memory().Value()).To(Equal(int64(4 * 1024 * 1024 * 1024)))
			Expect(JavaConfig.AppSpec.Resources.Requests.Cpu().Value()).To(Equal(int64(3)))
			Expect(JavaConfig.AppSpec.Resources.Requests.Memory().Value()).To(Equal(int64(3 * 1024 * 1024 * 1024)))
		})

		It("Test app config env order", func() {
			text := `
java:
  settings:
    registry: "registry.logan.local"
  oEnvs:
    app:
      test:
        env:
          # Podpreset
          - name: SPRING_ZIPKIN_ENABLED
            value: "false"
          - name: MY_OENV_APP
            value: "B"
  app:
    env:
      # Podpreset
      - name: SPRING_ZIPKIN_ENABLED
        value: "true"
      - name: MY_ENV_APP
        value: "A"
`
			err := NewConfigFromString(text)
			Expect(err).NotTo(HaveOccurred())

			Expect(JavaConfig.AppSpec.Env[0].Name).Should(Equal("SPRING_ZIPKIN_ENABLED"))
			Expect(JavaConfig.AppSpec.Env[0].Value).Should(Equal("false"))
			Expect(JavaConfig.AppSpec.Env[1].Name).Should(Equal("MY_ENV_APP"))
			Expect(JavaConfig.AppSpec.Env[1].Value).Should(Equal("A"))
			Expect(JavaConfig.AppSpec.Env[2].Name).Should(Equal("MY_OENV_APP"))
			Expect(JavaConfig.AppSpec.Env[2].Value).Should(Equal("B"))

		})

		It("Test PHP app config sidecar env order", func() {
			text := `
php:
  settings:
    registry: "registry.logan.local"
    appHealthPort: 5678
  oEnvs:
    app:
      dev:
        subDomain: "dev.logan.local"
        nodeSelector:
          logan/env: dev
      test:
        nodeSelector:
          logan/env: test
        subDomain: "test.logan.local"
      prod:
        settings:
          registry: "harbor.logan.inner"
        subDomain: "logan.com"
    sidecar:
      test:
        env:
          - name: A
            value: "A"
          - name: B
            value: "B"
  sideCarContainers:
    - name: sidecar
      env:
        - name: A
          value: "newA"
        - name: C
          value: "C"
`
			err := NewConfigFromString(text)
			Expect(err).NotTo(HaveOccurred())

			for _, c := range *PhpConfig.SidecarContainers {
				Expect(c.Env[0].Name).Should(Equal("A"))
				Expect(c.Env[0].Value).Should(Equal("A"))
				Expect(c.Env[1].Name).Should(Equal("C"))
				Expect(c.Env[1].Value).Should(Equal("C"))
				Expect(c.Env[2].Name).Should(Equal("B"))
				Expect(c.Env[2].Value).Should(Equal("B"))
			}
		})

		It("Test PHP app config sidecar", func() {
			text := `
php:
  oEnvs:
    sidecar:
      test:
  sideCarContainers:
    - name: sidecar
      image: '${REGISTRY}/logancloud/logan-pulse-sidecar:0.1.2'
      imagePullPolicy: Always
      env:
        - name: SPRING_PROFILES_ACTIVE
          value: "${ENV}"
      lifecycle:
        preStop:
          exec:
            command: ["/bin/sh", "-c", "/opt/shell/sidecar_pre_stop.sh"]
      ports:
        - name: http
          containerPort: 5678
          protocol: TCP
      resources:
        limits:
          cpu: '2'
          memory: 2Gi
        requests:
          cpu: '1'
          memory: 512Mi
      volumeMounts:
        - mountPath: /opt/data
          name: shared-data
`
			err := NewConfigFromString(text)
			Expect(err).NotTo(HaveOccurred())

			for _, c := range *PhpConfig.SidecarContainers {
				Expect(c.Name).Should(Equal("sidecar"))
				Expect(c.Image).Should(Equal("${REGISTRY}/logancloud/logan-pulse-sidecar:0.1.2"))
				Expect(c.ImagePullPolicy).Should(Equal(coreV1.PullAlways))
				Expect(c.Lifecycle.PreStop.Exec.Command).Should(Equal([]string{"/bin/sh", "-c", "/opt/shell/sidecar_pre_stop.sh"}))

				Expect(c.Resources.Limits.Cpu().Value()).To(Equal(int64(2)))
				Expect(c.Resources.Limits.Memory().Value()).To(Equal(int64(2 * 1024 * 1024 * 1024)))
				Expect(c.Resources.Requests.Cpu().Value()).To(Equal(int64(1)))
				Expect(c.Resources.Requests.Memory().Value()).To(Equal(int64(0.5 * 1024 * 1024 * 1024)))
			}
		})

		It("Test PHP app config sidecarServices ", func() {
			text := `
php:
  settings:
    registry: "registry.logan.local"

  sidecarServices:
    - name: ${APP}-sidecar
      port: 5678
`
			err := NewConfigFromString(text)
			Expect(err).NotTo(HaveOccurred())

			for _, s := range *PhpConfig.SidecarServices {
				Expect(s.Name).Should(Equal("${APP}-sidecar"))
				Expect(s.Port).Should(Equal(int32(5678)))
			}
		})

	})

})
