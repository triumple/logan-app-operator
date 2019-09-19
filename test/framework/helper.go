package framework

import (
	"github.com/google/uuid"
	"github.com/onsi/ginkgo/config"
	"k8s.io/apimachinery/pkg/types"
	"os"
	"strconv"
	"time"
)

const (
	defaultWaitSec = 1

	WAITTIME_KEY = "WAIT_TIME"
	FOCUS_KEY    = "GINKGO_FOCUS"
)

func GenResource() types.NamespacedName {
	id := uuid.New().String()
	bootKey := types.NamespacedName{
		Name:      "foo-" + id,
		Namespace: "bar-" + id}
	return bootKey
}

func WaitUpdate(sec int) {
	time.Sleep(time.Duration(sec) * time.Second)
}

func WaitDefaultUpdate() {
	WaitUpdate(waitTime)
}

var waitTime int

func init() {
	focus, found := os.LookupEnv(FOCUS_KEY)
	if found {
		config.GinkgoConfig.FocusString = focus
	}

	t, found := os.LookupEnv(WAITTIME_KEY)
	if found {
		waitTime, _ = strconv.Atoi(t)
	} else {
		waitTime = defaultWaitSec
	}
}
