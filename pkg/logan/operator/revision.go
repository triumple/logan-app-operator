package operator

import (
	"encoding/json"
	"fmt"
	v1 "github.com/logancloud/logan-app-operator/pkg/apis/app/v1"
	"github.com/logancloud/logan-app-operator/pkg/logan"
	"github.com/logancloud/logan-app-operator/pkg/logan/config"
	"github.com/sergi/go-diff/diffmatchpatch"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

const (
	// RevisionPhaseRunning is the revision phase for Running
	RevisionPhaseRunning = "Running"
	// RevisionPhaseActive is the revision phase for Active
	RevisionPhaseActive = "Active"
	// RevisionPhaseComplete is the revision phase for Complete
	RevisionPhaseComplete = "Complete"
	// RevisionPhaseCancel is the revision phase for Cancelled
	RevisionPhaseCancel = "Cancelled"
)

// InitBootRevision will init a revision from boot
func InitBootRevision(boot *v1.Boot) *v1.BootRevision {
	revisionBoot := &v1.BootRevision{
		ObjectMeta: metav1.ObjectMeta{
			Name:        boot.Name,
			Namespace:   boot.Namespace,
			Annotations: initRevisionAnnotations(boot),
		},
		Spec:     boot.Spec,
		BootType: boot.BootType,
		AppKey:   boot.AppKey,
	}
	replica := int32(0)
	revisionBoot.Spec.Replicas = &replica

	revisionBoot.Spec.Env = cleanEnv(revisionBoot.Spec.Env)

	return revisionBoot
}

func initRevisionAnnotations(boot *v1.Boot) map[string]string {
	if boot.Annotations != nil {
		if val, isok := boot.Annotations[config.BootProfileAnnotationKey]; isok {
			return map[string]string{
				config.BootProfileAnnotationKey: val,
			}
		}
	}
	return map[string]string{}
}

func cleanEnv(envs []corev1.EnvVar) []corev1.EnvVar {
	ret := make([]corev1.EnvVar, 0)
	if envs != nil {
		for _, env := range envs {
			if _, found := logan.BizEnvs[env.Name]; !found {
				ret = append(ret, env)
			}
		}
	}
	return ret
}

// RevisionDiff will compute the diff with two revision
func RevisionDiff(current, latest v1.BootRevision) string {
	currentCopy := current.DeepCopy()
	currentCopy.ObjectMeta = metav1.ObjectMeta{}
	currentYaml, _ := yaml.Marshal(currentCopy)

	latestCopy := latest.DeepCopy()
	latestCopy.ObjectMeta = metav1.ObjectMeta{}
	latestYaml, _ := yaml.Marshal(latestCopy)

	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(string(latestYaml), string(currentYaml), false)
	log, err := marshalChangelog(diffs)
	if err != nil {
		return ""
	}
	return log
}

// marshalChangelog marshal the []diffmatchpatch.Diff to string
func marshalChangelog(changelog []diffmatchpatch.Diff) (string, error) {
	changelogSe, err := json.Marshal(changelog)
	changelogStr := fmt.Sprintf("%s", changelogSe)
	return changelogStr, err
}

func updateRevisionAnnotation(revision *v1.BootRevision, revisionAnnotationMap map[string]string) bool {
	updated := false
	if revision.Annotations == nil {
		revision.Annotations = make(map[string]string)
	}
	for aKey, aValue := range revisionAnnotationMap {
		if metaDataVal, exist := revision.Annotations[aKey]; exist {
			// Annotation Map contains the key
			if metaDataVal != aValue {
				revision.Annotations[aKey] = aValue
				updated = true
			}
		} else {
			// Annotation Map does not contain the key
			revision.Annotations[aKey] = aValue
			updated = true
		}
	}
	return updated
}
