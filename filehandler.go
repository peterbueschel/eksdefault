//Copyright 2018 Peter Büschel
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.package awsdefault

package eksdefault

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sort"

	"github.com/peterbueschel/awsdefault"
	"gopkg.in/yaml.v2"
)

var (
	NoProfilSet = errors.New(
		"no AWS profile configured for this context. " +
			"Define it by adding 'aws-profile: <profile>' to the context inside the kube config",
	)
)

type (
	// Profile stored in the AWS shared credentials file consisting of an
	// AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY
	KubeContext struct {
		Name       string `yaml:"name"`
		AWSprofile string `yaml:"aws-profile"`
		*Context   `yaml:"context"`
	}
	Context struct {
		Cluster   string `yaml:"cluster"`
		User      string `yaml:"user"`
		Namespace string `yaml:"namespace"`
	}

	// CredentialsFile stores the content and path of the AWS credentials file
	KubeConfig struct {
		ApiVersion     string        `yaml:"apiVersion"`
		Kind           string        `yaml:"kind"`
		Preferences    interface{}   `yaml:"preferences"`
		Contexts       []KubeContext `yaml:"contexts"`
		CurrentContext string        `yaml:"current-context"`
		Clusters       []struct {
			Name    string      `yaml:"name"`
			Cluster interface{} `yaml:"cluster"`
		} `yaml:"clusters"`
		Users []struct {
			Name string      `yaml:"name"`
			User interface{} `yaml:"user"`
		} `yaml:"users"`
		Path string `yaml:"-"`
	}
)

func inList(item string, list []string) bool {
	for _, i := range list {
		if item == i {
			return true
		}
	}
	return false
}

// GetCredentialsFile reads the AWS credentials file either from the HOME directory or
// from a path given by the environment variable AWS_SHARED_CREDENTIALS_FILE
func GetConfigFile() (*KubeConfig, error) {
	home := func() string {
		if runtime.GOOS == "windows" {
			return os.Getenv("USERPROFILE")
		}
		return os.Getenv("HOME")
	}
	path := filepath.Join(home(), ".kube", "config")
	if p := os.Getenv("KUBECONFIG"); len(p) > 0 {
		path = p
	}
	k := &KubeConfig{Path: path}
	f, err := ioutil.ReadFile(path)
	if err != nil {
		return k, err
	}
	err = yaml.Unmarshal(f, &k)
	if err != nil {
		return k, err
	}
	dupl := make(map[string]int)
	for _, ctx := range k.Contexts {
		dupl[ctx.Name]++
	}
	for d, v := range dupl {
		if v > 1 {
			return &KubeConfig{}, fmt.Errorf("[KUBECONFIG] found duplicated context name: %s", d)
		}
	}
	sort.Slice(k.Contexts, func(i, j int) bool {
		return k.Contexts[i].Name < k.Contexts[j].Name
	})
	return k, nil
}

// GetProfilesNames returns a sorted list of all available profiles inside the AWS credentials file.
func (k *KubeConfig) GetContextNames() (names []string) {
	for _, p := range k.Contexts {
		names = append(names, p.Name)
	}
	sort.Strings(names)
	return
}

// GetProfileBy returns the profile by a given name
func (k *KubeConfig) GetContextBy(name string) (*KubeContext, int, error) {
	for idx, p := range k.Contexts {
		if p.Name == name {
			if p.Context == nil {
				p.Context = &Context{}
			}
			return &p, idx, nil
		}
	}
	return &KubeContext{}, -1, fmt.Errorf("[GETCONTEXT] cannot find context named '%s'", name)
}

func (k *KubeConfig) SaveContexts() error {
	sort.Slice(k.Contexts, func(i, j int) bool {
		return k.Contexts[i].Name < k.Contexts[j].Name
	})
	cnf, err := yaml.Marshal(&k)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(k.Path, cnf, 0644)
}

func (k *KubeConfig) SetContextTo(contextName string) error {
	ctx, _, err := k.GetContextBy(contextName)
	if err != nil {
		return err
	}
	if len(ctx.AWSprofile) < 1 {
		return NoProfilSet
	}
	awsfile, err := awsdefault.GetCredentialsFile()
	if err != nil {
		return err
	}
	err = awsfile.SetDefaultTo(ctx.AWSprofile)
	if err != nil {
		return err
	}
	k.CurrentContext = contextName
	return k.SaveContexts()
}

// AddProfileTo
func (k *KubeConfig) AddProfileTo(contextName, profileName string) error {
	ctx, idx, err := k.GetContextBy(contextName)
	if err != nil {
		return err
	}
	// check if profile exsists in AWS Credentials file
	awsfile, err := awsdefault.GetCredentialsFile()
	if err != nil {
		return err
	}
	if inList(profileName, awsfile.GetProfilesNames()) {
		ctx.AWSprofile = profileName
		k.Contexts[idx] = *ctx
		return k.SaveContexts()
	}
	return fmt.Errorf("given profile name '%s' does not exists in '%s'", profileName, awsfile.Path)
}

// AddProfileTo
func (k *KubeConfig) AddNamespaceTo(contextName, namespace string) error {
	ctx, idx, err := k.GetContextBy(contextName)
	if err != nil {
		return err
	}
	ctx.Context.Namespace = namespace
	k.Contexts[idx] = *ctx
	return k.SaveContexts()
}

// UnSetDefault deletes the default section inside the AWS credentials file.
func (k *KubeConfig) UnSetDefault() error {
	k.CurrentContext = ""
	return k.SaveContexts()
}
