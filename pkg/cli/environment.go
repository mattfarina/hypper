package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/pflag"
	"helm.sh/helm/v3/pkg/helmpath"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// defaultMaxHistory sets the maximum number of releases to 0: unlimited
const defaultMaxHistory = 10

// EnvSettings describes all of the environment settings.
type EnvSettings struct {
	namespace string
	config    *genericclioptions.ConfigFlags

	// KubeConfig is the path to the kubeconfig file
	KubeConfig string
	// KubeContext is the name of the kubeconfig context.
	KubeContext string
	// Bearer KubeToken used for authentication
	KubeToken string
	// Username to impersonate for the operation
	KubeAsUser string
	// Groups to impersonate for the operation, multiple groups parsed from a comma delimited list
	KubeAsGroups []string
	// Kubernetes API Server Endpoint for authentication
	KubeAPIServer string
	// Custom certificate authority file.
	KubeCaFile string
	// Debug indicates whether or not Helm is running in Debug mode.
	Debug bool
	// RegistryConfig is the path to the registry config file.
	RegistryConfig string
	// RepositoryConfig is the path to the repositories file.
	RepositoryConfig string
	// RepositoryCache is the path to the repository cache directory.
	RepositoryCache string
	// PluginsDirectory is the path to the plugins directory.
	PluginsDirectory string
	// MaxHistory is the max release history maintained.
	MaxHistory int

	Verbose  bool
	NoColors bool
	NoEmojis bool
}

// New is a constructor of EnvSettings
func New() *EnvSettings {
	env := &EnvSettings{
		namespace:        os.Getenv("HELM_NAMESPACE"),
		MaxHistory:       envIntOr("HELM_MAX_HISTORY", defaultMaxHistory),
		KubeContext:      os.Getenv("HELM_KUBECONTEXT"),
		KubeToken:        os.Getenv("HELM_KUBETOKEN"),
		KubeAsUser:       os.Getenv("HELM_KUBEASUSER"),
		KubeAsGroups:     envCSV("HELM_KUBEASGROUPS"),
		KubeAPIServer:    os.Getenv("HELM_KUBEAPISERVER"),
		KubeCaFile:       os.Getenv("HELM_KUBECAFILE"),
		PluginsDirectory: envOr("HELM_PLUGINS", helmpath.DataPath("plugins")),
		RegistryConfig:   envOr("HELM_REGISTRY_CONFIG", helmpath.ConfigPath("registry.json")),
		RepositoryConfig: envOr("HELM_REPOSITORY_CONFIG", helmpath.ConfigPath("repositories.yaml")),
		RepositoryCache:  envOr("HELM_REPOSITORY_CACHE", helmpath.CachePath("repository")),
	}
	env.Debug, _ = strconv.ParseBool(os.Getenv("HYPPER_DEBUG"))
	env.Verbose, _ = strconv.ParseBool(os.Getenv("HYPPER_TRACE"))
	env.NoColors, _ = strconv.ParseBool(os.Getenv("HYPPER_NOCOLORS"))
	env.NoEmojis, _ = strconv.ParseBool(os.Getenv("HYPPER_NOEMOJIS"))

	// bind to kubernetes config flags
	env.config = &genericclioptions.ConfigFlags{
		Namespace:        &env.namespace,
		Context:          &env.KubeContext,
		BearerToken:      &env.KubeToken,
		APIServer:        &env.KubeAPIServer,
		CAFile:           &env.KubeCaFile,
		KubeConfig:       &env.KubeConfig,
		Impersonate:      &env.KubeAsUser,
		ImpersonateGroup: &env.KubeAsGroups,
	}
	return env
}

