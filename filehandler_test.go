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
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/peterbueschel/awsdefault"
)

var (
	testFileContent = []byte("")
	testFilePath    = "testdata/config"
)

func TestMain(m *testing.M) {
	// setup
	var err error
	testFileContent, err = ioutil.ReadFile("testdata/.kube/config")
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	code := m.Run()
	// teardown
	err = os.Remove(testFilePath)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	if err := os.Setenv("HOME", "testdata"); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	awsfile, err := awsdefault.GetCredentialsFile()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	if err = awsfile.UnSetDefault(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	//
	if err = ioutil.WriteFile("testdata/.kube/config", testFileContent, 0644); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	os.Exit(code)
}

func TestGetConfigFile(t *testing.T) {
	tests := []struct {
		name    string
		want    *KubeConfig
		envVar  string
		envVal  string
		wantErr bool
	}{
		{
			name:    "0positiv - read valid config from HOME",
			envVar:  "HOME",
			envVal:  "testdata",
			wantErr: false,
			want: &KubeConfig{
				Path:           "testdata/.kube/config",
				CurrentContext: "cntxB",
				Contexts: []KubeContext{
					KubeContext{Name: "cntxA", AWSprofile: "live"},
					KubeContext{Name: "cntxB", AWSprofile: "live"},
					KubeContext{Name: "cntxC", AWSprofile: "dev"},
					KubeContext{Name: "minikube", AWSprofile: ""},
				},
			},
		},
		{
			name:    "1negativ - read malformed config file",
			envVar:  "HOME",
			envVal:  "testdata/malformed",
			wantErr: true,
			want: &KubeConfig{
				Path:     "testdata/malformed/.kube/config",
				Contexts: []KubeContext{},
			},
		},
		{
			name:    "2negativ - read duplicated config file",
			envVar:  "HOME",
			envVal:  "testdata/duplicated",
			wantErr: true,
			want: &KubeConfig{
				Path:     "testdata/duplicated/.kube/config",
				Contexts: []KubeContext{},
			},
		},
		{
			name:    "3negativ - read not existing file",
			envVar:  "HOME",
			envVal:  "testdata/xxxxx",
			wantErr: true,
			want: &KubeConfig{
				Path:     "testdata/xxxxx",
				Contexts: []KubeContext{},
			},
		},
		{
			name:    "3positiv - read valid config file from KUBECONFIG",
			envVar:  "KUBECONFIG",
			envVal:  "testdata/byEnv/.kube/config",
			wantErr: false,
			want: &KubeConfig{
				Path:           "testdata/byEnv/.kube/config",
				CurrentContext: "cntxB",
				Contexts: []KubeContext{
					KubeContext{Name: "cntxA", AWSprofile: ""},
					KubeContext{Name: "cntxB", AWSprofile: ""},
					KubeContext{Name: "cntxC", AWSprofile: ""},
					KubeContext{Name: "minikube", AWSprofile: ""},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := os.Setenv(tt.envVar, tt.envVal); err != nil {
				t.Errorf("GetConfigFile() error = %v", err)
				return
			}
			got, err := GetConfigFile()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetConfigFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got.Contexts) != len(tt.want.Contexts) {
				t.Errorf("GetConfigFile() got.Contexts = %v, want.Contexts %v", got.Contexts, tt.want.Contexts)
				return
			}
			if got.CurrentContext != tt.want.CurrentContext {
				t.Errorf("GetConfigFile() got.CurrentContext = %v, want.CurrentContext %v", got.CurrentContext, tt.want.CurrentContext)
				return
			}
			for idx, ctx := range got.Contexts {
				if tt.want.Contexts[idx].Name != ctx.Name {
					t.Errorf("GetConfigFile() got.Contexts.Name = %v, want.Contexts.Name %v", got.Contexts[idx].Name, tt.want.Contexts[idx].Name)
				}
				if tt.want.Contexts[idx].AWSprofile != ctx.AWSprofile {
					t.Errorf("GetConfigFile() got.Contexts.AWSprofile = %v, want.Contexts.Name %v", got.Contexts[idx].AWSprofile, tt.want.Contexts[idx].AWSprofile)
				}
			}
		})
	}
}

func TestKubeConfig_GetContextNames(t *testing.T) {
	type fields struct {
		Contexts       []KubeContext
		CurrentContext string
		Path           string
	}

	tests := []struct {
		name      string
		fields    fields
		wantNames []string
	}{
		{
			name: "0positiv - simple return the correct profile names",
			fields: fields{
				Path:           "testdata/.kube/config",
				CurrentContext: "cntxB",
				Contexts: []KubeContext{
					KubeContext{Name: "cntxA", AWSprofile: "live"},
					KubeContext{Name: "cntxC", AWSprofile: "dev"},
					KubeContext{Name: "cntxB", AWSprofile: "live"},
					KubeContext{Name: "minikube", AWSprofile: ""},
				},
			},
			wantNames: []string{"cntxA", "cntxB", "cntxC", "minikube"},
		},
		{
			name: "1positiv - empty context list",
			fields: fields{
				Path:           "testdata/.kube/config",
				CurrentContext: "cntxA",
				Contexts:       []KubeContext{},
			},
			wantNames: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &KubeConfig{
				Contexts:       tt.fields.Contexts,
				CurrentContext: tt.fields.CurrentContext,
				Path:           tt.fields.Path,
			}
			gotNames := k.GetContextNames()
			if len(gotNames) != len(tt.wantNames) {
				t.Errorf("KubeConfig.GetContextNames() = %v, want %v", gotNames, tt.wantNames)
				return
			}
			for idx, n := range gotNames {
				if tt.wantNames[idx] != n {
					t.Errorf("KubeConfig.GetContextNames() = %v, want %v", gotNames, tt.wantNames)
					return
				}
			}
		})
	}
}

