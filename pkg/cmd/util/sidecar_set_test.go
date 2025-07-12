package util

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetPodHotUpgradeInfoInAnnotations(t *testing.T) {
	testCases := []struct {
		name     string
		pod      *corev1.Pod
		expected map[string]string
	}{
		{
			name: "Pod with valid hot-upgrade annotation",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						SidecarSetWorkingHotUpgradeContainer: `{"container1":"image-v2"}`,
					},
				},
			},
			expected: map[string]string{
				"container1": "image-v2",
			},
		},
		{
			name: "Pod without hot-upgrade annotation",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"another-annotation": "some-value",
					},
				},
			},
			expected: map[string]string{},
		},
		{
			name: "Pod with empty annotations",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
			},
			expected: map[string]string{},
		},
		{
			name: "Pod with nil annotations",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: nil,
				},
			},
			expected: map[string]string{},
		},
		{
			name: "Pod with invalid JSON in annotation",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						SidecarSetWorkingHotUpgradeContainer: `{"container1":"image-v2"`, // Malformed JSON
					},
				},
			},
			expected: map[string]string{},
		},
		{
			name: "Pod with empty JSON object in annotation",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						SidecarSetWorkingHotUpgradeContainer: `{}`,
					},
				},
			},
			expected: map[string]string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Test panicked: %v", r)
				}
			}()

			result := GetPodHotUpgradeInfoInAnnotations(tc.pod)

			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("expected: %v, got: %v", tc.expected, result)
			}
		})
	}
}
