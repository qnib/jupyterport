package qniblib

import "strings"

type DockerImage struct {
	Name 	string
}
func (d *DockerImage) String() string {
	return d.Name
}

type DockerImages struct {
	Images []DockerImage
}

func (di *DockerImages) GetImages() []DockerImage {
	return di.Images
}

func (di *DockerImages) String() string {
	res := []string{}
	for _, i := range di.Images {
		res = append(res, i.String())
	}
	return strings.Join(res, ",")
}