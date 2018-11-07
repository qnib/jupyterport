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
	"path"
	"strconv"
	"strings"
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
func (s *KubernetesSpawner) ListNotebooks(user User) (map[string]Notebook, error) {
	log.Printf("kube.ListNotebooks(%s)", user.Name)
	nbs := make(map[string]Notebook)
	var err error
	// TODO: add selector for given user
	pods, err := s.cset.CoreV1().Pods(s.namespace).List(metav1.ListOptions{LabelSelector: fmt.Sprintf("user=%s", strings.ToLower(user.Name))})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
	for _, pod := range pods.Items {
		token := ""
		name := ""
		if v, ok := pod.Labels["name"]; ok {
			name = v
		}
		if v, ok := pod.Labels["token"]; ok {
			token = v
		}
		iurl := fmt.Sprintf("http://%s-%s.default.svc.cluster.local:%d", user.Name, name, InternalNotebookPort)
		path := fmt.Sprintf("/user/%s/%s", strings.ToLower(user.Name), name)
		status := "Undefined"
		for _, p := range pod.Status.ContainerStatuses {
			if p.Ready {
				status = "Running"
				break
			}
			if ! p.Ready && p.State.Waiting != nil {
				status = p.State.Waiting.Reason
				break
			}
			if ! p.Ready && p.State.Terminated != nil {
				status = "Terminated"
			}
		}
		log.Printf("Found notebook '%s' (pod:%s): Internal:%s // Path:%s // %s", name, pod.GetName(), iurl, path, status)
		log.Printf(" Status > %v\n", pod.Status)
		nbs[pod.GetName()] = NewNotebook(string(pod.GetUID()), s.Type, name, user, iurl, path, token, status)
	}
	return nbs, err
}

// SpawnNotebooks create a notebook
func (s *KubernetesSpawner) SpawnNotebook(user User, token string, r *http.Request) (nb Notebook, err error) {
	deploymentsClient := s.cset.AppsV1().Deployments(s.namespace)
	nbname := r.FormValue("nbname")
	deployment, err := getDeployment(user, token, r)
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
	svc, err := getSrv(user, nbname)
	log.Printf("Creating service: ")
	svcRes, err := srvClient.Create(svc)
	if err != nil {
		log.Printf("NOK %s\n", err.Error())
		return
	}
	log.Printf("OK %q\n", svcRes.GetObjectMeta().GetName())
	iurl := fmt.Sprintf("http://%s-%s.default.svc.cluster.local:%d", strings.ToLower(user.Name), nbname, InternalNotebookPort)
	path := fmt.Sprintf("/user/%s/%s", strings.ToLower(user.Name), nbname)
	log.Printf("Found notebook '%s': Internal:%s // Path:%s", nbname, iurl, path)
	nb  = NewNotebook(string(svcRes.GetObjectMeta().GetUID()), s.Type, svcRes.GetObjectMeta().GetName(), user, iurl, path, token, "Pending")
	return
}

