package testutils

import (
	"k8s.io/client-go/rest"
)

func DeployTestRunner(cfg *rest.Config, namespace string) error {
	return DeployFromYaml(cfg, namespace, TestRunnerYaml)
}

const TestRunnerYaml = `
apiVersion: v1
kind: Pod
metadata:
  labels:
    gloo: testrunner
  name: testrunner
spec:
  containers:
  - image: soloio/testrunner:testing-8671e8b9
    imagePullPolicy: IfNotPresent
    command:
      - sleep
      - "36000"
    name: testrunner
  restartPolicy: Always`