func TestKubeConfig_GetContextBy(t *testing.T) {
	type fields struct {
		Contexts       []KubeContext
		CurrentContext string
		Path           string
	}
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *KubeContext
		wantIdx int
		wantErr bool
	}{
		{
			name: "0positiv - return existing context",
			fields: fields{
				Path:           "testdata/.kube/config",
				CurrentContext: "cntxB",
				Contexts: []KubeContext{
					KubeContext{Name: "cntxA", AWSprofile: "live"},
					KubeContext{Name: "cntxC", AWSprofile: "dev"},
					KubeContext{Name: "cntxB", AWSprofile: "live"},
					KubeContext{Name: "minikube", AWSprofile: ""},
				},
			},
			args: args{
				name: "minikube",
			},
			want:    &KubeContext{Name: "minikube", AWSprofile: ""},
			wantIdx: 3,
			wantErr: false,
		},
		{
			name: "1negativ - return nil",
			fields: fields{
				Path:           "testdata/.kube/config",
				CurrentContext: "cntxB",
				Contexts: []KubeContext{
					KubeContext{Name: "cntxA", AWSprofile: "live"},
					KubeContext{Name: "cntxC", AWSprofile: "dev"},
					KubeContext{Name: "cntxB", AWSprofile: "live"},
					KubeContext{Name: "minikube", AWSprofile: ""},
				},
			},
			args: args{
				name: "xxxxxx",
			},
			want:    &KubeContext{},
			wantIdx: -1,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &KubeConfig{
				Contexts:       tt.fields.Contexts,
				CurrentContext: tt.fields.CurrentContext,
				Path:           tt.fields.Path,
			}
			got, idx, err := k.GetContextBy(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetContextBy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.Name != tt.want.Name {
				t.Errorf("GetContextBy() got = %v, want = %v", got.Name, tt.want.Name)
				return
			}
			if idx != tt.wantIdx {
				t.Errorf("GetContextBy() index got = %v, want = %v", idx, tt.wantIdx)
				return
			}

		})
	}
}

