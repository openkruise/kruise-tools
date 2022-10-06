/*
Copyright 2022 The Kruise Authors.
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/openkruise/kruise-tools/cmd/resourcedistributiongenerator/generator"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestResourceDistributionGenerator(t *testing.T) {
	tests := []struct {
		name         string
		in           string
		expect       string
		filename     []string
		envFilename  []string
		setupFile    func(t *testing.T, FileSources []string) func()
		setupEnvFile func(t *testing.T, EnvFileSources []string) func()
	}{
		{
			name:         "configmap-literals-envs-files-resourceOptions-options-allTargets",
			setupFile:    setupBinaryFile([]byte{0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64}),
			filename:     []string{"foo1", "foo2"},
			setupEnvFile: setupEnvFile([][]string{{"key1=value1", "#", "", "key2=value2"}, {"key3=value3"}}),
			envFilename:  []string{"file1.env", "file2.env"},
			in: `
apiVersion: config.kubernetes.io/v1
kind: ResourceList
functionConfig:
  apiVersion: apps.kruise.io/v1alpha1
  kind: ResourceDistributionGenerator
  metadata:
    name: rdname
  resource:
    resourceKind: ConfigMap
    resourceName: cmname
    literals:
      - JAVA_HOME=/opt/java/jdk
      - A=b
    envs:
      - file1.env
      - file2.env
    files:
      - foo1
      - foo2
    resourceOptions:
      annotations:
        dashboard: "1"
      immutable: true
      labels:
        test: true
        rsla: rs
  options:
    labels:
      app.kubernetes.io/name: "app1"
    annotations:
      an: rdan
  targets:
    allNamespaces: true
    includedNamespaces:
      - ns-1
    excludedNamespaces:
      - ns-2
    namespaceLabelSelector:
      matchLabels:
        group: "test"
      matchExpressions:
        - key: ffxc
          operator: In
          values:
            - l
            - a
        - key: exc
          operator: NotIn
          values:
            - albc
            - a
        - key: abc
          operator: Exists
`,
			expect: `
apiVersion: apps.kruise.io/v1alpha1
kind: ResourceDistribution
spec:
  resource:
    apiVersion: v1
    data:
      A: b
      JAVA_HOME: /opt/java/jdk
      foo1: hello world
      foo2: hello world
      key1: value1
      key2: value2
      key3: value3
    immutable: true
    kind: ConfigMap
    metadata:
      annotations:
        dashboard: "1"
      labels:
        rsla: rs
        test: "true"
      name: cmname
  targets:
    allNamespaces: true
    excludedNamespaces:
      list:
      - name: ns-2
    includedNamespaces:
      list:
      - name: ns-1
    namespaceLabelSelector:
      matchExpressions:
      - key: abc
        operator: Exists
      - key: exc
        operator: NotIn
        values:
        - a
        - albc
      - key: ffxc
        operator: In
        values:
        - a
        - l
      matchLabels:
        group: test
metadata:
  annotations:
    an: rdan
  labels:
    app.kubernetes.io/name: app1
  name: rdname
`,
		},
		{
			name: "configmap-literals-includedNamespaces",
			in: `
apiVersion: config.kubernetes.io/v1
kind: ResourceList
functionConfig:
  apiVersion: apps.kruise.io/v1alpha1
  kind: ResourceDistributionGenerator
  metadata:
    name: rdname
  resource:
    resourceKind: ConfigMap
    resourceName: cmname
    literals:
      - JAVA_HOME= /opt/java/jdk
      - A=b
  targets:
    includedNamespaces:
      - ns-1
`,
			expect: `
apiVersion: apps.kruise.io/v1alpha1
kind: ResourceDistribution
spec:
  resource:
    apiVersion: v1
    data:
      A: b
      JAVA_HOME: ' /opt/java/jdk'
    kind: ConfigMap
    metadata:
      name: cmname
  targets:
    includedNamespaces:
      list:
      - name: ns-1
metadata:
  name: rdname
`,
		},
		{
			name: "configmap-literals-includedNamespaces",
			in: `
apiVersion: config.kubernetes.io/v1
kind: ResourceList
functionConfig:
  apiVersion: apps.kruise.io/v1alpha1
  kind: ResourceDistributionGenerator
  metadata:
    name: rdname
  resource:
    resourceKind: ConfigMap
    resourceName: cmname
    literals:
      - JAVA_HOME=/opt/java/jdk
      - A=b
  targets:
    includedNamespaces:
      - ns-1
`,
			expect: `
apiVersion: apps.kruise.io/v1alpha1
kind: ResourceDistribution
spec:
  resource:
    apiVersion: v1
    data:
      A: b
      JAVA_HOME: /opt/java/jdk
    kind: ConfigMap
    metadata:
      name: cmname
  targets:
    includedNamespaces:
      list:
      - name: ns-1
metadata:
  name: rdname
`,
		},
		{
			name:         "configmap-envs-resourceOptions-allNamespaces",
			setupEnvFile: setupEnvFile([][]string{{"key1=value1"}}),
			envFilename:  []string{"file1.env"},
			in: `
apiVersion: config.kubernetes.io/v1
kind: ResourceList
functionConfig:
  apiVersion: apps.kruise.io/v1alpha1
  kind: ResourceDistributionGenerator
  metadata:
    name: rdname
  resource:
    resourceKind: ConfigMap
    resourceName: cmname
    envs:
      - file1.env
    resourceOptions:
      annotations:
        dashboard: "1"
      immutable: false
  targets:
    allNamespaces: true
`,
			expect: `
apiVersion: apps.kruise.io/v1alpha1
kind: ResourceDistribution
spec:
  resource:
    apiVersion: v1
    data:
      key1: value1
    kind: ConfigMap
    metadata:
      annotations:
        dashboard: "1"
      name: cmname
  targets:
    allNamespaces: true
metadata:
  name: rdname
`,
		},
		{
			name:      "configmap-files-resourceOptions-matchLabels",
			setupFile: setupBinaryFile([]byte{0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64}),
			filename:  []string{"foo1", "foo2"},
			in: `
apiVersion: config.kubernetes.io/v1
kind: ResourceList
functionConfig:
  apiVersion: apps.kruise.io/v1alpha1
  kind: ResourceDistributionGenerator
  metadata:
    name: rdname
  resource:
    resourceKind: ConfigMap
    resourceName: cmname
    files:
      - foo1
    resourceOptions:
      labels:
        rsla: rs
      immutable: true
  targets:
    namespaceLabelSelector:
      matchLabels:
        group: "test"
        app: "dev"
`,
			expect: `
apiVersion: apps.kruise.io/v1alpha1
kind: ResourceDistribution
spec:
  resource:
    apiVersion: v1
    data:
      foo1: hello world
    immutable: true
    kind: ConfigMap
    metadata:
      labels:
        rsla: rs
      name: cmname
  targets:
    namespaceLabelSelector:
      matchLabels:
        app: dev
        group: test
metadata:
  name: rdname
`,
		},
		{
			name:         "configmap-literals-envs-options-matchExpressions",
			setupFile:    setupBinaryFile([]byte{0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64}),
			filename:     []string{"foo1", "foo2"},
			setupEnvFile: setupEnvFile([][]string{{"key1=value1", "#", "", " key2=value2"}, {"key3=value3"}}),
			envFilename:  []string{"file1.env", "file2.env"},
			in: `
apiVersion: config.kubernetes.io/v1
kind: ResourceList
functionConfig:
  apiVersion: apps.kruise.io/v1alpha1
  kind: ResourceDistributionGenerator
  metadata:
    name: rdname
  resource:
    resourceKind: ConfigMap
    resourceName: cmname
    literals:
      - JAVA_HOME=/opt/java/jdk
      - A=b
    files:
      - foo1
  options:
    labels:
      app.kubernetes.io/name: "app1"
  targets:
    namespaceLabelSelector:
      matchExpressions:
        - key: ffxc
          operator: DoesNotExist
        - key: exc
          operator: Exists
`,
			expect: `
apiVersion: apps.kruise.io/v1alpha1
kind: ResourceDistribution
spec:
  resource:
    apiVersion: v1
    data:
      A: b
      JAVA_HOME: /opt/java/jdk
      foo1: hello world
    kind: ConfigMap
    metadata:
      name: cmname
  targets:
    namespaceLabelSelector:
      matchExpressions:
      - key: exc
        operator: Exists
      - key: ffxc
        operator: DoesNotExist
metadata:
  labels:
    app.kubernetes.io/name: app1
  name: rdname
`,
		},
		{
			name:         "configmap-envs-files-options-matchExpressions",
			setupFile:    setupBinaryFile([]byte{0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64}),
			filename:     []string{"foo1"},
			setupEnvFile: setupEnvFile([][]string{{"key1=value1", "#", "", "key2=value2"}}),
			envFilename:  []string{"file1.env"},
			in: `
apiVersion: config.kubernetes.io/v1
kind: ResourceList
functionConfig:
  apiVersion: apps.kruise.io/v1alpha1
  kind: ResourceDistributionGenerator
  metadata:
    name: rdname
  resource:
    resourceKind: ConfigMap
    resourceName: cmname
    envs:
      - file1.env
    files:
      - foo1
  options:
    annotations:
      node: 123 
      an: rdan
  targets:
    namespaceLabelSelector:
      matchExpressions:
        - key: exc
          operator: NotIn
          values:
            - albc
        - key: abc
          operator: Exists
`,
			expect: `
apiVersion: apps.kruise.io/v1alpha1
kind: ResourceDistribution
spec:
  resource:
    apiVersion: v1
    data:
      foo1: hello world
      key1: value1
      key2: value2
    kind: ConfigMap
    metadata:
      name: cmname
  targets:
    namespaceLabelSelector:
      matchExpressions:
      - key: abc
        operator: Exists
      - key: exc
        operator: NotIn
        values:
        - albc
metadata:
  annotations:
    an: rdan
    node: "123"
  name: rdname
`,
		},
		{
			name:         "secret-literals-envs-files-resourceOptions-options-allTargets",
			setupFile:    setupBinaryFile([]byte{0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64}),
			filename:     []string{"foo1", "foo2"},
			setupEnvFile: setupEnvFile([][]string{{"key1=value1", "#", "", "key2=value2"}, {"key3=value3"}}),
			envFilename:  []string{"file1.env", "file2.env"},
			in: `
apiVersion: config.kubernetes.io/v1
kind: ResourceList
functionConfig:
  apiVersion: apps.kruise.io/v1alpha1
  kind: ResourceDistributionGenerator
  metadata:
    name: rdname
  resource:
    resourceKind: Secret
    resourceName: cmname
    literals:
      - JAVA_HOME=/opt/java/jdk
      - A=b
    envs:
      - file1.env
      - file2.env
    files:
      - foo1
      - foo2
    resourceOptions:
      annotations:
        dashboard: "1"
      immutable: true
      labels:
        test: true
        rsla: rs
  options:
    labels:
      app.kubernetes.io/name: "app1"
    annotations:
      an: rdan
  targets:
    allNamespaces: true
    includedNamespaces:
      - ns-1
    namespaceLabelSelector:
      matchLabels:
        group: "test"
      matchExpressions:
        - key: ffxc
          operator: NotIn
          values:
            - l
            - a
        - key: exc
          operator: NotIn
          values:
            - albc
            - a
`,
			expect: `
apiVersion: apps.kruise.io/v1alpha1
kind: ResourceDistribution
spec:
  resource:
    apiVersion: v1
    data:
      A: Yg==
      JAVA_HOME: L29wdC9qYXZhL2pkaw==
      foo1: aGVsbG8gd29ybGQ=
      foo2: aGVsbG8gd29ybGQ=
      key1: dmFsdWUx
      key2: dmFsdWUy
      key3: dmFsdWUz
    immutable: true
    kind: Secret
    metadata:
      annotations:
        dashboard: "1"
      labels:
        rsla: rs
        test: "true"
      name: cmname
    type: Opaque
  targets:
    allNamespaces: true
    includedNamespaces:
      list:
      - name: ns-1
    namespaceLabelSelector:
      matchExpressions:
      - key: exc
        operator: NotIn
        values:
        - a
        - albc
      - key: ffxc
        operator: NotIn
        values:
        - a
        - l
      matchLabels:
        group: test
metadata:
  annotations:
    an: rdan
  labels:
    app.kubernetes.io/name: app1
  name: rdname
`,
		},
		{
			name:         "secret-envs-files-resourceOptions-allNamespaces-excludedNamespaces",
			setupFile:    setupBinaryFile([]byte{0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64}),
			filename:     []string{"foo1"},
			setupEnvFile: setupEnvFile([][]string{{"key1=value1", "#", "", "key2=value2"}}),
			envFilename:  []string{"file1.env"},
			in: `
apiVersion: config.kubernetes.io/v1
kind: ResourceList
functionConfig:
  apiVersion: apps.kruise.io/v1alpha1
  kind: ResourceDistributionGenerator
  metadata:
    name: rdname
  resource:
    resourceKind: Secret
    resourceName: cmname
    envs:
      - file1.env
    files:
      - foo1
    resourceOptions:
      annotations:
        dashboard: "1"
        node: 2
      immutable: false
      labels:
        test: true
        rsla: rs
  targets:
    allNamespaces: true
    excludedNamespaces:
      - ns-2
      - ns-3
`,
			expect: `
apiVersion: apps.kruise.io/v1alpha1
kind: ResourceDistribution
spec:
  resource:
    apiVersion: v1
    data:
      foo1: aGVsbG8gd29ybGQ=
      key1: dmFsdWUx
      key2: dmFsdWUy
    kind: Secret
    metadata:
      annotations:
        dashboard: "1"
        node: "2"
      labels:
        rsla: rs
        test: "true"
      name: cmname
    type: Opaque
  targets:
    allNamespaces: true
    excludedNamespaces:
      list:
      - name: ns-2
      - name: ns-3
metadata:
  name: rdname
`,
		},
		{
			name: "secret-literals-options-allNamespaces-includedNamespaces",
			in: `
apiVersion: config.kubernetes.io/v1
kind: ResourceList
functionConfig:
  apiVersion: apps.kruise.io/v1alpha1
  kind: ResourceDistributionGenerator
  metadata:
    name: rdname
  resource:
    resourceKind: Secret
    resourceName: cmname
    literals:
      - JAVA_HOME=/opt/java/jdk
    type: Opaque
  options:
    labels:
      app.kubernetes.io/name: app1
    annotations:
      an: "rdan"
  targets:
    allNamespaces: true
    includedNamespaces:
      - ns-1
`,
			expect: `
apiVersion: apps.kruise.io/v1alpha1
kind: ResourceDistribution
spec:
  resource:
    apiVersion: v1
    data:
      JAVA_HOME: L29wdC9qYXZhL2pkaw==
    kind: Secret
    metadata:
      name: cmname
    type: Opaque
  targets:
    allNamespaces: true
    includedNamespaces:
      list:
      - name: ns-1
metadata:
  annotations:
    an: rdan
  labels:
    app.kubernetes.io/name: app1
  name: rdname
`,
		},
		{
			name: "secret-kubernetes.io/tls-literals-allNamespaces-matchExpressions",
			in: `
apiVersion: config.kubernetes.io/v1
kind: ResourceList
functionConfig:
  apiVersion: apps.kruise.io/v1alpha1
  kind: ResourceDistributionGenerator
  metadata:
    name: rdname
  resource:
    resourceKind: Secret
    resourceName: cmname
    literals:
      - tls.crt=LS0tLS1CRUd...tCg==
      - tls.key=LS0tLS1CRUd...0tLQo=
    type: kubernetes.io/tls 
  targets:
    allNamespaces: true
    namespaceLabelSelector:
      matchLabels:
        group: "test"
`,
			expect: `
apiVersion: apps.kruise.io/v1alpha1
kind: ResourceDistribution
spec:
  resource:
    apiVersion: v1
    data:
      tls.crt: TFMwdExTMUNSVWQuLi50Q2c9PQ==
      tls.key: TFMwdExTMUNSVWQuLi4wdExRbz0=
    kind: Secret
    metadata:
      name: cmname
    type: kubernetes.io/tls
  targets:
    allNamespaces: true
    namespaceLabelSelector:
      matchLabels:
        group: test
metadata:
  name: rdname
`,
		},
		{
			name:         "secret-kubernetes.io/tls-envs-files-allNamespaces-matchExpressions",
			setupEnvFile: setupEnvFile([][]string{{"tls.crt=LS0tLS1CRUd...tCg==", "tls.key=LS0tLS1CRUd...0tLQo="}}),
			envFilename:  []string{"file1.env"},
			in: `
apiVersion: config.kubernetes.io/v1
kind: ResourceList
functionConfig:
  apiVersion: apps.kruise.io/v1alpha1
  kind: ResourceDistributionGenerator
  metadata:
    name: rdname
  resource:
    resourceKind: Secret
    resourceName: cmname
    envs:
      - file1.env
    type: kubernetes.io/tls
  targets:
    allNamespaces: true
    namespaceLabelSelector:
      matchExpressions:
        - key: ffxc
          operator: Exists
        - key: exc
          operator: NotIn
          values:
            - albc
            - a

`,
			expect: `
apiVersion: apps.kruise.io/v1alpha1
kind: ResourceDistribution
spec:
  resource:
    apiVersion: v1
    data:
      tls.crt: TFMwdExTMUNSVWQuLi50Q2c9PQ==
      tls.key: TFMwdExTMUNSVWQuLi4wdExRbz0=
    kind: Secret
    metadata:
      name: cmname
    type: kubernetes.io/tls
  targets:
    allNamespaces: true
    namespaceLabelSelector:
      matchExpressions:
      - key: exc
        operator: NotIn
        values:
        - a
        - albc
      - key: ffxc
        operator: Exists
metadata:
  name: rdname
`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setupFile != nil {
				if teardown := test.setupFile(t, test.filename); teardown != nil {
					defer teardown()
				}
			}
			if test.setupEnvFile != nil {
				if teardown := test.setupEnvFile(t, test.envFilename); teardown != nil {
					defer teardown()
				}
			}

			cmd := generator.BuildCmd()
			cmd.SetIn(bytes.NewBufferString(test.in))
			buf := &bytes.Buffer{}
			cmd.SetOut(buf)
			if err := cmd.Execute(); err != nil {
				require.NoError(t, err)
			}

			//Extract resourceDistribution from the output
			out, _ := yaml.Parse(buf.String())
			items, _ := out.Pipe(yaml.Get("items"))
			elements, _ := items.Elements()
			rd, _ := elements[0].String()

			if !(strings.TrimSpace(rd) == strings.TrimSpace(test.expect)) {
				reportDiffAndFail(t, []byte(strings.TrimSpace(rd)), strings.TrimSpace(test.expect))
			}
		})
	}
}

func setupEnvFile(lines [][]string) func(*testing.T, []string) func() {
	return func(t *testing.T, EnvFileSources []string) func() {
		filenames := EnvFileSources
		for i, filename := range filenames {
			data := []byte("")
			for _, l := range lines[i] {
				data = append(data, []byte(l)...)
				data = append(data, []byte("\r\n")...)
			}
			ioutil.WriteFile(filename, data, 0644)
		}

		return func() {
			for _, file := range filenames {
				os.Remove(file)
			}
		}
	}
}

func setupBinaryFile(data []byte) func(*testing.T, []string) func() {
	return func(t *testing.T, FileSources []string) func() {
		files := FileSources
		for _, file := range files {
			ioutil.WriteFile(file, data, 0644)
		}

		return func() {
			for _, file := range files {
				os.Remove(file)
			}
		}
	}
}

// Pretty printing of file differences.
func reportDiffAndFail(
	t *testing.T, actual []byte, expected string) {
	t.Helper()
	sE, maxLen := convertToArray(expected)
	sA, _ := convertToArray(string(actual))
	fmt.Println("===== ACTUAL BEGIN ========================================")
	fmt.Print(string(actual))
	fmt.Println("===== ACTUAL END ==========================================")
	format := fmt.Sprintf("%%s  %%-%ds %%s\n", maxLen+4)
	var limit int
	if len(sE) < len(sA) {
		limit = len(sE)
	} else {
		limit = len(sA)
	}
	fmt.Printf(format, " ", "EXPECTED", "ACTUAL")
	fmt.Printf(format, " ", "--------", "------")
	for i := 0; i < limit; i++ {
		fmt.Printf(format, hint(sE[i], sA[i]), sE[i], sA[i])
	}
	if len(sE) < len(sA) {
		for i := len(sE); i < len(sA); i++ {
			fmt.Printf(format, "X", "", sA[i])
		}
	} else {
		for i := len(sA); i < len(sE); i++ {
			fmt.Printf(format, "X", sE[i], "")
		}
	}
	t.Fatalf("Expected not equal to actual")
}

func hint(a, b string) string {
	if a == b {
		return " "
	}
	return "X"
}

func convertToArray(x string) ([]string, int) {
	a := strings.Split(strings.TrimSuffix(x, "\n"), "\n")
	maxLen := 0
	for i, v := range a {
		z := tabToSpace(v)
		if len(z) > maxLen {
			maxLen = len(z)
		}
		a[i] = z
	}
	return a, maxLen
}

func tabToSpace(input string) string {
	var result []string
	for _, i := range input {
		if i == 9 {
			result = append(result, "  ")
		} else {
			result = append(result, string(i))
		}
	}
	return strings.Join(result, "")
}
