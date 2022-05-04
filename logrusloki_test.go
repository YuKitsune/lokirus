package logrusloki_test

import (
	"fmt"
	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
	"log"
	"os"
	"testing"
	"time"
)

const host string = "localhost"
const port string = "3100"

func TestMain(m *testing.M) {

	image := "grafana/loki"
	version := "2.5.0"

	// Set up the loki container
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	// pull mongodb docker image
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: image,
		Tag:        version,

		// Expose the Loki port
		PortBindings: map[docker.Port][]docker.PortBinding{
			docker.Port(fmt.Sprintf("%s/tcp", port)): {{HostIP: "", HostPort: port}},
		},
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})

	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	// Kill the container after a while
	err = resource.Expire(uint((5 * time.Minute) / time.Second))
	if err != nil {
		log.Fatalf("Could not set resource expiry: %s", err)
	}

	code := m.Run()

	// When we're done, kill and remove the container
	if err = pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

// Todo

func TestLokiHook_Fires(t *testing.T) {

}

func TestLokiHook_AddsCustomLabels(t *testing.T) {

}

func TestLokiHook_ReMapsLevel(t *testing.T) {

}

func TestLokiHook_UsesCustomHttpClient(t *testing.T) {

}

func TestLokiHook_UsesCustomFormatter(t *testing.T) {

}
