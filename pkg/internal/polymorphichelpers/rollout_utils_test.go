package polymorphichelpers

import (
	"testing"

	kruiseappsv1alpha1 "github.com/openkruise/kruise-api/apps/v1alpha1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	intstrutil "k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/fake"
)

func newTestPod(name, namespace, revision string, labels map[string]string, isReady bool) *corev1.Pod {
	// Create a new map to avoid modifying the original
	podLabels := make(map[string]string)
	for k, v := range labels {
		podLabels[k] = v
	}
	podLabels["controller-revision-hash"] = revision

	readyCondition := corev1.ConditionFalse
	if isReady {
		readyCondition = corev1.ConditionTrue
	}

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    podLabels,
		},
		Status: corev1.PodStatus{
			Conditions: []corev1.PodCondition{
				{
					Type:   corev1.PodReady,
					Status: readyCondition,
				},
			},
		},
	}
}

func TestGetPodsByLabelSelector(t *testing.T) {
	pod1 := newTestPod("pod-1", "test-ns", "rev1", map[string]string{"app": "my-app"}, true)
	pod2 := newTestPod("pod-2", "test-ns", "rev1", map[string]string{"app": "other-app"}, true)

	fakeClient := fake.NewSimpleClientset(pod1, pod2)

	selector := &metav1.LabelSelector{MatchLabels: map[string]string{"app": "my-app"}}
	pods, err := getPodsByLabelSelector(fakeClient, "test-ns", selector)

	assert.NoError(t, err)
	assert.Len(t, pods, 1, "Should find exactly one pod")
	assert.Equal(t, "pod-1", pods[0].Name, "The found pod should be pod-1")
}

func TestFilterOldNewReadyPodsFromCloneSet(t *testing.T) {
	clone := &kruiseappsv1alpha1.CloneSet{
		ObjectMeta: metav1.ObjectMeta{Name: "my-clone", Namespace: "test-ns"},
		Spec:       kruiseappsv1alpha1.CloneSetSpec{Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "my-clone"}}},
		Status:     kruiseappsv1alpha1.CloneSetStatus{UpdateRevision: "rev-new"},
	}

	oldPod := newTestPod("old-pod", "test-ns", "rev-old", clone.Spec.Selector.MatchLabels, true)
	newReadyPod := newTestPod("new-ready-pod", "test-ns", "rev-new", clone.Spec.Selector.MatchLabels, true)
	newNotReadyPod := newTestPod("new-not-ready-pod", "test-ns", "rev-new", clone.Spec.Selector.MatchLabels, false)

	fakeClient := fake.NewSimpleClientset(oldPod, newReadyPod, newNotReadyPod)

	oldPods, newNotReadyPods, updatedReadyPods, err := filterOldNewReadyPodsFromCloneSet(fakeClient, clone)

	assert.NoError(t, err)
	assert.Len(t, oldPods, 1, "Should be one old pod")
	assert.Equal(t, "old-pod", oldPods[0].Name)

	assert.Len(t, newNotReadyPods, 1, "Should be one new not-ready pod")
	assert.Equal(t, "new-not-ready-pod", newNotReadyPods[0].Name)

	assert.Len(t, updatedReadyPods, 1, "Should be one updated ready pod")
	assert.Equal(t, "new-ready-pod", updatedReadyPods[0].Name)
}

func TestPodReady(t *testing.T) {
	testCases := []struct {
		name     string
		pod      *corev1.Pod
		expected bool
	}{
		{"Pod with Ready condition true", newTestPod("p1", "ns", "rev", nil, true), true},
		{"Pod with Ready condition false", newTestPod("p2", "ns", "rev", nil, false), false},
		{
			"Pod without Ready condition",
			&corev1.Pod{Status: corev1.PodStatus{Conditions: []corev1.PodCondition{{Type: corev1.PodInitialized, Status: corev1.ConditionTrue}}}},
			false,
		},
		{"Pod with no conditions", &corev1.Pod{}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, podReady(tc.pod))
		})
	}
}

func TestGeneratePodsInfoForCloneSet(t *testing.T) {
	clone := &kruiseappsv1alpha1.CloneSet{
		ObjectMeta: metav1.ObjectMeta{Name: "my-clone", Namespace: "test-ns"},
		Spec:       kruiseappsv1alpha1.CloneSetSpec{Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "my-clone"}}},
		Status:     kruiseappsv1alpha1.CloneSetStatus{UpdateRevision: "rev-new"},
	}

	readyPod := newTestPod("ready-1", "test-ns", "rev-new", clone.Spec.Selector.MatchLabels, true)
	notReadyPod := newTestPod("not-ready-1", "test-ns", "rev-new", clone.Spec.Selector.MatchLabels, false)
	fakeClient := fake.NewSimpleClientset(readyPod, notReadyPod)

	result := generatePodsInfoForCloneSet(fakeClient, clone)
	expected := "Updated ready pods: [ready-1]\nUpdated not ready pods: [not-ready-1]\n"

	assert.Equal(t, expected, result)
}

func TestCalculatePartitionReplicas(t *testing.T) {
	replicas := int32(10)
	str_10_percent := intstrutil.FromString("10%")
	str_99_percent := intstrutil.FromString("99%")
	str_100_percent := intstrutil.FromString("100%")
	int_2 := intstrutil.FromInt(2)

	testCases := []struct {
		name      string
		partition *intstrutil.IntOrString
		replicas  *int32
		expected  int
		expectErr bool
	}{
		{"Nil partition", nil, &replicas, 0, false},
		{"Nil replicas", &int_2, nil, 1, false}, // Defaults to 1 replica
		{"Absolute value", &int_2, &replicas, 2, false},
		{"10% of 10 replicas", &str_10_percent, &replicas, 1, false},
		{"99% of 10 replicas", &str_99_percent, &replicas, 9, false}, // Rounds up to 10, then decremented to 9
		{"100% of 10 replicas", &str_100_percent, &replicas, 10, false},
		{"Invalid format", &intstrutil.IntOrString{Type: intstrutil.String, StrVal: "abc"}, &replicas, 0, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := CalculatePartitionReplicas(tc.partition, tc.replicas)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}
