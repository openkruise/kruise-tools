/*
Copyright 2021 The Kruise Authors.
Copyright 2016 The Kubernetes Authors.

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

package polymorphichelpers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"text/tabwriter"

	internalapps "github.com/openkruise/kruise-tools/pkg/internal/apps"

	kruiseappsv1alpha1 "github.com/openkruise/kruise-api/apps/v1alpha1"
	kruiseappsv1beta1 "github.com/openkruise/kruise-api/apps/v1beta1"
	kruiseclientsets "github.com/openkruise/kruise-api/client/clientset/versioned"
	kruiseclientappsv1alpha1 "github.com/openkruise/kruise-api/client/clientset/versioned/typed/apps/v1alpha1"
	kruiseclientappsv1beta1 "github.com/openkruise/kruise-api/client/clientset/versioned/typed/apps/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/kubernetes"
	clientappsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/kubectl/pkg/describe"
	deploymentutil "k8s.io/kubectl/pkg/util/deployment"
	sliceutil "k8s.io/kubectl/pkg/util/slice"
)

const (
	ChangeCauseAnnotation = "kubernetes.io/change-cause"
)

// HistoryViewer provides an interface for resources have historical information.
type HistoryViewer interface {
	ViewHistory(namespace, name string, revision int64) (string, error)
}

type HistoryVisitor struct {
	clientset       kubernetes.Interface
	kruiseclientset kruiseclientsets.Interface
	result          HistoryViewer
}

func (v *HistoryVisitor) VisitDeployment(elem internalapps.GroupKindElement) {
	v.result = &DeploymentHistoryViewer{v.clientset}
}

func (v *HistoryVisitor) VisitStatefulSet(kind internalapps.GroupKindElement) {
	v.result = &StatefulSetHistoryViewer{v.clientset}
}

func (v *HistoryVisitor) VisitDaemonSet(kind internalapps.GroupKindElement) {
	v.result = &DaemonSetHistoryViewer{v.clientset}
}

func (v *HistoryVisitor) VisitJob(kind internalapps.GroupKindElement)                   {}
func (v *HistoryVisitor) VisitPod(kind internalapps.GroupKindElement)                   {}
func (v *HistoryVisitor) VisitReplicaSet(kind internalapps.GroupKindElement)            {}
func (v *HistoryVisitor) VisitReplicationController(kind internalapps.GroupKindElement) {}
func (v *HistoryVisitor) VisitCronJob(kind internalapps.GroupKindElement)               {}
func (v *HistoryVisitor) VisitRollout(kind internalapps.GroupKindElement)               {}

// HistoryViewerFor returns an implementation of HistoryViewer interface for the given schema kind
func HistoryViewerFor(kind schema.GroupKind, c kubernetes.Interface, kc kruiseclientsets.Interface) (HistoryViewer, error) {
	elem := internalapps.GroupKindElement(kind)
	visitor := &HistoryVisitor{
		clientset:       c,
		kruiseclientset: kc,
	}

	// Determine which HistoryViewer we need here
	err := elem.Accept(visitor)

	if err != nil {
		return nil, fmt.Errorf("error retrieving history for %q, %v", kind.String(), err)
	}

	if visitor.result == nil {
		return nil, fmt.Errorf("no history viewer has been implemented for %q", kind.String())
	}

	return visitor.result, nil
}

type DeploymentHistoryViewer struct {
	c kubernetes.Interface
}

type CloneSetHistoryViewer struct {
	k  kubernetes.Interface
	kc kruiseclientsets.Interface
}

type AdvancedStatefulSetHistoryViewer struct {
	k  kubernetes.Interface
	kc kruiseclientsets.Interface
}

type AdvancedDaemonSetHistoryViewer struct {
	k  kubernetes.Interface
	kc kruiseclientsets.Interface
}

func (v *HistoryVisitor) VisitCloneSet(kind internalapps.GroupKindElement) {
	v.result = &CloneSetHistoryViewer{v.clientset, v.kruiseclientset}
}

func (v *HistoryVisitor) VisitAdvancedStatefulSet(kind internalapps.GroupKindElement) {
	v.result = &AdvancedStatefulSetHistoryViewer{v.clientset, v.kruiseclientset}
}

func (v *HistoryVisitor) VisitAdvancedDaemonSet(kind internalapps.GroupKindElement) {
	v.result = &AdvancedDaemonSetHistoryViewer{v.clientset, v.kruiseclientset}
}

// TODO impl ViewHistory func for CloneSet
func (h *CloneSetHistoryViewer) ViewHistory(namespace, name string, revision int64) (string, error) {

	cs, history, err := clonesetHistory(h.k.AppsV1(), h.kc.AppsV1alpha1(), namespace, name)
	if err != nil {
		return "", err
	}

	return printHistory(history, revision, func(history *appsv1.ControllerRevision) (*corev1.PodTemplateSpec, error) {
		cloneSetOfHistory, err := applyCloneSetHistory(cs, history)
		if err != nil {
			return nil, err
		}
		return &cloneSetOfHistory.Spec.Template, err
	})
}

func (h *AdvancedStatefulSetHistoryViewer) ViewHistory(namespace, name string, revision int64) (string, error) {
	asts, history, err := advancedstsHistory(h.k.AppsV1(), h.kc.AppsV1beta1(), namespace, name)
	if err != nil {
		return "", err
	}
	return printHistory(history, revision, func(history *appsv1.ControllerRevision) (*corev1.PodTemplateSpec, error) {
		astsOfHistory, err := applyAdvancedStatefulSetHistory(asts, history)
		if err != nil {
			return nil, err
		}
		return &astsOfHistory.Spec.Template, err
	})
}

func (h *AdvancedDaemonSetHistoryViewer) ViewHistory(namespace, name string, revision int64) (string, error) {
	ads, history, err := advancedDaemonSetHistory(h.k.AppsV1(), h.kc.AppsV1alpha1(), namespace, name)
	if err != nil {
		return "", err
	}
	return printHistory(history, revision, func(history *appsv1.ControllerRevision) (*corev1.PodTemplateSpec, error) {
		adsOfHistory, err := applyAdvancedDaemonSetHistory(ads, history)
		if err != nil {
			return nil, err
		}
		return &adsOfHistory.Spec.Template, err
	})
}

// ViewHistory returns a revision-to-replicaset map as the revision history of a deployment
// TODO: this should be a describer
func (h *DeploymentHistoryViewer) ViewHistory(namespace, name string, revision int64) (string, error) {
	versionedAppsClient := h.c.AppsV1()
	deployment, err := versionedAppsClient.Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to retrieve deployment %s: %v", name, err)
	}
	_, allOldRSs, newRS, err := deploymentutil.GetAllReplicaSets(deployment, versionedAppsClient)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve replica sets from deployment %s: %v", name, err)
	}
	allRSs := allOldRSs
	if newRS != nil {
		allRSs = append(allRSs, newRS)
	}

	historyInfo := make(map[int64]*corev1.PodTemplateSpec)
	for _, rs := range allRSs {
		v, err := deploymentutil.Revision(rs)
		if err != nil {
			continue
		}
		historyInfo[v] = &rs.Spec.Template
		changeCause := getChangeCause(rs)
		if historyInfo[v].Annotations == nil {
			historyInfo[v].Annotations = make(map[string]string)
		}
		if len(changeCause) > 0 {
			historyInfo[v].Annotations[ChangeCauseAnnotation] = changeCause
		}
	}

	if len(historyInfo) == 0 {
		return "No rollout history found.", nil
	}

	if revision > 0 {
		// Print details of a specific revision
		template, ok := historyInfo[revision]
		if !ok {
			return "", fmt.Errorf("unable to find the specified revision")
		}
		return printTemplate(template)
	}

	// Sort the revisionToChangeCause map by revision
	revisions := make([]int64, 0, len(historyInfo))
	for r := range historyInfo {
		revisions = append(revisions, r)
	}
	sliceutil.SortInts64(revisions)

	return tabbedString(func(out io.Writer) error {
		fmt.Fprintf(out, "REVISION\tCHANGE-CAUSE\n")
		for _, r := range revisions {
			// Find the change-cause of revision r
			changeCause := historyInfo[r].Annotations[ChangeCauseAnnotation]
			if len(changeCause) == 0 {
				changeCause = "<none>"
			}
			fmt.Fprintf(out, "%d\t%s\n", r, changeCause)
		}
		return nil
	})
}

func printTemplate(template *corev1.PodTemplateSpec) (string, error) {
	buf := bytes.NewBuffer([]byte{})
	w := describe.NewPrefixWriter(buf)
	describe.DescribePodTemplate(template, w)
	return buf.String(), nil
}

type DaemonSetHistoryViewer struct {
	c kubernetes.Interface
}

// ViewHistory returns a revision-to-history map as the revision history of a deployment
// TODO: this should be a describer
func (h *DaemonSetHistoryViewer) ViewHistory(namespace, name string, revision int64) (string, error) {
	ds, history, err := daemonSetHistory(h.c.AppsV1(), namespace, name)
	if err != nil {
		return "", err
	}
	return printHistory(history, revision, func(history *appsv1.ControllerRevision) (*corev1.PodTemplateSpec, error) {
		dsOfHistory, err := applyDaemonSetHistory(ds, history)
		if err != nil {
			return nil, err
		}
		return &dsOfHistory.Spec.Template, err
	})
}

// printHistory returns the podTemplate of the given revision if it is non-zero
// else returns the overall revisions
func printHistory(history []*appsv1.ControllerRevision, revision int64, getPodTemplate func(history *appsv1.ControllerRevision) (*corev1.PodTemplateSpec, error)) (string, error) {
	historyInfo := make(map[int64]*appsv1.ControllerRevision)
	for _, history := range history {
		// TODO: for now we assume revisions don't overlap, we may need to handle it
		historyInfo[history.Revision] = history
	}
	if len(historyInfo) == 0 {
		return "No rollout history found.", nil
	}

	// Print details of a specific revision
	if revision > 0 {
		history, ok := historyInfo[revision]
		if !ok {
			return "", fmt.Errorf("unable to find the specified revision")
		}
		podTemplate, err := getPodTemplate(history)
		if err != nil {
			return "", fmt.Errorf("unable to parse history %s", history.Name)
		}
		return printTemplate(podTemplate)
	}

	// Print an overview of all Revisions
	// Sort the revisionToChangeCause map by revision
	revisions := make([]int64, 0, len(historyInfo))
	for r := range historyInfo {
		revisions = append(revisions, r)
	}
	sliceutil.SortInts64(revisions)

	return tabbedString(func(out io.Writer) error {
		fmt.Fprintf(out, "REVISION\tCHANGE-CAUSE\n")
		for _, r := range revisions {
			// Find the change-cause of revision r
			changeCause := historyInfo[r].Annotations[ChangeCauseAnnotation]
			if len(changeCause) == 0 {
				changeCause = "<none>"
			}
			fmt.Fprintf(out, "%d\t%s\n", r, changeCause)
		}
		return nil
	})
}

type StatefulSetHistoryViewer struct {
	c kubernetes.Interface
}

// ViewHistory returns a list of the revision history of a statefulset
// TODO: this should be a describer
func (h *StatefulSetHistoryViewer) ViewHistory(namespace, name string, revision int64) (string, error) {
	sts, history, err := statefulSetHistory(h.c.AppsV1(), namespace, name)
	if err != nil {
		return "", err
	}
	return printHistory(history, revision, func(history *appsv1.ControllerRevision) (*corev1.PodTemplateSpec, error) {
		stsOfHistory, err := applyStatefulSetHistory(sts, history)
		if err != nil {
			return nil, err
		}
		return &stsOfHistory.Spec.Template, err
	})
}

// controlledHistories returns all ControllerRevisions in namespace that selected by selector and owned by accessor
// TODO: Rename this to controllerHistory when other controllers have been upgraded
func controlledHistoryV1(
	apps clientappsv1.AppsV1Interface,
	namespace string,
	selector labels.Selector,
	accessor metav1.Object) ([]*appsv1.ControllerRevision, error) {
	var result []*appsv1.ControllerRevision
	historyList, err := apps.ControllerRevisions(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return nil, err
	}
	for i := range historyList.Items {
		history := historyList.Items[i]
		// Only add history that belongs to the API object
		if metav1.IsControlledBy(&history, accessor) {
			result = append(result, &history)
		}
	}
	return result, nil
}

// controlledHistories returns all ControllerRevisions in namespace that selected by selector and owned by accessor
func controlledHistory(
	apps clientappsv1.AppsV1Interface,
	namespace string,
	selector labels.Selector,
	accessor metav1.Object) ([]*appsv1.ControllerRevision, error) {
	var result []*appsv1.ControllerRevision
	historyList, err := apps.ControllerRevisions(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return nil, err
	}
	for i := range historyList.Items {
		history := historyList.Items[i]
		// Only add history that belongs to the API object
		if metav1.IsControlledBy(&history, accessor) {
			result = append(result, &history)
		}
	}
	return result, nil
}

// daemonSetHistory returns the DaemonSet named name in namespace and all ControllerRevisions in its history.
func daemonSetHistory(
	apps clientappsv1.AppsV1Interface,
	namespace, name string) (*appsv1.DaemonSet, []*appsv1.ControllerRevision, error) {
	ds, err := apps.DaemonSets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve DaemonSet %s: %v", name, err)
	}
	selector, err := metav1.LabelSelectorAsSelector(ds.Spec.Selector)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create selector for DaemonSet %s: %v", ds.Name, err)
	}
	accessor, err := meta.Accessor(ds)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create accessor for DaemonSet %s: %v", ds.Name, err)
	}
	history, err := controlledHistory(apps, ds.Namespace, selector, accessor)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to find history controlled by DaemonSet %s: %v", ds.Name, err)
	}
	return ds, history, nil
}

// advancedDaemonSetHistory returns the Advanced DaemonSet named name in namespace and all ControllerRevisions in its history.
func advancedDaemonSetHistory(
	apps clientappsv1.AppsV1Interface, appsv1alpha1 kruiseclientappsv1alpha1.AppsV1alpha1Interface,
	namespace, name string) (*kruiseappsv1alpha1.DaemonSet, []*appsv1.ControllerRevision, error) {
	ds, err := appsv1alpha1.DaemonSets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve DaemonSet %s: %v", name, err)
	}
	selector, err := metav1.LabelSelectorAsSelector(ds.Spec.Selector)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create selector for DaemonSet %s: %v", ds.Name, err)
	}
	accessor, err := meta.Accessor(ds)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create accessor for DaemonSet %s: %v", ds.Name, err)
	}
	history, err := controlledHistoryV1(apps, ds.Namespace, selector, accessor)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to find history controlled by DaemonSet %s: %v", ds.Name, err)
	}
	return ds, history, nil
}

func clonesetHistory(
	apps clientappsv1.AppsV1Interface, appsv1alpha1 kruiseclientappsv1alpha1.AppsV1alpha1Interface,
	namespace, name string) (*kruiseappsv1alpha1.CloneSet, []*appsv1.ControllerRevision, error) {
	cs, err := appsv1alpha1.CloneSets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, nil, err
	}
	selector, err := metav1.LabelSelectorAsSelector(cs.Spec.Selector)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create selector for StatefulSet %s: %s", name, err.Error())
	}
	accessor, err := meta.Accessor(cs)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to obtain accessor for StatefulSet %s: %s", name, err.Error())
	}
	history, err := controlledHistoryV1(apps, namespace, selector, accessor)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to find history controlled by StatefulSet %s: %v", name, err)
	}

	return cs, history, nil
}

func advancedstsHistory(
	apps clientappsv1.AppsV1Interface, appsv1beta1 kruiseclientappsv1beta1.AppsV1beta1Interface,
	namespace, name string) (*kruiseappsv1beta1.StatefulSet, []*appsv1.ControllerRevision, error) {
	asts, err := appsv1beta1.StatefulSets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, nil, err
	}
	selector, err := metav1.LabelSelectorAsSelector(asts.Spec.Selector)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create selector for Advanced StatefulSet %s: %s", name, err.Error())
	}
	accessor, err := meta.Accessor(asts)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to obtain accessor for Advanced  StatefulSet %s: %s", name, err.Error())
	}

	history, err := controlledHistoryV1(apps, namespace, selector, accessor)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to find history controlled by StatefulSet %s: %v", name, err)
	}
	return asts, history, nil
}

// statefulSetHistory returns the StatefulSet named name in namespace and all ControllerRevisions in its history.
func statefulSetHistory(
	apps clientappsv1.AppsV1Interface,
	namespace, name string) (*appsv1.StatefulSet, []*appsv1.ControllerRevision, error) {
	sts, err := apps.StatefulSets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve Statefulset %s: %s", name, err.Error())
	}
	selector, err := metav1.LabelSelectorAsSelector(sts.Spec.Selector)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create selector for StatefulSet %s: %s", name, err.Error())
	}
	accessor, err := meta.Accessor(sts)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to obtain accessor for StatefulSet %s: %s", name, err.Error())
	}
	history, err := controlledHistoryV1(apps, namespace, selector, accessor)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to find history controlled by StatefulSet %s: %v", name, err)
	}
	return sts, history, nil
}

// applyDaemonSetHistory returns a specific revision of DaemonSet by applying the given history to a copy of the given DaemonSet
func applyDaemonSetHistory(ds *appsv1.DaemonSet, history *appsv1.ControllerRevision) (*appsv1.DaemonSet, error) {
	dsBytes, err := json.Marshal(ds)
	if err != nil {
		return nil, err
	}
	patched, err := strategicpatch.StrategicMergePatch(dsBytes, history.Data.Raw, ds)
	if err != nil {
		return nil, err
	}
	result := &appsv1.DaemonSet{}
	err = json.Unmarshal(patched, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// applyStatefulSetHistory returns a specific revision of StatefulSet by applying the given history to a copy of the given StatefulSet
func applyStatefulSetHistory(sts *appsv1.StatefulSet, history *appsv1.ControllerRevision) (*appsv1.StatefulSet, error) {
	stsBytes, err := json.Marshal(sts)
	if err != nil {
		return nil, err
	}
	patched, err := strategicpatch.StrategicMergePatch(stsBytes, history.Data.Raw, sts)
	if err != nil {
		return nil, err
	}
	result := &appsv1.StatefulSet{}
	err = json.Unmarshal(patched, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// applyCloneSetHistory returns a specific revision of CloneSet by applying the given history to a copy of the given CloneSet
func applyCloneSetHistory(cs *kruiseappsv1alpha1.CloneSet,
	history *appsv1.ControllerRevision) (*kruiseappsv1alpha1.CloneSet, error) {
	csBytes, err := json.Marshal(cs)
	if err != nil {
		return nil, err
	}
	patched, err := strategicpatch.StrategicMergePatch(csBytes, history.Data.Raw, cs)

	if err != nil {
		return nil, err
	}
	result := &kruiseappsv1alpha1.CloneSet{}

	err = json.Unmarshal(patched, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func applyAdvancedStatefulSetHistory(asts *kruiseappsv1beta1.StatefulSet,
	history *appsv1.ControllerRevision) (*kruiseappsv1beta1.StatefulSet, error) {
	astsBytes, err := json.Marshal(asts)
	if err != nil {
		return nil, err
	}
	patched, err := strategicpatch.StrategicMergePatch(astsBytes, history.Data.Raw, asts)
	if err != nil {
		return nil, err
	}
	result := &kruiseappsv1beta1.StatefulSet{}
	err = json.Unmarshal(patched, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func applyAdvancedDaemonSetHistory(ads *kruiseappsv1alpha1.DaemonSet,
	history *appsv1.ControllerRevision) (*kruiseappsv1alpha1.DaemonSet, error) {
	adsBytes, err := json.Marshal(ads)
	if err != nil {
		return nil, err
	}
	patched, err := strategicpatch.StrategicMergePatch(adsBytes, history.Data.Raw, ads)
	if err != nil {
		return nil, err
	}
	result := &kruiseappsv1alpha1.DaemonSet{}
	err = json.Unmarshal(patched, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// TODO: copied here until this becomes a describer
func tabbedString(f func(io.Writer) error) (string, error) {
	out := new(tabwriter.Writer)
	buf := &bytes.Buffer{}
	out.Init(buf, 0, 8, 2, ' ', 0)

	err := f(out)
	if err != nil {
		return "", err
	}

	out.Flush()
	str := buf.String()
	return str, nil
}

// getChangeCause returns the change-cause annotation of the input object
func getChangeCause(obj runtime.Object) string {
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return ""
	}
	return accessor.GetAnnotations()[ChangeCauseAnnotation]
}
