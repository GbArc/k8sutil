package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	//"github.com/zclconf/go-cty/cty/json"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

type K8SHandler struct {
	k8sset    *kubernetes.Clientset
	config    *rest.Config
	namespace string
}

type DItem struct {
	Id    string   `json: "id"`
	Names []string `json: "names"`
}

func (k *K8SHandler) init() {

	var k8sConfig string
	flag.StringVar(&k8sConfig, "kubeconfig", filepath.Join(os.Getenv("HOME"), ".kube", "config"), "path to kubeconfig")
	flag.StringVar(&k8sConfig, "k", filepath.Join(os.Getenv("HOME"), ".kube", "config"), "path to kubeconfig")
	flag.StringVar(&k.namespace, "namespace", "default", "namespace")
	flag.StringVar(&k.namespace, "n", "default", "namespace")

	flag.Parse()
	var err error
	k.config, err = clientcmd.BuildConfigFromFlags("", k8sConfig)
	if err != nil {
		panic(err)
	}

	k.k8sset, err = kubernetes.NewForConfig(k.config)
	if err != nil {
		panic(err)
	}
	fmt.Println("Using namespace ", k.namespace)
}

func (k *K8SHandler) SetImage(deploymentName string, imageId string) {
	deployments := k.k8sset.AppsV1().Deployments(k.namespace)
	d, err := deployments.Get(context.TODO(), deploymentName, metav1.GetOptions{})
	if err == nil && len(d.Spec.Template.Spec.Containers) > 0 {
		fmt.Println("found image: " + d.Spec.Template.Spec.Containers[0].Image + " updating to: " + imageId)
		d.Spec.Template.Spec.Containers[0].Image = imageId
		deployments.Update(context.TODO(), d, metav1.UpdateOptions{})
	} else {
		fmt.Println(err)
	}
}

func (k *K8SHandler) SetCommand(deploymentName string, command []string, args []string) {
	deployments := k.k8sset.AppsV1().Deployments(k.namespace)
	d, err := deployments.Get(context.TODO(), deploymentName, metav1.GetOptions{})
	if err == nil && len(d.Spec.Template.Spec.Containers) > 0 {
		d.Spec.Template.Spec.Containers[0].Args = args
		d.Spec.Template.Spec.Containers[0].Command = command
		livenessProbe := d.Spec.Template.Spec.Containers[0].LivenessProbe
		if livenessProbe != nil {
			fmt.Println(livenessProbe.String())
		}
		readinessProbe := d.Spec.Template.Spec.Containers[0].ReadinessProbe
		if readinessProbe != nil {
			fmt.Println(readinessProbe.String())
		}
		d.Spec.Template.Spec.Containers[0].LivenessProbe = nil
		d.Spec.Template.Spec.Containers[0].ReadinessProbe = nil

		deployments.Update(context.TODO(), d, metav1.UpdateOptions{})
	} else {
		fmt.Println(err)
	}
}

func (k *K8SHandler) GetImage(deploymentName string) string {
	var image string = ""
	deployments := k.k8sset.AppsV1().Deployments(k.namespace)
	d, err := deployments.Get(context.TODO(), deploymentName, metav1.GetOptions{})
	if err == nil && len(d.Spec.Template.Spec.Containers) > 0 {
		image = d.Spec.Template.Spec.Containers[0].Image
		fmt.Println(deploymentName + " : " + image)
	}
	return image
}

func (k *K8SHandler) GetDeployments() {
	deployments := k.k8sset.AppsV1().Deployments(k.namespace)

	dlist, err := deployments.List(context.TODO(), metav1.ListOptions{})
	if err == nil {
		for _, d := range dlist.Items {
			var image string = "none"
			if len(d.Spec.Template.Spec.Containers) > 0 {
				image = d.Spec.Template.Spec.Containers[0].Image
				for _, cmd := range d.Spec.Template.Spec.Containers[0].Command {
					fmt.Println("command: ", cmd)
				}
				if d.Spec.Template.Spec.Containers[0].Args != nil {
					for _, arg := range d.Spec.Template.Spec.Containers[0].Args {
						fmt.Println("arg: ", arg)
					}
				}
				fmt.Println("\033[32;1m" + d.Name + "\033[0m  current image: \033[33;1m" + image + "\033[0m")
			}
		}
	}
}

func (k *K8SHandler) GetPods() {
	fmt.Println("namespace[", k.namespace, "]")

	pods, err := k.k8sset.CoreV1().Pods(k.namespace).List(context.TODO(), metav1.ListOptions{})
	if err == nil {
		fmt.Println("Pods in namespace " + k.namespace + " : " + string(len(pods.Items)))
		for _, pod := range pods.Items {
			fmt.Println(pod.Name)
			for _, container := range pod.Spec.Containers {
				fmt.Println("\t" + container.Name + " image: " + container.Image)
			}
		}
	} else {
		fmt.Println(err)
	}
}

func (k *K8SHandler) Exec(podName string, cmd []string) (string, string, error) {
	outBuffer := &bytes.Buffer{}
	errBuffer := &bytes.Buffer{}
	client := k.k8sset.CoreV1().RESTClient()
	req := client.Post().Resource("pods").Name(podName).Namespace(k.namespace).SubResource("exec")
	options := &v1.PodExecOptions{
		Command: cmd,
		Stdin:   false,
		Stdout:  true,
		Stderr:  true,
		TTY:     true,
	}
	req.VersionedParams(options, scheme.ParameterCodec)
	exec, err := remotecommand.NewSPDYExecutor(k.config, "POST", req.URL())
	if err == nil {
		err = exec.Stream(remotecommand.StreamOptions{Stdin: nil, Stdout: outBuffer, Stderr: errBuffer})
		fmt.Println("execing!")
	}
	return outBuffer.String(), errBuffer.String(), err
}
