/*
Copyright 2025 The Kruise Authors.

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

package top

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"
)

// SumUsageForSelector sums the resource usage for all pods matching a given selector.
func SumUsageForSelector(kubeClient kubernetes.Interface, metricsClient metricsclientset.Interface, namespace string, selector labels.Selector) (*resource.Quantity, *resource.Quantity, error) {
	podList, err := kubeClient.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list pods for selector %q: %v", selector.String(), err)
	}

	if len(podList.Items) == 0 {
		return resource.NewMilliQuantity(0, resource.DecimalSI), resource.NewQuantity(0, resource.BinarySI), nil
	}

	podMetricsList, err := metricsClient.MetricsV1beta1().PodMetricses(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get pod metrics: %v", err)
	}

	metricsMap := make(map[string]corev1.ResourceList)
	for _, m := range podMetricsList.Items {
		totalUsage := corev1.ResourceList{}
		for _, c := range m.Containers {
			for resName, resQuant := range c.Usage {
				if current, ok := totalUsage[resName]; ok {
					current.Add(resQuant)
					totalUsage[resName] = current
				} else {
					totalUsage[resName] = resQuant
				}
			}
		}
		metricsMap[m.Name] = totalUsage
	}

	totalCPU := &resource.Quantity{}
	totalMemory := &resource.Quantity{}
	for _, pod := range podList.Items {
		if metrics, ok := metricsMap[pod.Name]; ok {
			if cpu, found := metrics[corev1.ResourceCPU]; found {
				totalCPU.Add(cpu)
			}
			if memory, found := metrics[corev1.ResourceMemory]; found {
				totalMemory.Add(memory)
			}
		}
	}

	return totalCPU, totalMemory, nil
}

// FormatResourceUsage formats the resource quantities for printing.
func FormatResourceUsage(cpu *resource.Quantity, memory *resource.Quantity) (string, string) {
	cpuString := fmt.Sprintf("%vm", cpu.MilliValue())
	memoryString := fmt.Sprintf("%dMi", memory.Value()/(1024*1024))
	return cpuString, memoryString
}
