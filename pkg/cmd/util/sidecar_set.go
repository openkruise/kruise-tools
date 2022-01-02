package util

import (
	"encoding/json"
	corev1 "k8s.io/api/core/v1"
)

const (
	// SidecarSetWorkingHotUpgradeContainer record which hot upgrade container is working currently
	SidecarSetWorkingHotUpgradeContainer = "kruise.io/sidecarset-working-hotupgrade-container"
)

func GetPodHotUpgradeInfoInAnnotations(pod *corev1.Pod) map[string]string {
	hotUpgradeWorkContainer := make(map[string]string)
	currentStr, ok := pod.Annotations[SidecarSetWorkingHotUpgradeContainer]
	if !ok {
		//klog.V(6).Infof("Pod(%s.%s) annotations(%s) Not Found", pod.Namespace, pod.Name, SidecarSetWorkingHotUpgradeContainer)
		return hotUpgradeWorkContainer
	}
	if err := json.Unmarshal([]byte(currentStr), &hotUpgradeWorkContainer); err != nil {
		//klog.Errorf("Parse Pod(%s.%s) annotations(%s) Value(%s) failed: %s", pod.Namespace, pod.Name,
		//	SidecarSetWorkingHotUpgradeContainer, currentStr, err.Error())
		return hotUpgradeWorkContainer
	}
	return hotUpgradeWorkContainer
}
