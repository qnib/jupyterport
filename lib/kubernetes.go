package qniblib // import "github.com/qnib/jupyterport/lib"

import (
	"fmt"
	"github.com/codegangsta/cli"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
	"net/http"
	"strconv"
)

type KubernetesSpawner struct {
	Type string
	cset *kubernetes.Clientset
	namespace	string
}

func NewKubernetesSpawner() KubernetesSpawner {
	return KubernetesSpawner{Type: "kubernetes"}
}

func (s *KubernetesSpawner) Init(ctx *cli.Context) (err error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}
	s.namespace = ctx.String("k8s-namespace")
	log.Printf("Kubernetes Namespace: %s", s.namespace)
	insec := true
	if insec {
		config.Insecure = insec
		config.CAFile = ""
		config.CertFile = ""
	}

	s.cset, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	log.Printf("config.Host:%s", config.Host)
	return
}

// ListNotebooks returns the notebooks for a given user
func (s *KubernetesSpawner) ListNotebooks(user, extAddr string) (map[string]Notebook, error) {
	nbs := make(map[string]Notebook)
	var err error
	// TODO: add selector for given user
	pods, err := s.cset.CoreV1().Pods(s.namespace).List(metav1.ListOptions{LabelSelector: "app-type=jupyter-notebook"})
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
		eurl := fmt.Sprintf("http://%s:%s", extAddr, port)
		path := fmt.Sprintf("/user/%s/%s", user, name)
		log.Printf("Found notebook '%s': Internal:%s External:%s Path:%s", pod.GetName(), iurl, eurl, path)
		nbs[pod.GetName()] = NewNotebook(string(pod.GetUID()), s.Type, pod.GetName(), user, iurl, eurl, path, token)
	}
	return nbs, err
}

// SpawnNotebooks create a notebook
func (s *KubernetesSpawner) SpawnNotebook(user string, r *http.Request, token, extAddr string) (nb Notebook, err error) {
	cntname := r.FormValue("cntname")
	cntport := r.FormValue("cntport")
	cntimg := r.FormValue("cntimage")
	deploymentsClient := s.cset.AppsV1().Deployments(s.namespace)

	deployment, err := getDeployment(user, r, token)
	if err != nil {
		log.Printf("getDeployment(): %q\n", err.Error())
		return
	}
	log.Printf("Creating deployment: ")
	dplRes, err := deploymentsClient.Create(deployment)
	if err != nil {
		log.Printf("NOK %q\n", err.Error())
		return
	}
	log.Printf("OK %q\n", dplRes.GetObjectMeta().GetName())
	srvClient := s.cset.CoreV1().Services(s.namespace)
	svc, err := getSrv(user, cntname, cntport, cntimg, token)
	log.Printf("Creating service: ")
	svcRes, err := srvClient.Create(svc)
	if err != nil {
		log.Printf("NOK %s\n", err.Error())
		return
	}
	log.Printf("OK %q\n", svcRes.GetObjectMeta().GetName())
	iurl := fmt.Sprintf("http://%s-%s.default.svc.cluster.local:%d", user, cntname, InternalNotebookPort)
	eurl := fmt.Sprintf("http://%s:%d", extAddr, cntport)
	path := fmt.Sprintf("/user/%s/%s", user, cntname)
	log.Printf("Found notebook '%s': Internal:%s External:%s Path:%s", cntname, iurl, eurl, path)
	nb  = NewNotebook(string(svcRes.GetObjectMeta().GetUID()), s.Type, svcRes.GetObjectMeta().GetName(), user, iurl, eurl, path, token)
	return
}

func getDeployment(user string, r *http.Request, token string) (depl *appsv1.Deployment, err error) {
	cntname := r.FormValue("cntname")
	cntport := r.FormValue("cntport")
	cntimg := r.FormValue("cntimage")
	nbimage := r.FormValue("nbimage")
	dataImage := r.FormValue("dataimage")
	gpus, err := strconv.Atoi(r.FormValue("cnt-gpu"))
	if err != nil {
		return
	}
	rcuda, err := strconv.Atoi(r.FormValue("cnt-rcuda"))
	if err != nil {
		return
	}
	log.Printf("Resource -> qnib.org/gpu:%d / qnib.org/rcuda:%d", gpus, rcuda)
	uid := int64(0)
	gid := int64(0)
	depl = &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s", user, cntname),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":  fmt.Sprintf("%s-%s", user, cntname),
					"app-type": "jupyter-notebook",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":  fmt.Sprintf("%s-%s", user, cntname),
						"app-type": "jupyter-notebook",
						"port": cntport,
						"token": token,
						"user": user,
						"name": cntname,
					},
				},
				Spec: apiv1.PodSpec{
					Volumes: []apiv1.Volume{
						{
							Name: "notebooks",
							VolumeSource: apiv1.VolumeSource{},
						},
						{
							Name: "data",
							VolumeSource: apiv1.VolumeSource{},
						},
					},
					Containers: []apiv1.Container{
						{
							Name:  "jupyter",
							Image: cntimg,
							SecurityContext: &apiv1.SecurityContext{RunAsUser: &uid, RunAsGroup: &gid},
							Env: []apiv1.EnvVar{
								{Name: "JUPYTERHUB_ROUTE",Value: fmt.Sprintf("/user/%s/%s", user, cntname)},
								{Name: "JUPYTERHUB_API_TOKEN",Value: token},
							},
							Ports: []apiv1.ContainerPort{
								{
									Name:          "notebook",
									Protocol:      apiv1.ProtocolTCP,
									// TODO: Hard wired interal port
									ContainerPort: InternalNotebookPort,
								},

							},
							WorkingDir: "/notebooks",
							Resources: apiv1.ResourceRequirements{
								Limits: apiv1.ResourceList{
									"qnib.org/gpu": *resource.NewQuantity(int64(gpus), resource.DecimalSI),
									"qnib.org/rcuda": *resource.NewQuantity(int64(rcuda), resource.DecimalSI),
								},
							},
							VolumeMounts: []apiv1.VolumeMount{
								{Name: "notebooks", MountPath: "/notebooks"},
								{Name: "data", MountPath: "/data"},
							},
						},
					},
					InitContainers: []apiv1.Container{
						{
							Name:  "notebooks",
							Image: nbimage,
							VolumeMounts: []apiv1.VolumeMount{{Name: "notebooks", MountPath: "/dst"}},
							Command: []string{"/copy", "/notebooks", "/dst"},
						},
						{
							Name:  "data",
							Image: dataImage,
							VolumeMounts: []apiv1.VolumeMount{{Name: "data", MountPath: "/dst"}},
							Command: []string{"/copy", "/data", "/dst"},
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
			},
		},
	}
	return
}

