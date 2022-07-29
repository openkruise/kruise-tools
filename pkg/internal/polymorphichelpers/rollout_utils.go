package polymorphichelpers

import (
	"context"
	"fmt"

	kruiseappsv1alpha1 "github.com/openkruise/kruise-api/apps/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	intstrutil "k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/integer"
)

func getPodsByLabelSelector(client kubernetes.Interface, ns string, labelSelector *metav1.LabelSelector) ([]*corev1.Pod, error) {
	var podsList []*corev1.Pod
	pods, err := client.CoreV1().Pods(ns).List(context.TODO(), metav1.ListOptions{LabelSelector: labels.Set(labelSelector.MatchLabels).String()})
	if err != nil {
		return nil, err
	}

	for i := range pods.Items {
		podsList = append(podsList, &pods.Items[i])
	}

	return podsList, nil
}

func filterOldNewReadyPodsFromCloneSet(client kubernetes.Interface, clone *kruiseappsv1alpha1.CloneSet) (oldPods []*corev1.Pod,
	newNotReadyPods []*corev1.Pod, updatedReadyPods []*corev1.Pod, err error) {
	pods, err := getPodsByLabelSelector(client, clone.Namespace, clone.Spec.Selector)
	if err != nil {
		return
	}

	for i := range pods {
		if podRevision, ok := pods[i].GetLabels()["controller-revision-hash"]; ok {
			if podRevision == clone.Status.UpdateRevision {
				if podReady(pods[i]) {
					updatedReadyPods = append(updatedReadyPods, pods[i])
				} else {
					newNotReadyPods = append(newNotReadyPods, pods[i])
				}
			} else {
				oldPods = append(oldPods, pods[i])
			}
		}
	}
	return
}
func podReady(p *corev1.Pod) bool {
	cs := p.Status.Conditions
	for _, c := range cs {
		if c.Type == corev1.PodReady {
			return c.Status == corev1.ConditionTrue
		}
	}
	return false
}

func generatePodsInfoForCloneSet(client kubernetes.Interface, clone *kruiseappsv1alpha1.CloneSet) string {
	var notReadyPodsSlice, ReadyPodsSlice []string
	_, notReadyNewPods, readyNewPods, err := filterOldNewReadyPodsFromCloneSet(client, clone)
	if err != nil {
		return ""
	}
	for i := range notReadyNewPods {
		notReadyPodsSlice = append(notReadyPodsSlice, notReadyNewPods[i].Name)
	}
	for j := range readyNewPods {
		ReadyPodsSlice = append(ReadyPodsSlice, readyNewPods[j].Name)
	}

	return fmt.Sprintf("Updated ready pods: %v\nUpdated not ready pods: %v\n", ReadyPodsSlice, notReadyPodsSlice)
}

// CalculatePartitionReplicas returns absolute value of partition for workload. This func can solve some
// corner cases about percentage-type partition, such as:
// - if partition > "0%" and replicas > 0, we will ensure at least 1 old pod is reserved.
// - if partition < "100%" and replicas > 1, we will ensure at least 1 pod is upgraded.
func CalculatePartitionReplicas(partition *intstrutil.IntOrString, replicasPointer *int32) (int, error) {
	if partition == nil {
		return 0, nil
	}

	replicas := 1
	if replicasPointer != nil {
		replicas = int(*replicasPointer)
	}

	// 'roundUp=true' will ensure at least 1 old pod is reserved if partition > "0%" and replicas > 0.
	pValue, err := intstrutil.GetScaledValueFromIntOrPercent(partition, replicas, true)
	if err != nil {
		return pValue, err
	}

	// if partition < "100%" and replicas > 1, we will ensure at least 1 pod is upgraded.
	if replicas > 1 && pValue == replicas && partition.Type == intstrutil.String && partition.StrVal != "100%" {
		pValue = replicas - 1
	}

	pValue = integer.IntMax(integer.IntMin(pValue, replicas), 0)
	return pValue, nil
}
