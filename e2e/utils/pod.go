/*
   MIT License

   Copyright (c) Microsoft Corporation.

   Permission is hereby granted, free of charge, to any person obtaining a copy
   of this software and associated documentation files (the "Software"), to deal
   in the Software without restriction, including without limitation the rights
   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
   copies of the Software, and to permit persons to whom the Software is
   furnished to do so, subject to the following conditions:

   The above copyright notice and this permission notice shall be included in all
   copies or substantial portions of the Software.

   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
   SOFTWARE
*/

package utils

import (
	"context"
	"fmt"
	"regexp"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateCurlPodManifest(nsName, gwName, curlTarget string) *corev1.Pod {
	annotations := make(map[string]string)
	if gwName != "" {
		annotations["kubernetes.azure.com/static-gateway-configuration"] = gwName
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "agnhost-pod-" + string(uuid.NewUUID())[0:4],
			Namespace:   nsName,
			Annotations: annotations,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:            "app",
					Image:           "registry.k8s.io/e2e-test-images/agnhost:2.36",
					ImagePullPolicy: v1.PullIfNotPresent,
					Command: []string{
						"/bin/sh", "-c", "curl -s -m 5 --retry-delay 60 --retry 10 " + curlTarget,
					},
				},
			},
			RestartPolicy: v1.RestartPolicyNever,
		},
	}
}

func CreateNginxPodManifest(nsName, gwName string) *corev1.Pod {
	annotations := make(map[string]string)
	if gwName != "" {
		annotations["kubernetes.azure.com/static-gateway-configuration"] = gwName
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "nginx-pod-" + string(uuid.NewUUID())[0:4],
			Namespace:   nsName,
			Annotations: annotations,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:            "app",
					Image:           "nginx",
					ImagePullPolicy: v1.PullIfNotPresent,
					Ports:           []v1.ContainerPort{{ContainerPort: 80}},
				},
			},
			RestartPolicy: v1.RestartPolicyNever,
		},
	}
}

func GetExpectedPodLog(pod *v1.Pod, c clientset.Interface, expectLogRegex *regexp.Regexp) (string, error) {
	var log []byte
	err := wait.PollUntilContextTimeout(context.Background(), poll, pollTimeoutForProvision, true, func(ctx context.Context) (bool, error) {
		pod, err := c.CoreV1().Pods(pod.Namespace).Get(context.Background(), pod.Name, metav1.GetOptions{})
		if err != nil {
			if retriable(err) {
				return false, nil
			}
			return false, err
		}
		if pod.Status.Phase != v1.PodSucceeded {
			Logf("Waiting for the pod to succeed, current status: %s", pod.Status.Phase)
			if pod.Status.Phase == v1.PodFailed {
				return false, fmt.Errorf("test pod is in Failed phase")
			}
			return false, nil
		}
		if pod.Status.ContainerStatuses[0].State.Terminated == nil || pod.Status.ContainerStatuses[0].State.Terminated.Reason != "Completed" {
			Logf("Waiting for the container to be completed")
			return false, nil
		}
		log, err = c.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &v1.PodLogOptions{}).Do(context.Background()).Raw()
		if err != nil {
			Logf("Got %v when retrieving test pod log, retrying", err)
			return false, nil
		}
		return expectLogRegex.MatchString(string(log)), nil
	})
	if err != nil {
		return "", err
	}
	found := expectLogRegex.FindString(string(log))
	return found, nil
}

func WaitGetPodIP(pod *v1.Pod, c client.Client) (string, error) {
	var podIP string
	key := types.NamespacedName{
		Name:      pod.Name,
		Namespace: pod.Namespace,
	}
	err := wait.PollUntilContextTimeout(context.Background(), poll, pollTimeout, true, func(ctx context.Context) (bool, error) {
		err := c.Get(ctx, key, pod)
		if err != nil {
			if retriable(err) {
				return false, nil
			}
			return false, err
		}
		if pod.Status.Phase == v1.PodRunning {
			podIP = pod.Status.PodIP
			return true, nil
		}
		Logf("Waiting for the pod to Running, current status: %s", pod.Status.Phase)
		return false, nil
	})
	return podIP, err
}