//{Name: "JUPYTER_WEBSOCKET_URL", Value: fmt.Sprintf("%s:%s", strings.Replace(extAddr, "http", "ws",1), cntport)},
func getDeployment(user User, token string, r *http.Request) (depl *appsv1.Deployment, err error) {
	nbname := r.FormValue("nbname")
	cntimg := r.FormValue("cntimage")
	nbimage := r.FormValue("nbimage")
	dataImage := r.FormValue("dataimage")
	dataloc := r.FormValue("dataloc")
	wdloc := r.FormValue("wdloc")
	uidStr := user.UID
	gidStr := user.GID
	baseDir := r.FormValue("basedir")
	saveDir := r.FormValue("savedir")
	workDir := r.FormValue("workdir")
	workPath := r.FormValue("workpath")
	if workDir == "" {
		workDir = defaultWorkDir
	}
	gpus, err := strconv.Atoi(r.FormValue("cnt-gpu"))
	if err != nil {
		return
	}
	rcuda, err := strconv.Atoi(r.FormValue("cnt-rcuda"))
	if err != nil {
		return
	}
	log.Printf("Resource -> qnib.org/gpu:%d / qnib.org/rcuda:%d", gpus, rcuda)
	uidInt, err := strconv.Atoi(uidStr)
	if err != nil {
		log.Printf("Failed to convert %s: %s", uidStr, err.Error())
		uidInt = 0
	}
	gidInt, err := strconv.Atoi(gidStr)
	if err != nil {
		log.Printf("Failed to convert %s: %s", gidStr, err.Error())
		gidInt = 0
	}
	uid := int64(uidInt)
	gid := int64(gidInt)
	p := false
	depl = &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s", strings.ToLower(user.Name), nbname),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":  fmt.Sprintf("%s-%s", strings.ToLower(user.Name), nbname),
					"app-type": "jupyter-notebook",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":  fmt.Sprintf("%s-%s", strings.ToLower(user.Name), nbname),
						"app-type": "jupyter-notebook",
						"token": token,
						"user": strings.ToLower(user.Name),
						"name": nbname,
					},
				},
				Spec: apiv1.PodSpec{
					Volumes: []apiv1.Volume{},
					Containers: []apiv1.Container{
						{
							Name:  "jupyter",
							Image: cntimg,
							SecurityContext: &apiv1.SecurityContext{RunAsUser: &uid, RunAsGroup: &gid, Privileged: &p},
							Env: []apiv1.EnvVar{
								{Name: "JUPYTER_BASE_URL",Value: fmt.Sprintf("/user/%s/%s", strings.ToLower(user.Name), nbname)},
								{Name: "JUPYTER_API_TOKEN",Value: token},
								{Name: "JUPYTERPORT_SAVE_PATH", Value: saveDir},
								{Name: "JUPYTERPORT_BASE_DIR", Value: baseDir},
							},
							Ports: []apiv1.ContainerPort{
								{
									Name:          "notebook",
									Protocol:      apiv1.ProtocolTCP,
									// TODO: Hard wired internal port
									ContainerPort: InternalNotebookPort,
								},

							},
							WorkingDir: path.Join(workDir, workPath),
							Resources: apiv1.ResourceRequirements{
								Limits: apiv1.ResourceList{
									"qnib.org/gpu": *resource.NewQuantity(int64(gpus), resource.DecimalSI),
									"qnib.org/rcuda": *resource.NewQuantity(int64(rcuda), resource.DecimalSI),
								},
							},

							VolumeMounts: []apiv1.VolumeMount{
								{Name: "data", MountPath: "/data"},
								{Name: "workdir", MountPath: workDir},
							},
						},
					},
					InitContainers: []apiv1.Container{},
				},
			},
		},
	}
	hpd := apiv1.HostPathDirectory
	///////////////// Data
	// Data Location
	switch dataloc {
	case "local":
		datVol := apiv1.Volume{
			Name:         "data",
			VolumeSource: apiv1.VolumeSource{},
		}
		depl.Spec.Template.Spec.Volumes = append(depl.Spec.Template.Spec.Volumes, datVol)
	case "nfs":
		datVol := apiv1.Volume{
			Name:         "data",
			VolumeSource: apiv1.VolumeSource{HostPath: &apiv1.HostPathVolumeSource{Path: nfsDataPath, Type: &hpd}},
		}
		depl.Spec.Template.Spec.Volumes = append(depl.Spec.Template.Spec.Volumes, datVol)

	}
	// Data Content
	switch {
	case dataImage != "":
		diC := apiv1.Container{
			Name:         "data",
			Image:        dataImage,
			SecurityContext: &apiv1.SecurityContext{RunAsUser: &uid, RunAsGroup: &gid, Privileged: &p},
			VolumeMounts: []apiv1.VolumeMount{{Name: "data", MountPath: "/dst"}},
			Command:      []string{"/copy", "/data", "/dst"},
		}
		log.Printf("Add initContainer '%s' to stage data", dataImage)
		depl.Spec.Template.Spec.InitContainers = append(depl.Spec.Template.Spec.InitContainers, diC)
	}
	///////////////// Workdir
	// Workdir Location
	switch wdloc {
	case "local":
		datVol := apiv1.Volume{
			Name:         "workdir",
			VolumeSource: apiv1.VolumeSource{},
		}
		depl.Spec.Template.Spec.Volumes = append(depl.Spec.Template.Spec.Volumes, datVol)
	case "nfs":
		datVol := apiv1.Volume{
			Name:         "workdir",
			VolumeSource: apiv1.VolumeSource{HostPath: &apiv1.HostPathVolumeSource{Path: nfsWorkdDirPath, Type: &hpd}},
		}
		depl.Spec.Template.Spec.Volumes = append(depl.Spec.Template.Spec.Volumes, datVol)

	}
	// Workdir Content
	switch {
	case nbimage != "":
		nbC := apiv1.Container{
				Name:  "workdir",
				Image: nbimage,
				SecurityContext: &apiv1.SecurityContext{RunAsUser: &uid, RunAsGroup: &gid, Privileged: &p},
				VolumeMounts: []apiv1.VolumeMount{{Name: "workdir", MountPath: "/dst"}},
				Command: []string{"/copy", "/notebooks", path.Join("/dst/", workPath)},
		}
		depl.Spec.Template.Spec.InitContainers = append(depl.Spec.Template.Spec.InitContainers, nbC)
	/*
	case nbimage == "":
		hpd := apiv1.HostPathDirectory
		nbVol := apiv1.Volume{
			Name: "workdir",
			VolumeSource: apiv1.VolumeSource{HostPath: &apiv1.HostPathVolumeSource{Path: workDir, Type: &hpd}},
		}
		depl.Spec.Template.Spec.Volumes = append(depl.Spec.Template.Spec.Volumes, nbVol)
	*/
	default:
		panic("Notebook source is missing")
	}
	return
}

func (s *KubernetesSpawner) DeleteNotebook(user User, nbname string) (err error) {
	deploymentsClient := s.cset.AppsV1().Deployments(s.namespace)
	deplName := fmt.Sprintf("%s-%s", strings.ToLower(user.Name), nbname)
	log.Printf("Delete deployment '%s': ", deplName)
	err = deploymentsClient.Delete(deplName, &metav1.DeleteOptions{})
	if err != nil {
		log.Printf("NOK %s\n", err.Error())
		return
	}
	log.Printf("OK\n")
	srvClient := s.cset.CoreV1().Services(s.namespace)
	log.Printf("Delete service '%s': ", deplName)
	err = srvClient.Delete(deplName, &metav1.DeleteOptions{})
	if err != nil {
		log.Printf("NOK %s\n", err.Error())
		return
	}
	log.Printf("OK\n")
	return
}

func getSrv(user User, name string) (svc *apiv1.Service, err error) {
	svc = &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s", strings.ToLower(user.Name), name),
			Labels: map[string]string{
				"app":  fmt.Sprintf("%s-%s", strings.ToLower(user.Name), name),
				"app-type": "jupyter-notebook",
			},
		},
		Spec: apiv1.ServiceSpec{
			Type:  apiv1.ServiceType("ClusterIP"),
			Selector: map[string]string{"app": fmt.Sprintf("%s-%s", strings.ToLower(user.Name), name)},
			Ports: []apiv1.ServicePort{
				{Name: "notebook", Protocol: "TCP", Port: InternalNotebookPort, TargetPort: intstr.FromInt(InternalNotebookPort)},
			},
		},
	}
	return
}