func TestKubeConfig_SetContextTo(t *testing.T) {
	type fields struct {
		Contexts       []KubeContext
		CurrentContext string
		Path           string
	}
	type args struct {
		contextName string
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantErr     bool
		wantCtx     string
		wantProfile string
		envVar      string
		envVal      string
		setDefault  string
	}{
		{
			name: "0positive - change current context",
			fields: fields{
				Path:           testFilePath,
				CurrentContext: "cntxB",
				Contexts: []KubeContext{
					KubeContext{Name: "cntxA", AWSprofile: "live"},
					KubeContext{Name: "cntxC", AWSprofile: "dev"},
					KubeContext{Name: "cntxB", AWSprofile: "live"},
					KubeContext{Name: "minikube", AWSprofile: ""},
				},
			},
			args:        args{"cntxA"},
			wantErr:     false,
			wantCtx:     "cntxA",
			wantProfile: "live",
			envVar:      "HOME", // for awsdefault
			envVal:      "testdata",
		},
		{
			name: "1negative - context not found but profile and current context keep the same",
			fields: fields{
				Path:           "testdata/.kube/config", // fix contexts
				CurrentContext: "cntxB",
				Contexts: []KubeContext{
					KubeContext{Name: "cntxA", AWSprofile: "live"},
					KubeContext{Name: "cntxC", AWSprofile: "dev"},
					KubeContext{Name: "cntxB", AWSprofile: "live"},
					KubeContext{Name: "minikube", AWSprofile: ""},
				},
			},
			args:        args{"xxxx"},
			wantErr:     true,
			wantCtx:     "cntxB",
			wantProfile: "dev",
			envVar:      "HOME", // for awsdefault
			envVal:      "testdata",
			setDefault:  "dev",
		},
		{
			name: "2negative - no aws credentials file found, also not context set",
			fields: fields{
				Path:           "testdata/.kube/config", // fix contexts
				CurrentContext: "cntxB",
				Contexts: []KubeContext{
					KubeContext{Name: "cntxA", AWSprofile: "live"},
					KubeContext{Name: "cntxC", AWSprofile: "dev"},
					KubeContext{Name: "cntxB", AWSprofile: "live"},
					KubeContext{Name: "minikube", AWSprofile: ""},
				},
			},
			args:        args{"cntxC"},
			wantErr:     true,
			wantCtx:     "cntxB",
			wantProfile: "",
			envVar:      "HOME",           // for awsdefault and kube config
			envVal:      "testdata/byEnv", // only kube config file is there, no aws file
		},
		{
			name: "3negative - no aws profile found for context, nothing should changed",
			fields: fields{
				Path:           "testdata/.kube/config", // fix contexts
				CurrentContext: "cntxB",
				Contexts: []KubeContext{
					KubeContext{Name: "cntxA", AWSprofile: "live"},
					KubeContext{Name: "cntxC", AWSprofile: "dev"},
					KubeContext{Name: "cntxB", AWSprofile: "live"},
					KubeContext{Name: "minikube", AWSprofile: ""},
				},
			},
			args:        args{"minikube"},
			wantErr:     true,
			wantCtx:     "cntxB",
			wantProfile: "no default",
			envVar:      "HOME", // for awsdefault and kube config
			envVal:      "testdata",
		},
		{
			name: "4negative - unknown AWS profile configured, nothing should changed",
			fields: fields{
				Path:           "testdata/.kube/config", // fix contexts
				CurrentContext: "cntxB",
				Contexts: []KubeContext{
					KubeContext{Name: "cntxA", AWSprofile: "unknown"},
					KubeContext{Name: "cntxC", AWSprofile: "dev"},
					KubeContext{Name: "cntxB", AWSprofile: "live"},
					KubeContext{Name: "minikube", AWSprofile: ""},
				},
			},
			args:        args{"cntxA"},
			wantErr:     true,
			wantCtx:     "cntxB",
			wantProfile: "no default",
			envVar:      "HOME", // for awsdefault and kube config
			envVal:      "testdata",
		},
		{
			name: "5positive - unknown AWS profile for the current context configured, but new context has known aws profile.",
			fields: fields{
				Path:           "testdata/.kube/config", // fix contexts
				CurrentContext: "cntxB",
				Contexts: []KubeContext{
					KubeContext{Name: "cntxA", AWSprofile: "live"},
					KubeContext{Name: "cntxC", AWSprofile: "dev"},
					KubeContext{Name: "cntxB", AWSprofile: "xxxxxx"},
					KubeContext{Name: "minikube", AWSprofile: ""},
				},
			},
			args:        args{"cntxC"},
			wantErr:     false,
			wantCtx:     "cntxC",
			wantProfile: "dev",
			envVar:      "HOME", // for awsdefault and kube config
			envVal:      "testdata",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//set up
			if err := os.Setenv(tt.envVar, tt.envVal); err != nil {
				t.Errorf("SetContextTo() error = %v", err)
				return
			}
			if len(tt.setDefault) > 0 {
				preAwsfile, err := awsdefault.GetCredentialsFile()
				if err != nil {
					t.Errorf("KubeConfig.SetContextTo() error read result credentials file = %v", err)
					return
				}
				if err := preAwsfile.SetDefaultTo(tt.setDefault); err != nil {
					t.Errorf("SetContextTo() setDefault error = %v", err)
					return
				}
			}

			k := &KubeConfig{
				Contexts:       tt.fields.Contexts,
				CurrentContext: tt.fields.CurrentContext,
				Path:           tt.fields.Path,
			}
			if err := k.SetContextTo(tt.args.contextName); (err != nil) != tt.wantErr {
				t.Errorf("KubeConfig.SetContextTo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err := os.Setenv("KUBECONFIG", k.Path); err != nil {
				t.Errorf("SetContextTo() error = %v", err)
				return
			}
			r, err := GetConfigFile()
			if err != nil {
				t.Errorf("KubeConfig.SetContextTo() error read result config file = %v", err)
				return
			}
			if r.CurrentContext != tt.wantCtx {
				t.Errorf("KubeConfig.SetContextTo() got = %v, want = %v", r.CurrentContext, tt.wantCtx)
			}
			if len(tt.wantProfile) > 0 {
				awsfile, err := awsdefault.GetCredentialsFile()
				if err != nil {
					t.Errorf("KubeConfig.SetContextTo() error read result credentials file = %v", err)
					return
				}
				n, _, err := awsfile.GetUsedProfileNameAndIndex()
				if err != nil {
					t.Errorf("KubeConfig.SetContextTo() error read result aws default profile = %v", err)
				}
				if n != tt.wantProfile {
					t.Errorf("KubeConfig.SetContextTo() wrong aws profile. Got = %v, want = %v", n, tt.wantProfile)
				}

				// teardown
				if err = awsfile.UnSetDefault(); err != nil {
					t.Errorf("KubeConfig.SetContextTo() error unset aws default profile = %v", err)
					return
				}
			}

		})
	}
}

