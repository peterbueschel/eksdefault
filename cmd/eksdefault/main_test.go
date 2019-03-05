package main

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/peterbueschel/awsdefault"
)

var (
	self            = "eksdefault"
	testFileContent = []byte("")
)

func TestMain(m *testing.M) {
	// setup
	var err error
	testFileContent, err = ioutil.ReadFile("testdata/.kube/config")
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	// run
	code := m.Run()

	// teardown
	os.Setenv("HOME", "testdata")
	awsfile, err := awsdefault.GetCredentialsFile()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	if err = awsfile.UnSetDefault(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	if err = ioutil.WriteFile("testdata/.kube/config", testFileContent, 0644); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	os.Exit(code)
}
func Test_runMain(t *testing.T) {
	table := `ID       CURRENT       KUBE CONTEXT       AWS PROFILE       CLUSTER        USER           NAMESPACE
0                      cntxA              live              clstrA         userA          aaaaa
1        *             cntxB              live              clstrB         userB          bbbbb
2                      cntxC              dev               clstrC         userC          ccccc
3                      minikube                             minikube       minikube       
`
	tableChangeProfile := `ID       CURRENT       KUBE CONTEXT       AWS PROFILE          CLUSTER        USER           NAMESPACE
0                      cntxA              live                 clstrA         userA          aaaaa
1        *             cntxB              anotherprofile       clstrB         userB          bbbbb
2                      cntxC              dev                  clstrC         userC          ccccc
3                      minikube                                minikube       minikube       
`
	tableChangeProfile2 := `ID       CURRENT       KUBE CONTEXT       AWS PROFILE          CLUSTER        USER           NAMESPACE
0                      cntxA              live                 clstrA         userA          aaaaa
1        *             cntxB              live                 clstrB         userB          bbbbb
2                      cntxC              anotherprofile       clstrC         userC          ccccc
3                      minikube                                minikube       minikube       
`
	tableChangeNamespace := `ID       CURRENT       KUBE CONTEXT       AWS PROFILE       CLUSTER        USER           NAMESPACE
0                      cntxA              live              clstrA         userA          aaaaa
1        *             cntxB              live              clstrB         userB          anotherNamespace
2                      cntxC              dev               clstrC         userC          ccccc
3                      minikube                             minikube       minikube       
`
	tableChangeNamespace2 := `ID       CURRENT       KUBE CONTEXT       AWS PROFILE       CLUSTER        USER           NAMESPACE
0                      cntxA              live              clstrA         userA          aaaaa
1        *             cntxB              live              clstrB         userB          bbbbb
2                      cntxC              dev               clstrC         userC          anotherNamespace
3                      minikube                             minikube       minikube       
`
	tableNewContext := `ID       CURRENT       KUBE CONTEXT       AWS PROFILE       CLUSTER        USER           NAMESPACE
0                      I                  love              my             new            tool
1                      cntxA              live              clstrA         userA          aaaaa
2        *             cntxB              live              clstrB         userB          bbbbb
3                      cntxC              dev               clstrC         userC          ccccc
4                      minikube                             minikube       minikube       
`
	tableCopyContext := `ID       CURRENT       KUBE CONTEXT       AWS PROFILE       CLUSTER        USER           NAMESPACE
0                      cntxA              live              clstrA         userA          aaaaa
1        *             cntxB              live              clstrB         userB          bbbbb
2                      cntxC              dev               clstrC         userC          ccccc
3                      copied             live              clstrB         userB          bbbbb
4                      minikube                             minikube       minikube       
`
	type args struct {
		args []string
	}
	tests := []struct {
		name               string
		args               args
		want               string
		wantErr            bool
		envVar             string
		envVal             string
		verify             string
		wantCurrentContext string
	}{
		{
			name:    "negative - no file found",
			args:    args{[]string{self, "ls", "-s"}},
			want:    "",
			wantErr: true,
			envVar:  "KUBECONFIG",
			envVal:  "testdata/.kube/somewhere",
		},
		// Get
		{
			name:    "0getCurrentContext - positive - with short flag",
			args:    args{[]string{self, "ls", "-s"}},
			want:    "cntxA\ncntxB\ncntxC\nminikube\n",
			wantErr: false,
			envVar:  "KUBECONFIG",
			envVal:  "testdata/.kube/config",
		},
		{
			name:    "1getCurrentContext - positive - without short flag - as table",
			args:    args{[]string{self, "ls"}},
			want:    table,
			wantErr: false,
			envVar:  "KUBECONFIG",
			envVal:  "testdata/.kube/config",
		},
		// Add
		{
			name:    "0addContext - negative - missing argument",
			args:    args{[]string{self, "new"}},
			want:    "",
			wantErr: true,
			envVar:  "KUBECONFIG",
			envVal:  "testdata/.kube/config",
		},
		{
			name:    "1addContext - positive - new context",
			args:    args{[]string{self, "new", "I", "-p", "love", "-c", "my", "-u", "new", "-n", "tool"}},
			want:    "",
			wantErr: false,
			envVar:  "KUBECONFIG",
			envVal:  "testdata/.kube/config",
			verify:  tableNewContext,
		},
		{
			name:    "2addContext - negative - name already exists",
			args:    args{[]string{self, "new", "cntxA", "-p", "love", "-c", "my", "-u", "new", "-n", "tool"}},
			want:    "",
			wantErr: true,
			envVar:  "KUBECONFIG",
			envVal:  "testdata/.kube/config",
			verify:  table,
		},
		// Copy
		{
			name:    "0copyContext - negative - missing argument",
			args:    args{[]string{self, "copy"}},
			want:    "",
			wantErr: true,
			envVar:  "KUBECONFIG",
			envVal:  "testdata/.kube/config",
		},
		{
			name:    "1copyContext - positive - copy context and overwrites nothing",
			args:    args{[]string{self, "cp", "copied"}},
			want:    "",
			wantErr: false,
			envVar:  "KUBECONFIG",
			envVal:  "testdata/.kube/config",
			verify:  tableCopyContext,
		},
		{
			name:    "2copyContext - positive - copy context and overwrites all",
			args:    args{[]string{self, "cp", "I", "-p", "love", "-c", "my", "-u", "new", "-n", "tool"}},
			want:    "",
			wantErr: false,
			envVar:  "KUBECONFIG",
			envVal:  "testdata/.kube/config",
			verify:  tableNewContext,
		},
		{
			name:    "3copyContext - negative - name already exists",
			args:    args{[]string{self, "cp", "cntxA", "-p", "love", "-c", "my", "-u", "new", "-n", "tool"}},
			want:    "",
			wantErr: true,
			envVar:  "KUBECONFIG",
			envVal:  "testdata/.kube/config",
			verify:  table,
		},
		{
			name:    "3copyContext - negative - current context not found in file",
			args:    args{[]string{self, "cp", "valid but current-context wrong"}},
			want:    "",
			wantErr: true,
			envVar:  "KUBECONFIG",
			envVal:  "testdata/.kube/wrongCurrentContext",
		},
		// Namespace
		{
			name:    "0useNamespace - positive - current-context",
			args:    args{[]string{self, "namespace", "anotherNamespace"}},
			want:    "",
			wantErr: false,
			envVar:  "KUBECONFIG",
			envVal:  "testdata/.kube/config",
			verify:  tableChangeNamespace,
		},
		{
			name:    "1useNamespace - positive - use ID of current context",
			args:    args{[]string{self, "namespace", "anotherNamespace", "1"}},
			want:    "",
			wantErr: false,
			envVar:  "KUBECONFIG",
			envVal:  "testdata/.kube/config",
			verify:  tableChangeNamespace,
		},
		{
			name:    "2useNamespace - positive - use ID",
			args:    args{[]string{self, "namespace", "anotherNamespace", "2"}},
			want:    "",
			wantErr: false,
			envVar:  "KUBECONFIG",
			envVal:  "testdata/.kube/config",
			verify:  tableChangeNamespace2,
		},
		{
			name:    "3useNamespace - positive - use context name",
			args:    args{[]string{self, "namespace", "anotherNamespace", "cntxC"}},
			want:    "",
			wantErr: false,
			envVar:  "KUBECONFIG",
			envVal:  "testdata/.kube/config",
			verify:  tableChangeNamespace2,
		},
		{
			name:    "4useNamespace - negative - missing arg",
			args:    args{[]string{self, "namespace"}},
			want:    "",
			wantErr: true,
			envVar:  "KUBECONFIG",
			envVal:  "testdata/.kube/config",
		},
		{
			name:    "5useNamespace - negative - unknown id",
			args:    args{[]string{self, "namespace", "anyway", "123456"}},
			want:    "",
			wantErr: true,
			envVar:  "KUBECONFIG",
			envVal:  "testdata/.kube/config",
		},
		// Profile
		{
			name:    "0useProfile - positive - current-context",
			args:    args{[]string{self, "profile", "anotherprofile"}},
			want:    "",
			wantErr: false,
			envVar:  "KUBECONFIG",
			envVal:  "testdata/.kube/config",
			verify:  tableChangeProfile,
		},
		{
			name:    "1useProfile - positive - use ID of current context",
			args:    args{[]string{self, "profile", "anotherprofile", "1"}},
			want:    "",
			wantErr: false,
			envVar:  "KUBECONFIG",
			envVal:  "testdata/.kube/config",
			verify:  tableChangeProfile,
		},
		{
			name:    "2useProfile - positive - use ID",
			args:    args{[]string{self, "profile", "anotherprofile", "2"}},
			want:    "",
			wantErr: false,
			envVar:  "KUBECONFIG",
			envVal:  "testdata/.kube/config",
			verify:  tableChangeProfile2,
		},
		{
			name:    "3useProfile - positive - use context name",
			args:    args{[]string{self, "profile", "anotherprofile", "cntxC"}},
			want:    "",
			wantErr: false,
			envVar:  "KUBECONFIG",
			envVal:  "testdata/.kube/config",
			verify:  tableChangeProfile2,
		},
		{
			name:    "4useProfile - negative - missing arg",
			args:    args{[]string{self, "profile"}},
			want:    "",
			wantErr: true,
			envVar:  "KUBECONFIG",
			envVal:  "testdata/.kube/config",
		},
		{
			name:    "5useProfile - negative - unknown id",
			args:    args{[]string{self, "profile", "anyway", "123456"}},
			want:    "",
			wantErr: true,
			envVar:  "KUBECONFIG",
			envVal:  "testdata/.kube/config",
		},
		// Set
		{
			name:               "0setDefaultContext - positive",
			args:               args{[]string{self, "set", "cntxA"}},
			want:               "",
			wantErr:            false,
			envVar:             "KUBECONFIG",
			envVal:             "testdata/.kube/config",
			wantCurrentContext: "cntxA\n",
		},
		{
			name:               "1setDefaultContext - positive - use ID",
			args:               args{[]string{self, "set", "0"}},
			want:               "",
			wantErr:            false,
			envVar:             "KUBECONFIG",
			envVal:             "testdata/.kube/config",
			wantCurrentContext: "cntxA\n",
		},
		{
			name:               "2setDefaultContext - negative - context not exists",
			args:               args{[]string{self, "set", "xxxx"}},
			want:               "",
			wantErr:            true,
			envVar:             "KUBECONFIG",
			envVal:             "testdata/.kube/config",
			wantCurrentContext: "cntxB\n", // still the current-context
		},
		{
			name:               "2setDefaultContext - negative - ID not exists",
			args:               args{[]string{self, "set", "123456"}},
			want:               "",
			wantErr:            true,
			envVar:             "KUBECONFIG",
			envVal:             "testdata/.kube/config",
			wantCurrentContext: "cntxB\n", // still the current-context
		},
		{
			name:               "3setDefaultContext - negative - missing arg",
			args:               args{[]string{self, "set"}},
			want:               "",
			wantErr:            true,
			envVar:             "KUBECONFIG",
			envVal:             "testdata/.kube/config",
			wantCurrentContext: "cntxB\n", // still the current-context
		},
		{
			name:               "3setDefaultContext - negative - missing aws profile",
			args:               args{[]string{self, "set", "minikube"}},
			want:               "",
			wantErr:            true,
			envVar:             "KUBECONFIG",
			envVal:             "testdata/.kube/config",
			wantCurrentContext: "cntxB\n", // still the current-context
		},
		// GetCurrent
		{
			name:    "0getCurrentContext - positive",
			args:    args{[]string{self, "is"}},
			want:    "cntxB\n",
			wantErr: false,
			envVar:  "KUBECONFIG",
			envVal:  "testdata/.kube/config",
		},
		{
			name:    "1getCurrentContext - positive - no current-context",
			args:    args{[]string{self, "is"}},
			want:    "no current-context set\n",
			wantErr: false,
			envVar:  "KUBECONFIG",
			envVal:  "testdata/.kube/noCurrentContext",
		},
		// Unset
		{
			name:               "0unSetDefaultContext - positive",
			args:               args{[]string{self, "rm"}},
			want:               "",
			wantErr:            false,
			envVar:             "KUBECONFIG",
			envVal:             "testdata/.kube/config",
			wantCurrentContext: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := os.Setenv("HOME", "testdata"); err != nil { // aws credentials
				t.Errorf("runMain() error = %v", err)
				return
			}
			if err := os.Setenv(tt.envVar, tt.envVal); err != nil {
				t.Errorf("runMain() error = %v", err)
				return
			}
			got, err := runMain(tt.args.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("runMain() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("runMain() =\n%#+v, want\n%#+v", got, tt.want)
			}
			if len(tt.verify) > 0 {
				got, err := runMain([]string{self, "ls"})
				if err != nil {
					t.Errorf("runMain() error verify = %v", err)
				}
				if got != tt.verify {
					t.Errorf("runMain() got =\n%+v, verify =\n%+v", got, tt.verify)
				}
			}
			if len(tt.wantCurrentContext) > 0 {
				got, err := runMain([]string{self, "is"})
				if err != nil {
					t.Errorf("runMain() error get current-context = %v", err)
				}
				if got != tt.wantCurrentContext {
					t.Errorf("runMain() got = %+v, wantCurrentContext = %+v", got, tt.wantCurrentContext)
				}
			}
			// teardown
			os.Unsetenv(tt.envVar)
			if err = ioutil.WriteFile("testdata/.kube/config", testFileContent, 0644); err != nil {
				log.Fatal(err)
				os.Exit(1)
			}
		})
	}
}
