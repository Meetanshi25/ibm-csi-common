/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * IBM Cloud Kubernetes Service, 5737-D43
 * (C) Copyright IBM Corp. 2024 All Rights Reserved.
 * The source code for this program is not published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/

package e2e

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var kubeClient *kubernetes.Clientset
var namespace string = "kube-system" // Change this to your namespace
var fileLock sync.Mutex
var testResultFile string

const (
	workerPoolLabelKey = "ibm-cloud.kubernetes.io/worker-pool-name"
	hostnamekey        = "kubernetes.io/hostname"
)

func init() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting test result directory: %s", err.Error())
		return
	}
	testResultFile = filepath.Join(cwd, "..", "..", "e2e-test.out")
}

func UpdateTestResults(data string) {
	fileLock.Lock()
	defer fileLock.Unlock()

	fpointer, err := os.OpenFile(testResultFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer fpointer.Close()
	if _, err = fpointer.WriteString(data + "\n"); err != nil {
		panic(err)
	}
}

func TestConfigMap(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "ConfigMap Suite")
}

var _ = ginkgo.BeforeSuite(func() {
	kubeConfigPath := os.Getenv("KUBECONFIG")
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	kubeClient, err = kubernetes.NewForConfig(config)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
})

var _ = ginkgo.Describe("Enable EIT, but keep worker pool empty", func() {
	const (
		configMapName     = "addon-vpc-file-csi-driver-configmap"
		validationMapName = "file-csi-driver-status"
	)

	ginkgo.It("Enable EIT, but keep worker pool empty", func() {
		ctx := context.Background()

		// Retrieve the existing ConfigMap
		configMap, err := kubeClient.CoreV1().ConfigMaps(namespace).Get(ctx, configMapName, metav1.GetOptions{})
		gomega.Expect(err).ToNot(gomega.HaveOccurred())

		// Update a key in the ConfigMap
		configMap.Data["ENABLE_EIT"] = "true"
		configMap.Data["EIT_ENABLED_WORKER_POOLS"] = ""
		_, err = kubeClient.CoreV1().ConfigMaps(namespace).Update(ctx, configMap, metav1.UpdateOptions{})
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
		if err != nil {
			UpdateTestResults("Enable EIT on all worker pools: FAIL")
			return
		}

		/*
			hostnames := make(map[string][]string)
			workerNodeYaml, err := yaml.Marshal(hostnames)
			if err != nil {
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				UpdateTestResults("Enable EIT on all worker pools: FAIL")
			}*/

		/*
			resultAsExpected1 := gomega.Eventually(func() (map[string]string, error) {
				validationConfigMap, err := kubeClient.CoreV1().ConfigMaps(namespace).Get(ctx, validationMapName, metav1.GetOptions{})
				if err != nil {
					return nil, err
				}
				return validationConfigMap.Data, nil
			}, 5*time.Minute, 60*time.Second).Should(gomega.HaveKeyWithValue("EIT_ENABLED_WORKER_NODES", string(workerNodeYaml)))
		*/

		resultAsExpected := gomega.Eventually(func() (map[string]string, error) {
			validationConfigMap, err := kubeClient.CoreV1().ConfigMaps(namespace).Get(ctx, validationMapName, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}
			return validationConfigMap.Data, nil
		}, 5*time.Minute, 60*time.Second).Should(gomega.HaveKeyWithValue("EIT_ENABLED_WORKER_NODES", ""))

		if resultAsExpected {
			UpdateTestResults("Enable EIT, but keep worker pool empty: PASS")
			return
		}
		UpdateTestResults("Enable EIT, but keep worker pool empty: FAIL")
		// Give some time for the changes to propagate
	})

	ginkgo.It("Enable EIT on only one worker pool", func() {
		ctx := context.Background()

		// Retrieve the existing ConfigMap
		configMap, err := kubeClient.CoreV1().ConfigMaps(namespace).Get(ctx, configMapName, metav1.GetOptions{})
		gomega.Expect(err).ToNot(gomega.HaveOccurred())

		workerpools := []string{"default"}
		// Update a key in the ConfigMap
		configMap.Data["ENABLE_EIT"] = "true"
		configMap.Data["EIT_ENABLED_WORKER_POOLS"] = "default"
		_, err = kubeClient.CoreV1().ConfigMaps(namespace).Update(ctx, configMap, metav1.UpdateOptions{})
		gomega.Expect(err).ToNot(gomega.HaveOccurred())

		workerPoolReq, err := labels.NewRequirement(workerPoolLabelKey, selection.In, workerpools)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
		selector := labels.NewSelector()
		selector = selector.Add(*workerPoolReq)
		nodeList, err := kubeClient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{LabelSelector: selector.String()})
		gomega.Expect(err).ToNot(gomega.HaveOccurred())

		hostnames := make(map[string][]string)
		for _, node := range nodeList.Items {
			hosts, ok := hostnames[node.Labels[workerPoolLabelKey]]
			if ok {
				hostnames[node.Labels[workerPoolLabelKey]] = append(hosts, node.Name)
				continue
			}
			hostnames[node.Labels[workerPoolLabelKey]] = []string{node.Name}
		}

		workerNodeYaml, err := yaml.Marshal(hostnames)
		if err != nil {
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
		}
		// Give some time for the changes to propagate
		resultAsExpected := gomega.Eventually(func() (map[string]string, error) {
			validationConfigMap, err := kubeClient.CoreV1().ConfigMaps(namespace).Get(ctx, validationMapName, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}
			return validationConfigMap.Data, nil
		}, 5*time.Minute, 60*time.Second).Should(gomega.HaveKeyWithValue("EIT_ENABLED_WORKER_NODES", string(workerNodeYaml)))
		if resultAsExpected {
			UpdateTestResults("Enable EIT on only one worker pool: PASS")
			return
		}
		UpdateTestResults("Enable EIT on only one worker pool: FAIL")
	})

	ginkgo.It("Enable EIT on multiple worker pools", func() {
		ctx := context.Background()

		// Retrieve the existing ConfigMap
		configMap, err := kubeClient.CoreV1().ConfigMaps(namespace).Get(ctx, configMapName, metav1.GetOptions{})
		gomega.Expect(err).ToNot(gomega.HaveOccurred())

		workerpools := []string{"default", "worker-pool1"}
		// Update a key in the ConfigMap
		configMap.Data["ENABLE_EIT"] = "true"
		configMap.Data["EIT_ENABLED_WORKER_POOLS"] = "default,worker-pool1"
		_, err = kubeClient.CoreV1().ConfigMaps(namespace).Update(ctx, configMap, metav1.UpdateOptions{})
		gomega.Expect(err).ToNot(gomega.HaveOccurred())

		workerPoolReq, err := labels.NewRequirement(workerPoolLabelKey, selection.In, workerpools)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
		selector := labels.NewSelector()
		selector = selector.Add(*workerPoolReq)
		nodeList, err := kubeClient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{LabelSelector: selector.String()})
		gomega.Expect(err).ToNot(gomega.HaveOccurred())

		hostnames := make(map[string][]string)
		for _, node := range nodeList.Items {
			hosts, ok := hostnames[node.Labels[workerPoolLabelKey]]
			if ok {
				hostnames[node.Labels[workerPoolLabelKey]] = append(hosts, node.Name)
				continue
			}
			hostnames[node.Labels[workerPoolLabelKey]] = []string{node.Name}
		}

		workerNodeYaml, err := yaml.Marshal(hostnames)
		if err != nil {
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
		}

		// Give some time for the changes to propagate
		resultAsExpected := gomega.Eventually(func() (map[string]string, error) {
			validationConfigMap, err := kubeClient.CoreV1().ConfigMaps(namespace).Get(ctx, validationMapName, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}
			return validationConfigMap.Data, nil
		}, 5*time.Minute, 60*time.Second).Should(gomega.HaveKeyWithValue("EIT_ENABLED_WORKER_NODES", string(workerNodeYaml)))
		if resultAsExpected {
			UpdateTestResults("Enable EIT on multiple worker pools: PASS")
			return
		}
		UpdateTestResults("Enable EIT on multiple worker pools: FAIL")
	})

	ginkgo.It("Disable EIT on all worker pools", func() {
		ctx := context.Background()

		// Retrieve the existing ConfigMap
		configMap, err := kubeClient.CoreV1().ConfigMaps(namespace).Get(ctx, configMapName, metav1.GetOptions{})
		gomega.Expect(err).ToNot(gomega.HaveOccurred())

		// Update a key in the ConfigMap
		configMap.Data["ENABLE_EIT"] = "false"
		_, err = kubeClient.CoreV1().ConfigMaps(namespace).Update(ctx, configMap, metav1.UpdateOptions{})
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
		hostnames := make(map[string][]string)
		workerNodeYaml, err := yaml.Marshal(hostnames)
		if err != nil {
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
		}
		// Give some time for the changes to propagate
		resultAsExpected := gomega.Eventually(func() (map[string]string, error) {
			validationConfigMap, err := kubeClient.CoreV1().ConfigMaps(namespace).Get(ctx, validationMapName, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}
			return validationConfigMap.Data, nil
		}, 5*time.Minute, 60*time.Second).Should(gomega.HaveKeyWithValue("EIT_ENABLED_WORKER_NODES", string(workerNodeYaml)))
		if resultAsExpected {
			UpdateTestResults("Disable EIT on all worker pools: PASS")
			return
		}
		UpdateTestResults("Disable EIT on all worker pools: FAIL")
	})

	ginkgo.It("Enable EIT on one worker pool, check the updated worker pool list, update with one more worker pool and verify", func() {
		ctx := context.Background()

		// Retrieve the existing ConfigMap
		configMap, err := kubeClient.CoreV1().ConfigMaps(namespace).Get(ctx, configMapName, metav1.GetOptions{})
		gomega.Expect(err).ToNot(gomega.HaveOccurred())

		workerpools := []string{"default"}
		// Update a key in the ConfigMap
		configMap.Data["ENABLE_EIT"] = "true"
		configMap.Data["EIT_ENABLED_WORKER_POOLS"] = "default"
		_, err = kubeClient.CoreV1().ConfigMaps(namespace).Update(ctx, configMap, metav1.UpdateOptions{})
		gomega.Expect(err).ToNot(gomega.HaveOccurred())

		workerPoolReq, err := labels.NewRequirement(workerPoolLabelKey, selection.In, workerpools)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
		selector := labels.NewSelector()
		selector = selector.Add(*workerPoolReq)
		nodeList, err := kubeClient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{LabelSelector: selector.String()})
		gomega.Expect(err).ToNot(gomega.HaveOccurred())

		hostnames := make(map[string][]string)
		for _, node := range nodeList.Items {
			hosts, ok := hostnames[node.Labels[workerPoolLabelKey]]
			if ok {
				hostnames[node.Labels[workerPoolLabelKey]] = append(hosts, node.Name)
				continue
			}
			hostnames[node.Labels[workerPoolLabelKey]] = []string{node.Name}
		}

		workerNodeYaml, err := yaml.Marshal(hostnames)
		if err != nil {
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
		}

		// Give some time for the changes to propagate
		resultAsExpected := gomega.Eventually(func() (map[string]string, error) {
			validationConfigMap, err := kubeClient.CoreV1().ConfigMaps(namespace).Get(ctx, validationMapName, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}
			return validationConfigMap.Data, nil
		}, 5*time.Minute, 60*time.Second).Should(gomega.HaveKeyWithValue("EIT_ENABLED_WORKER_NODES", string(workerNodeYaml)))
		if !resultAsExpected {
			UpdateTestResults("Enable EIT on one worker pool, check the updated worker pool list, update with one more worker pool and verify: FAIL")
			return
		}

		// Retrieve the existing ConfigMap
		configMap, err = kubeClient.CoreV1().ConfigMaps(namespace).Get(ctx, configMapName, metav1.GetOptions{})
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
		// Adding worker pool
		workerpools = []string{"default", "worker-pool1"}
		configMap.Data["EIT_ENABLED_WORKER_POOLS"] = "default,worker-pool1"
		_, err = kubeClient.CoreV1().ConfigMaps(namespace).Update(ctx, configMap, metav1.UpdateOptions{})
		gomega.Expect(err).ToNot(gomega.HaveOccurred())

		workerPoolReq, err = labels.NewRequirement(workerPoolLabelKey, selection.In, workerpools)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
		selector = labels.NewSelector()
		selector = selector.Add(*workerPoolReq)
		nodeList, err = kubeClient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{LabelSelector: selector.String()})
		gomega.Expect(err).ToNot(gomega.HaveOccurred())

		hostnames = make(map[string][]string)
		for _, node := range nodeList.Items {
			hosts, ok := hostnames[node.Labels[workerPoolLabelKey]]
			if ok {
				hostnames[node.Labels[workerPoolLabelKey]] = append(hosts, node.Name)
				continue
			}
			hostnames[node.Labels[workerPoolLabelKey]] = []string{node.Name}
		}

		workerNodeYaml, err = yaml.Marshal(hostnames)
		if err != nil {
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
		}

		// Give some time for the changes to propagate
		resultAsExpected = gomega.Eventually(func() (map[string]string, error) {
			validationConfigMap, err := kubeClient.CoreV1().ConfigMaps(namespace).Get(ctx, validationMapName, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}
			return validationConfigMap.Data, nil
		}, 5*time.Minute, 60*time.Second).Should(gomega.HaveKeyWithValue("EIT_ENABLED_WORKER_NODES", string(workerNodeYaml)))
		if !resultAsExpected {
			UpdateTestResults("Enable EIT on one worker pool, check the updated worker pool list, update with one more worker pool and verify: FAIL")
			return
		}
		UpdateTestResults("Enable EIT on one worker pool, check the updated worker pool list, update with one more worker pool and verify: PASS")
	})

	ginkgo.It("Enable EIT on a non existing worker pool", func() {
		ctx := context.Background()

		// Retrieve the existing ConfigMap
		configMap, err := kubeClient.CoreV1().ConfigMaps(namespace).Get(ctx, configMapName, metav1.GetOptions{})
		gomega.Expect(err).ToNot(gomega.HaveOccurred())

		// Update a key in the ConfigMap
		configMap.Data["ENABLE_EIT"] = "true"
		configMap.Data["EIT_ENABLED_WORKER_POOLS"] = "non-exist"
		_, err = kubeClient.CoreV1().ConfigMaps(namespace).Update(ctx, configMap, metav1.UpdateOptions{})
		gomega.Expect(err).ToNot(gomega.HaveOccurred())

		hostnames := make(map[string][]string)
		workerNodeYaml, err := yaml.Marshal(hostnames)
		if err != nil {
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
		}

		resultAsExpected := gomega.Eventually(func() (map[string]string, error) {
			validationConfigMap, err := kubeClient.CoreV1().ConfigMaps(namespace).Get(ctx, validationMapName, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}
			return validationConfigMap.Data, nil
		}, 5*time.Minute, 60*time.Second).Should(gomega.HaveKeyWithValue("EIT_ENABLED_WORKER_NODES", string(workerNodeYaml)))
		if !resultAsExpected {
			UpdateTestResults("Enable EIT on a non existing worker pool: FAIL")
			return
		}
		UpdateTestResults("Enable EIT on a non existing worker pool: PASS")
	})

	ginkgo.It("Enable EIT on multiple worker pools, remove one worker pool and verify", func() {
		ctx := context.Background()

		// Retrieve the existing ConfigMap
		configMap, err := kubeClient.CoreV1().ConfigMaps(namespace).Get(ctx, configMapName, metav1.GetOptions{})
		gomega.Expect(err).ToNot(gomega.HaveOccurred())

		workerpools := []string{"default", "worker-pool1"}
		// Update a key in the ConfigMap
		configMap.Data["ENABLE_EIT"] = "true"
		configMap.Data["EIT_ENABLED_WORKER_POOLS"] = "default,worker-pool1"
		_, err = kubeClient.CoreV1().ConfigMaps(namespace).Update(ctx, configMap, metav1.UpdateOptions{})
		gomega.Expect(err).ToNot(gomega.HaveOccurred())

		workerPoolReq, err := labels.NewRequirement(workerPoolLabelKey, selection.In, workerpools)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
		selector := labels.NewSelector()
		selector = selector.Add(*workerPoolReq)
		nodeList, err := kubeClient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{LabelSelector: selector.String()})
		gomega.Expect(err).ToNot(gomega.HaveOccurred())

		hostnames := make(map[string][]string)
		for _, node := range nodeList.Items {
			hosts, ok := hostnames[node.Labels[workerPoolLabelKey]]
			if ok {
				hostnames[node.Labels[workerPoolLabelKey]] = append(hosts, node.Name)
				continue
			}
			hostnames[node.Labels[workerPoolLabelKey]] = []string{node.Name}
		}

		workerNodeYaml, err := yaml.Marshal(hostnames)
		if err != nil {
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
		}

		resultAsExpected := gomega.Eventually(func() (map[string]string, error) {
			validationConfigMap, err := kubeClient.CoreV1().ConfigMaps(namespace).Get(ctx, validationMapName, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}
			return validationConfigMap.Data, nil
		}, 5*time.Minute, 60*time.Second).Should(gomega.HaveKeyWithValue("EIT_ENABLED_WORKER_NODES", string(workerNodeYaml)))
		if !resultAsExpected {
			UpdateTestResults("Enable EIT on multiple worker pools, remove one worker pool and verify: FAIL")
			return
		}

		// Retrieve the existing ConfigMap
		configMap, err = kubeClient.CoreV1().ConfigMaps(namespace).Get(ctx, configMapName, metav1.GetOptions{})
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
		// Removing one worker pool
		workerpools = []string{"default"}
		configMap.Data["EIT_ENABLED_WORKER_POOLS"] = "default"
		_, err = kubeClient.CoreV1().ConfigMaps(namespace).Update(ctx, configMap, metav1.UpdateOptions{})
		gomega.Expect(err).ToNot(gomega.HaveOccurred())

		workerPoolReq, err = labels.NewRequirement(workerPoolLabelKey, selection.In, workerpools)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
		selector = labels.NewSelector()
		selector = selector.Add(*workerPoolReq)
		nodeList, err = kubeClient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{LabelSelector: selector.String()})
		gomega.Expect(err).ToNot(gomega.HaveOccurred())

		hostnames = make(map[string][]string)
		for _, node := range nodeList.Items {
			hosts, ok := hostnames[node.Labels[workerPoolLabelKey]]
			if ok {
				hostnames[node.Labels[workerPoolLabelKey]] = append(hosts, node.Name)
				continue
			}
			hostnames[node.Labels[workerPoolLabelKey]] = []string{node.Name}
		}

		workerNodeYaml, err = yaml.Marshal(hostnames)
		if err != nil {
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
		}

		resultAsExpected = gomega.Eventually(func() (map[string]string, error) {
			validationConfigMap, err := kubeClient.CoreV1().ConfigMaps(namespace).Get(ctx, validationMapName, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}
			return validationConfigMap.Data, nil
		}, 5*time.Minute, 60*time.Second).Should(gomega.HaveKeyWithValue("EIT_ENABLED_WORKER_NODES", string(workerNodeYaml)))
		if !resultAsExpected {
			UpdateTestResults("Enable EIT on multiple worker pools, remove one worker pool and verify: FAIL")
			return
		}
		UpdateTestResults("Enable EIT on multiple worker pools, remove one worker pool and verify: PASS")
	})

})
