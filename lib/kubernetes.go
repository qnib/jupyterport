package qniblib // import "github.com/qnib/jupyterport/lib"

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
	"strconv"
)

type KubernetesSpawner struct {
	Type string
	cset *kubernetes.Clientset
}

func NewKubernetesSpawner() KubernetesSpawner {
	return KubernetesSpawner{Type: "kubernetes"}
}

func (s *KubernetesSpawner) Init() (err error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}
	s.cset, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	log.Printf("config.Host:%s", config.Host)
	return
}

// ListNotebooks returns the notebooks for a given user
func (s *KubernetesSpawner) ListNotebooks(user string) (map[string]Notebook, error) {
	nbs := make(map[string]Notebook)
	var err error
	// Logic
	pods, err := s.cset.CoreV1().Pods("default").List(metav1.ListOptions{LabelSelector: "app-type=jupyter-notebook"})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
	for _, pod := range pods.Items {
		token := ""
		port := ""
		name := ""
		if v, ok := pod.Labels["port"]; ok {
			port = v
		}
		if v, ok := pod.Labels["name"]; ok {
			name = v
		}
		if v, ok := pod.Labels["token"]; ok {
			token = v
		}
		iurl := fmt.Sprintf("http://%s-%s.default.svc.cluster.local:%d", user, name, InternalNotebookPort)
		eurl := fmt.Sprintf("http://%s:%s", baseIP, port)
		path := fmt.Sprintf("/user/%s/%s", user, name)
		log.Printf("Found notebook '%s': Internal:%s External:%s Path:%s", pod.GetName(), iurl, eurl, path)
		nbs[pod.Name] = NewNotebook(string(pod.GetUID()), s.Type, pod.GetName(), user, iurl, eurl, path, token)
	}
	return nbs, err
}

// SpawnNotebooks create a notebook
func (s *KubernetesSpawner) SpawnNotebook(user, name, port, image, token string) (nb Notebook, err error) {
	deploymentsClient := s.cset.AppsV1().Deployments(apiv1.NamespaceDefault)

	deployment, err := getDeployment(user, name, port, image, token)
	if err != nil {
		return
	}
	log.Printf("Creating deployment: ")
	dplRes, err := deploymentsClient.Create(deployment)
	if err != nil {
		log.Printf("NOK %q\n", err.Error())
		return
	}
	log.Printf("OK %q\n", dplRes.GetObjectMeta().GetName())
	srvClient := s.cset.CoreV1().Services(apiv1.NamespaceDefault)
	svc, err := getSrv(user, name, port, image, token)
	log.Printf("Creating service: ")
	svcRes, err := srvClient.Create(svc)
	if err != nil {
		log.Printf("NOK %s\n", err.Error())
		return
	}
	log.Printf("OK %q\n", svcRes.GetObjectMeta().GetName())
	iurl := fmt.Sprintf("http://%s-%s.default.svc.cluster.local:%d", user, name, InternalNotebookPort)
	eurl := fmt.Sprintf("http://%s:%d", baseIP, port)
	path := fmt.Sprintf("/user/%s/%s", user, name)
	log.Printf("Found notebook '%s': Internal:%s External:%s Path:%s", name, iurl, eurl, path)
	nb  = NewNotebook(string(svcRes.GetObjectMeta().GetUID()), s.Type, svcRes.GetObjectMeta().GetName(), user, iurl, eurl, path, token)
	return
}

func getDeployment(user, name, port, image, token string) (depl *appsv1.Deployment, err error) {
	depl = &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s", user, name),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":  fmt.Sprintf("%s-%s", user, name),
					"app-type": "jupyter-notebook",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":  fmt.Sprintf("%s-%s", user, name),
						"app-type": "jupyter-notebook",
						"port": port,
						"token": token,
						"user": user,
						"name": name,
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "notebook",
							Image: image,
							Env: []apiv1.EnvVar{
								{Name: "JUPYTERPORT_ROUTE",Value: fmt.Sprintf("/user/%s/%s", user, name)},
								{Name: "JUPYTERHUB_API_TOKEN",Value: token},
							},
							Ports: []apiv1.ContainerPort{
								{
									Name:          "notebook",
									Protocol:      apiv1.ProtocolTCP,
									// TODO: Hard wired interal port
									ContainerPort: InternalNotebookPort,
								},
								{
									Name:          "notebook-dev",
									Protocol:      apiv1.ProtocolTCP,
									// TODO: Test port in case a second notebook is started by hand to troubleshoot
									ContainerPort: InternalNotebookPort+1,
								},

							},
						},
					},
				},
			},
		},
	}
	return
}

func getSrv(user, name, port, image, token string) (svc *apiv1.Service, err error) {
	iPort, err := strconv.Atoi(port)
	svc = &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s", user, name),
		},
		Spec: apiv1.ServiceSpec{
			Type:  apiv1.ServiceType("NodePort"),
			Selector: map[string]string{"app": fmt.Sprintf("%s-%s", user, name)},
			Ports: []apiv1.ServicePort{
				{Name: "notebook", Protocol: "TCP", Port: InternalNotebookPort, TargetPort: intstr.FromInt(InternalNotebookPort), NodePort: int32(iPort)},
				{Name: "notebook-dev", Protocol: "TCP", Port: InternalNotebookPort+1, TargetPort: intstr.FromInt(InternalNotebookPort+1), NodePort: int32(iPort+1)}},
		},
	}
	return
}

