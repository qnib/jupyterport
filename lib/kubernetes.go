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
	cset *kubernetes.Clientset
}

func NewKubernetesSpawner() KubernetesSpawner {
	return KubernetesSpawner{}
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
		url := ""
		if v, ok := pod.Labels["port"]; ok {
			url = fmt.Sprintf("http://127.0.0.1:%s", v)
		}
		if v, ok := pod.Labels["user"]; ok {
			url = fmt.Sprintf("%suser/%s", url, v)
		}
		if v, ok := pod.Labels["name"]; ok {
			url = fmt.Sprintf("%s/%s", url, v)
		}
		if v, ok := pod.Labels["token"]; ok {
			url = fmt.Sprintf("%s/?token=%s", url, v)
		}

		log.Printf("Found notebook '%s': %s", pod.Name, url)
		nbs[pod.Name] = NewNotebook(string(pod.GetUID()), pod.GetName(), user, url, token)
	}
	return nbs, err
}

// SpawnNotebooks create a notebook
func (s *KubernetesSpawner) SpawnNotebooks(user, name, port, image, token string) (err error) {
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
	return
}

func getDeployment(user, name, port, image, token string) (depl *appsv1.Deployment, err error) {
	depl = &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s-deloyment", user, name),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":  fmt.Sprintf("%s-%s-deloyment", user, name),
					"app-type": "jupyter-notebook",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":  fmt.Sprintf("%s-%s-deloyment", user, name),
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
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									// TODO: Hard wired interal port
									ContainerPort: 8888,
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
			Name: fmt.Sprintf("%s-%s-service", user, name),
		},
		Spec: apiv1.ServiceSpec{
			Type:  apiv1.ServiceType("NodePort"),
			Selector: map[string]string{"app": fmt.Sprintf("%s-%s-deloyment", user, name)},
			Ports: []apiv1.ServicePort{{Protocol: "TCP", Port: 8888, TargetPort: intstr.FromInt(8888), NodePort: int32(iPort)}},
		},
	}
	return
}