func TestKubeConfig_AddProfileTo(t *testing.T) {
	type fields struct {
		Contexts       []KubeContext
		CurrentContext string
		Path           string
	}
	type args struct {
		contextName string
		profileName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		envVar  string
		envVal  string
	}{
		{
			name: "0positive - add a profile to the context",
			fields: fields{
				Path:           "testdata/.kube/config", // fix contexts
				CurrentContext: "cntxB",
				Contexts: []KubeContext{
					KubeContext{Name: "cntxA", AWSprofile: ""},
					KubeContext{Name: "cntxC", AWSprofile: "dev"},
					KubeContext{Name: "cntxB", AWSprofile: ""},
					KubeContext{Name: "minikube", AWSprofile: ""},
				},
			},
			args:    args{contextName: "cntxA", profileName: "dev"},
			wantErr: false,
			envVar:  "HOME", // for awsdefault and kube config
			envVal:  "testdata",
		},
		{
			name: "1negative - try to add a none existing profile",
			fields: fields{
				Path:           "testdata/.kube/config", // fix contexts
				CurrentContext: "cntxB",
				Contexts: []KubeContext{
					KubeContext{Name: "cntxA", AWSprofile: ""},
					KubeContext{Name: "cntxC", AWSprofile: "dev"},
					KubeContext{Name: "cntxB", AWSprofile: ""},
					KubeContext{Name: "minikube", AWSprofile: ""},
				},
			},
			args:    args{contextName: "cntxA", profileName: "xxxxx"},
			wantErr: true,
			envVar:  "HOME", // for awsdefault and kube config
			envVal:  "testdata",
		},
		{
			name: "2negative - try to change a none existing context",
			fields: fields{
				Path:           "testdata/.kube/config", // fix contexts
				CurrentContext: "cntxB",
				Contexts: []KubeContext{
					KubeContext{Name: "cntxA", AWSprofile: ""},
					KubeContext{Name: "cntxC", AWSprofile: "dev"},
					KubeContext{Name: "cntxB", AWSprofile: ""},
					KubeContext{Name: "minikube", AWSprofile: ""},
				},
			},
			args:    args{contextName: "xxxxx", profileName: "dev"},
			wantErr: true,
			envVar:  "HOME", // for awsdefault and kube config
			envVal:  "testdata",
		},
		{
			name: "3negative - try to read a none existing aws credentials file",
			fields: fields{
				Path:           "testdata/.kube/config", // fix contexts
				CurrentContext: "cntxB",
				Contexts: []KubeContext{
					KubeContext{Name: "cntxA", AWSprofile: ""},
					KubeContext{Name: "cntxC", AWSprofile: "dev"},
					KubeContext{Name: "cntxB", AWSprofile: ""},
					KubeContext{Name: "minikube", AWSprofile: ""},
				},
			},
			args:    args{contextName: "cntxC", profileName: "dev"},
			wantErr: true,
			envVar:  "HOME", // for awsdefault and kube config
			envVal:  "xxxxxxx",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//set up
			if err := os.Setenv(tt.envVar, tt.envVal); err != nil {
				t.Errorf("KubeConfig.AddProfileTo() error = %v", err)
				return
			}
			k := &KubeConfig{
				Contexts:       tt.fields.Contexts,
				CurrentContext: tt.fields.CurrentContext,
				Path:           tt.fields.Path,
			}
			if err := os.Setenv("KUBECONFIG", k.Path); err != nil {
				t.Errorf("KubeConfig.AddProfileTo() error = %v", err)
				return
			}
			if err := k.AddProfileTo(tt.args.contextName, tt.args.profileName); (err != nil) != tt.wantErr {
				t.Errorf("KubeConfig.AddProfileTo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			return

			r, err := GetConfigFile()
			if err != nil {
				t.Errorf("KubeConfig.AddProfileTo() error read result config file = %v", err)
				return
			}
			fmt.Printf("r = %+v\n", r)
			ctx, _, err := r.GetContextBy(tt.args.contextName)
			if err != nil {
				t.Errorf("KubeConfig.AddProfileTo() error read result context = %v", err)
				return
			}
			fmt.Printf("ctx = %+v\n", ctx)
			if ctx.AWSprofile != tt.args.profileName {
				t.Errorf("KubeConfig.AddProfileTo() got = %v, want = %v", ctx.AWSprofile, tt.args.profileName)
				return
			}
			// teardown
			if err = ioutil.WriteFile("testdata/.kube/config", testFileContent, 0644); err != nil {
				t.Errorf("KubeConfig.AddProfileTo() error reset config file = %v", err)
				return
			}
		})
	}
}

func TestKubeConfig_AddNamespaceTo(t *testing.T) {
	type fields struct {
		ApiVersion     string
		Kind           string
		Preferences    interface{}
		Contexts       []KubeContext
		CurrentContext string
		Path           string
	}
	type args struct {
		contextName string
		namespace   string
	}
	tests := []struct {
		name          string
		fields        fields
		args          args
		wantErr       bool
		wantNamespace string
	}{
		{
			name: "0positive - add a namespace to the context",
			fields: fields{
				Path:           "testdata/.kube/config", // fix contexts
				CurrentContext: "cntxB",
				Contexts: []KubeContext{
					KubeContext{Name: "cntxA", AWSprofile: ""},
					KubeContext{Name: "cntxC", AWSprofile: "dev"},
					KubeContext{Name: "cntxB", AWSprofile: ""},
					KubeContext{Name: "minikube", AWSprofile: ""},
				},
			},
			args:          args{contextName: "cntxA", namespace: "n1"},
			wantErr:       false,
			wantNamespace: "n1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &KubeConfig{
				ApiVersion:     tt.fields.ApiVersion,
				Kind:           tt.fields.Kind,
				Preferences:    tt.fields.Preferences,
				Contexts:       tt.fields.Contexts,
				CurrentContext: tt.fields.CurrentContext,
				Path:           tt.fields.Path,
			}
			if err := k.AddNamespaceTo(tt.args.contextName, tt.args.namespace); (err != nil) != tt.wantErr {
				t.Errorf("KubeConfig.AddNamespaceTo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			file, err := GetConfigFile()
			if err != nil {
				t.Errorf("KubeConfig.AddNamespaceTo() error = %v", err)
				return
			}
			ctx, _, err := file.GetContextBy(tt.args.contextName)
			if err != nil {
				t.Errorf("KubeConfig.AddNamespaceTo() error = %v", err)
				return
			}
			if ctx.Namespace != tt.wantNamespace {
				t.Errorf("KubeConfig.AddNamespaceTo() got = %v, want = %v", ctx.Namespace, tt.wantNamespace)
				return
			}

		})
	}
}

func TestKubeConfig_UnSetDefault(t *testing.T) {
	type fields struct {
		CurrentContext string
		Path           string
	}
	tests := []struct {
		name               string
		fields             fields
		wantErr            bool
		wantCurrentContext string
	}{
		{
			name: "0positive - remove current-context",
			fields: fields{
				CurrentContext: "cntxB",
				Path:           testFilePath,
			},
			wantErr:            false,
			wantCurrentContext: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := os.Setenv("KUBECONFIG", tt.fields.Path); err != nil {
				t.Errorf("UnSetDefault() error = %v", err)
				return
			}
			k := &KubeConfig{
				CurrentContext: tt.fields.CurrentContext,
				Path:           tt.fields.Path,
			}
			if err := k.UnSetDefault(); (err != nil) != tt.wantErr {
				t.Errorf("KubeConfig.UnSetDefault() error = %v, wantErr %v", err, tt.wantErr)
			}
			r, err := GetConfigFile()
			if err != nil {
				t.Errorf("KubeConfig.UnSetDefault() error read result config file = %v", err)
				return
			}
			if r.CurrentContext != tt.wantCurrentContext {
				t.Errorf("KubeConfig.UnSetDefault() got = %v, want = %v", r.CurrentContext, tt.wantCurrentContext)
			}

		})
	}
}
