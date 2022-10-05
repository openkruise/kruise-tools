/*
Copyright 2022 The Kruise Authors.
Copyright 2020 The Kubernetes Authors.

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

package generator

import (
	"encoding/base64"
	"strings"
	"unicode/utf8"

	"github.com/go-errors/errors"
	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/kv"
	"sigs.k8s.io/kustomize/api/loader"
	"sigs.k8s.io/kustomize/api/provider"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func makeBaseNode(meta *types.ObjectMeta) (*yaml.RNode, error) {
	rn, err := yaml.Parse(tmpl)
	if err != nil {
		return nil, err
	}

	return rn, nil
}

func setResource(rn *yaml.RNode, args *ResourceArgs) error {
	if err := setData(rn, args); err != nil {
		return err
	}

	if err := setImmutable(rn, args.ResourceOptions); err != nil {
		return err
	}
	if err := setResourceKind(rn, args.ResourceKind); err != nil {
		return err
	}
	// set Labels And Annotions for Secret or ConfigMap
	if args.ResourceOptions != nil {
		if err := setLabelsOrAnnotations(rn, args.ResourceOptions.Annotations, resourceAnnotationsPath); err != nil {
			return err
		}
		if err := setLabelsOrAnnotations(rn, args.ResourceOptions.Labels, resourceLabelsPath); err != nil {
			return err
		}
	}

	if err := setResourceName(rn, args.ResourceName); err != nil {
		return err
	}
	if err := setResourceType(rn, args); err != nil {
		return err
	}
	return nil
}

func setResourceName(rn *yaml.RNode, name string) error {
	if name == "" {
		return errors.Errorf("a ResourceDistribution must have a resource name ")
	}
	if err := rn.SetMapField(yaml.NewStringRNode(name), append(metadataPath, nameField)...); err != nil {
		return err
	}
	return nil
}

func setResourceType(rn *yaml.RNode, args *ResourceArgs) error {
	if args.ResourceKind == "Secret" {
		t := "Opaque"
		if args.Type != "" {
			t = args.Type
		}
		if err := rn.SetMapField(yaml.NewStringRNode(t), append(resourcePath, typeField)...); err != nil {
			return err
		}
	}
	return nil
}

func setResourceKind(
	rn *yaml.RNode, kind string) error {
	if kind == "" || kind != "Secret" && kind != "ConfigMap" {
		return errors.Errorf("resourceKind must be ConfigMap or Secret ")
	}
	if err := rn.SetMapField(yaml.NewStringRNode(kind), append(resourcePath, kindField)...); err != nil {
		return err
	}
	return nil
}

func setLabelsOrAnnotations(
	rn *yaml.RNode, labelsOrAnnotations map[string]string, labelsOrAnnotationsPath []string) error {
	if labelsOrAnnotations == nil {
		return nil
	}

	for _, k := range yaml.SortedMapKeys(labelsOrAnnotations) {
		v := labelsOrAnnotations[k]
		if err := rn.SetMapField(yaml.NewStringRNode(v), append(labelsOrAnnotationsPath, k)...); err != nil {
			return err
		}
	}
	return nil
}

func setData(rn *yaml.RNode, args *ResourceArgs) error {
	ldr, err := loader.NewLoader(loader.RestrictionRootOnly,
		"./", filesys.MakeFsOnDisk())
	if err != nil {
		return err
	}
	kvLdr := kv.NewLoader(ldr, provider.NewDefaultDepProvider().GetFieldValidator())

	m, err := makeValidatedDataMap(kvLdr, args.ResourceName, args.KvPairSources)
	if err != nil {
		return err
	}

	if args.ResourceKind == "ConfigMap" {
		if err = loadMapIntoConfigMapData(m, rn); err != nil {
			return err
		}
	} else {
		if err = loadMapIntoSecretData(m, rn); err != nil {
			return err
		}
	}
	return nil
}

// copy from sigs.k8s.io/kustomize/api/internal/generators/utils.go
func makeValidatedDataMap(
	ldr ifc.KvLoader, name string, sources types.KvPairSources) (map[string]string, error) {
	pairs, err := ldr.Load(sources)
	if err != nil {
		return nil, errors.WrapPrefix(err, "loading KV pairs", 0)
	}
	knownKeys := make(map[string]string)
	for _, p := range pairs {
		// legal key: alphanumeric characters, '-', '_' or '.'
		if err := ldr.Validator().ErrIfInvalidKey(p.Key); err != nil {
			return nil, err
		}
		if _, ok := knownKeys[p.Key]; ok {
			return nil, errors.Errorf(
				"configmap %s illegally repeats the key `%s`", name, p.Key)
		}
		knownKeys[p.Key] = p.Value
	}
	return knownKeys, nil
}

func setImmutable(
	rn *yaml.RNode, opts *types.GeneratorOptions) error {
	if opts == nil {
		return nil
	}
	if opts.Immutable {
		n := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: "true",
			Tag:   yaml.NodeTagBool,
		}
		if err := rn.SetMapField(yaml.NewRNode(n), append(resourcePath, immutableField)...); err != nil {
			return err
		}
	}
	return nil
}

// copy from sigs.k8s.io/kustomize/kyaml/yaml/datamap.go
// The resourcePath prefix is added to the fldName
func loadMapIntoConfigMapData(m map[string]string, rn *yaml.RNode) error {
	for _, k := range yaml.SortedMapKeys(m) {
		fldName, vrN := makeConfigMapValueRNode(m[k])
		if _, err := rn.Pipe(
			yaml.LookupCreate(yaml.MappingNode, append(resourcePath, fldName)...),
			yaml.SetField(k, vrN)); err != nil {
			return err
		}
	}
	return nil
}

// copy from sigs.k8s.io/kustomize/kyaml/yaml/datamap.go
func makeConfigMapValueRNode(s string) (field string, rN *yaml.RNode) {
	yN := &yaml.Node{Kind: yaml.ScalarNode}
	yN.Tag = yaml.NodeTagString
	if utf8.ValidString(s) {
		field = yaml.DataField
		yN.Value = s
	} else {
		field = yaml.BinaryDataField
		yN.Value = encodeBase64(s)
	}
	if strings.Contains(yN.Value, "\n") {
		yN.Style = yaml.LiteralStyle
	}
	return field, yaml.NewRNode(yN)
}

// copy from sigs.k8s.io/kustomize/kyaml/yaml/datamap.go
func encodeBase64(s string) string {
	const lineLen = 70
	encLen := base64.StdEncoding.EncodedLen(len(s))
	lines := encLen/lineLen + 1
	buf := make([]byte, encLen*2+lines)
	in := buf[0:encLen]
	out := buf[encLen:]
	base64.StdEncoding.Encode(in, []byte(s))
	k := 0
	for i := 0; i < len(in); i += lineLen {
		j := i + lineLen
		if j > len(in) {
			j = len(in)
		}
		k += copy(out[k:], in[i:j])
		if lines > 1 {
			out[k] = '\n'
			k++
		}
	}
	return string(out[:k])
}

// copy from sigs.k8s.io/kustomize/kyaml/yaml/datamap.go
// The resourcePath prefix is added to the yaml.DataField
func loadMapIntoSecretData(m map[string]string, rn *yaml.RNode) error {
	mapNode, err := rn.Pipe(yaml.LookupCreate(yaml.MappingNode, append(resourcePath, yaml.DataField)...))
	if err != nil {
		return err
	}
	for _, k := range yaml.SortedMapKeys(m) {
		vrN := makeSecretValueRNode(m[k])
		if _, err := mapNode.Pipe(yaml.SetField(k, vrN)); err != nil {
			return err
		}
	}
	return nil
}

// copy from sigs.k8s.io/kustomize/kyaml/yaml/datamap.go
func makeSecretValueRNode(s string) *yaml.RNode {
	yN := &yaml.Node{Kind: yaml.ScalarNode}
	// Purposely don't use YAML tags to identify the data as being plain text or
	// binary.  It kubernetes Secrets the values in the `data` map are expected
	// to be base64 encoded, and in ConfigMaps that same can be said for the
	// values in the `binaryData` field.
	yN.Tag = yaml.NodeTagString
	yN.Value = encodeBase64(s)
	if strings.Contains(yN.Value, "\n") {
		yN.Style = yaml.LiteralStyle
	}
	return yaml.NewRNode(yN)
}