// AddFlags binds flags to the given flagset.
func (s *EnvSettings) AddFlags(fs *pflag.FlagSet) {
	// Hypper specific
	fs.StringVarP(&s.namespace, "namespace", "n", s.namespace, "namespace scope for this request")
	fs.BoolVar(&s.Debug, "debug", s.Debug, "enable verbose output")
	fs.BoolVar(&s.NoColors, "no-colors", s.NoColors, "disable colors")
	fs.BoolVar(&s.NoEmojis, "no-emojis", s.NoEmojis, "disable emojis")

	// imported from Helm
	fs.StringVar(&s.KubeConfig, "kubeconfig", "", "path to the kubeconfig file")
	fs.StringVar(&s.KubeContext, "kube-context", s.KubeContext, "name of the kubeconfig context to use")
	fs.StringVar(&s.KubeToken, "kube-token", s.KubeToken, "bearer token used for authentication")
	fs.StringVar(&s.KubeAsUser, "kube-as-user", s.KubeAsUser, "username to impersonate for the operation")
	fs.StringArrayVar(&s.KubeAsGroups, "kube-as-group", s.KubeAsGroups, "group to impersonate for the operation, this flag can be repeated to specify multiple groups.")
	fs.StringVar(&s.KubeAPIServer, "kube-apiserver", s.KubeAPIServer, "the address and the port for the Kubernetes API server")
	fs.StringVar(&s.KubeCaFile, "kube-ca-file", s.KubeCaFile, "the certificate authority file for the Kubernetes API server connection")
	fs.StringVar(&s.RegistryConfig, "registry-config", s.RegistryConfig, "path to the registry config file")
	fs.StringVar(&s.RepositoryConfig, "repository-config", s.RepositoryConfig, "path to the file containing repository names and URLs")
	fs.StringVar(&s.RepositoryCache, "repository-cache", s.RepositoryCache, "path to the file containing cached repository indexes")
}

func envOr(name, def string) string {
	if v, ok := os.LookupEnv(name); ok {
		return v
	}
	return def
}

func envIntOr(name string, def int) int {
	if name == "" {
		return def
	}
	envVal := envOr(name, strconv.Itoa(def))
	ret, err := strconv.Atoi(envVal)
	if err != nil {
		return def
	}
	return ret
}

func envCSV(name string) (ls []string) {
	trimmed := strings.Trim(os.Getenv(name), ", ")
	if trimmed != "" {
		ls = strings.Split(trimmed, ",")
	}
	return
}

func (s *EnvSettings) EnvVars() map[string]string {
	envvars := map[string]string{
		"HELM_BIN":               os.Args[0],
		"HELM_CACHE_HOME":        helmpath.CachePath(""),
		"HELM_CONFIG_HOME":       helmpath.ConfigPath(""),
		"HELM_DATA_HOME":         helmpath.DataPath(""),
		"HELM_DEBUG":             fmt.Sprint(s.Debug),
		"HELM_PLUGINS":           s.PluginsDirectory,
		"HELM_REGISTRY_CONFIG":   s.RegistryConfig,
		"HELM_REPOSITORY_CACHE":  s.RepositoryCache,
		"HELM_REPOSITORY_CONFIG": s.RepositoryConfig,
		"HELM_NAMESPACE":         s.Namespace(),
		"HELM_MAX_HISTORY":       strconv.Itoa(s.MaxHistory),

		// broken, these are populated from helm flags and not kubeconfig.
		"HELM_KUBECONTEXT":   s.KubeContext,
		"HELM_KUBETOKEN":     s.KubeToken,
		"HELM_KUBEASUSER":    s.KubeAsUser,
		"HELM_KUBEASGROUPS":  strings.Join(s.KubeAsGroups, ","),
		"HELM_KUBEAPISERVER": s.KubeAPIServer,
		"HELM_KUBECAFILE":    s.KubeCaFile,

		//hypper specific
		"HYPPER_VERBOSE":  fmt.Sprint(s.Verbose),
		"HYPPER_NOCOLORS": fmt.Sprint(s.NoColors),
		"HYPPER_NOEMOJIS": fmt.Sprint(s.NoEmojis),
	}
	if s.KubeConfig != "" {
		envvars["KUBECONFIG"] = s.KubeConfig
	}
	return envvars
}

// Namespace gets the namespace from the configuration
func (s *EnvSettings) Namespace() string {
	if ns, _, err := s.config.ToRawKubeConfigLoader().Namespace(); err == nil {
		return ns
	}
	return "default"
}

// RESTClientGetter gets the kubeconfig from EnvSettings
func (s *EnvSettings) RESTClientGetter() genericclioptions.RESTClientGetter {
	return s.config
}
