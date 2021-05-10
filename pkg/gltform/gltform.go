// (C) Copyright 2021 Hewlett Packard Enterprise Development LP

package gltform

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

const fileExtension = ".gltform"

// Gljwt - the contents of the .gltform file
type Gljwt struct {
	// SpaceName is optional, and is only required for bmaas if we want to create a project
	SpaceName string `yaml:"space_name,omitempty"`
	// ProjectID - the bmaas/Quake project ID
	ProjectID string `yaml:"project_id"`
	// BmaasRestURL - the URL to be used for bmaas, at present it refers to a Quake portal URL
	BmaasRestURL string `yaml:"bmaas_rest_url"`
	// Token - the GL IAM token
	Token string `yaml:"access_token"`
}

// GetGLConfig - reads the .gltform file, note that the .gltform can be in the home directory of the
// user running terraform, or in the directory from which terraform is run
func GetGLConfig() (gljwt *Gljwt, err error) {
	homeDir, _ := os.UserHomeDir()
	workingDir, _ := os.Getwd()
	for _, p := range []string{homeDir, workingDir} {
		gljwt, err = loadGLConfig(p)
		if err == nil {
			break
		}
	}

	return gljwt, err
}

func loadGLConfig(dir string) (*Gljwt, error) {
	f, err := os.Open(filepath.Clean(filepath.Join(dir, fileExtension)))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return parseGLStream(f)
}

func parseGLStream(s io.Reader) (*Gljwt, error) {
	contents, err := ioutil.ReadAll(s)
	if err != nil {
		return nil, err
	}

	q := &Gljwt{}
	err = yaml.Unmarshal(contents, q)

	return q, err
}
