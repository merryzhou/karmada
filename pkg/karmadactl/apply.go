package karmadactl

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	restclient "k8s.io/client-go/rest"
	kubectlapply "k8s.io/kubectl/pkg/cmd/apply"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/templates"

	policyv1alpha1 "github.com/karmada-io/karmada/pkg/apis/policy/v1alpha1"
	"github.com/karmada-io/karmada/pkg/karmadactl/options"
	"github.com/karmada-io/karmada/pkg/util/names"
)

var metadataAccessor = meta.NewAccessor()

// CommandApplyOptions contains the input to the apply command.
type CommandApplyOptions struct {
	// global flags
	options.GlobalCommandOptions
	// apply flags
	KubectlApplyFlags *kubectlapply.ApplyFlags
	Namespace         string
	AllClusters       bool

	kubectlApplyOptions *kubectlapply.ApplyOptions
}

var (
	applyLong = templates.LongDesc(`
		Apply a configuration to a resource by file name or stdin and propagate them into member clusters.
		The resource name must be specified. This resource will be created if it doesn't exist yet.
		To use 'apply', always create the resource initially with either 'apply' or 'create --save-config'.

		JSON and YAML formats are accepted.

		Alpha Disclaimer: the --prune functionality is not yet complete. Do not use unless you are aware of what the current state is. See https://issues.k8s.io/34274.
		
		Note: It implements the function of 'kubectl apply' by default. 
		If you want to propagate them into member clusters, please use 'kubectl apply --all-clusters'.`)

	applyExample = templates.Examples(`
		# Apply the configuration without propagation into member clusters. It acts as 'kubectl apply'.
		%[1]s apply -f manifest.yaml

		# Apply resources from a directory and propagate them into all member clusters.
		%[1]s apply -f dir/ --all-clusters`)
)

// NewCommandApplyOptions returns an initialized CommandApplyOptions instance
func NewCommandApplyOptions() *CommandApplyOptions {
	streams := genericclioptions.IOStreams{In: getIn, Out: getOut, ErrOut: getErr}
	flags := kubectlapply.NewApplyFlags(nil, streams)
	return &CommandApplyOptions{
		KubectlApplyFlags: flags,
	}
}

// NewCmdApply creates the `apply` command
func NewCmdApply(karmadaConfig KarmadaConfig, parentCommand string) *cobra.Command {
	o := NewCommandApplyOptions()
	cmd := &cobra.Command{
		Use:     "apply (-f FILENAME | -k DIRECTORY)",
		Short:   "Apply a configuration to a resource by file name or stdin and propagate them into member clusters",
		Long:    applyLong,
		Example: fmt.Sprintf(applyExample, parentCommand),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := o.Complete(karmadaConfig, cmd, parentCommand, args); err != nil {
				return err
			}
			if err := o.Validate(cmd, args); err != nil {
				return err
			}
			return o.Run()
		},
	}

	o.GlobalCommandOptions.AddFlags(cmd.Flags())
	o.KubectlApplyFlags.AddFlags(cmd)
	cmd.Flags().StringVarP(&o.Namespace, "namespace", "n", o.Namespace, "If present, the namespace scope for this CLI request")
	cmd.Flags().BoolVarP(&o.AllClusters, "all-clusters", "", o.AllClusters, "If present, propagates a group of resources to all member clusters.")
	return cmd
}

// Complete completes all the required options
func (o *CommandApplyOptions) Complete(karmadaConfig KarmadaConfig, cmd *cobra.Command, parentCommand string, args []string) error {
	restConfig, err := karmadaConfig.GetRestConfig(o.KarmadaContext, o.KubeConfig)
	if err != nil {
		return err
	}
	kubeConfigFlags := NewConfigFlags(true).WithDeprecatedPasswordFlag()
	kubeConfigFlags.Namespace = &o.Namespace
	kubeConfigFlags.WrapConfigFn = func(config *restclient.Config) *restclient.Config { return restConfig }
	o.KubectlApplyFlags.Factory = cmdutil.NewFactory(kubeConfigFlags)
	kubectlApplyOptions, err := o.KubectlApplyFlags.ToOptions(cmd, parentCommand, args)
	if err != nil {
		return err
	}
	o.kubectlApplyOptions = kubectlApplyOptions
	return nil
}

// Validate verifies if CommandApplyOptions are valid and without conflicts.
func (o *CommandApplyOptions) Validate(cmd *cobra.Command, args []string) error {
	return o.kubectlApplyOptions.Validate(cmd, args)
}

// Run executes the `apply` command.
func (o *CommandApplyOptions) Run() error {
	if !o.AllClusters {
		return o.kubectlApplyOptions.Run()
	}

	if err := o.generateAndInjectPolices(); err != nil {
		return err
	}

	return o.kubectlApplyOptions.Run()
}

// generateAndInjectPolices generates and injects policies to the given resources.
// It returns an error if any of the policies cannot be generated.
func (o *CommandApplyOptions) generateAndInjectPolices() error {
	// load the resources
	infos, err := o.kubectlApplyOptions.GetObjects()
	if err != nil {
		return err
	}

	// generate policies and append them to the resources
	var results []*resource.Info
	for _, info := range infos {
		results = append(results, info)
		obj := o.generatePropagationObject(info)
		gvk := obj.GetObjectKind().GroupVersionKind()
		mapping, err := o.kubectlApplyOptions.Mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			return fmt.Errorf("unable to recognize resource: %v", err)
		}
		client, err := o.KubectlApplyFlags.Factory.ClientForMapping(mapping)
		if err != nil {
			return fmt.Errorf("unable to connect to a server to handle %q: %v", mapping.Resource, err)
		}
		policyName, _ := metadataAccessor.Name(obj)
		ret := &resource.Info{
			Namespace: info.Namespace,
			Name:      policyName,
			Object:    obj,
			Mapping:   mapping,
			Client:    client,
		}
		results = append(results, ret)
	}

	// store the results object to be sequentially applied
	o.kubectlApplyOptions.SetObjects(results)
	return nil
}

// generatePropagationObject generates a propagation object for the given resource info.
// It takes the resource namespace, name and GVK as input to generate policy name.
// TODO(carlory): allow users to select one or many member clusters to propagate resources.
func (o *CommandApplyOptions) generatePropagationObject(info *resource.Info) runtime.Object {
	gvk := info.Mapping.GroupVersionKind
	spec := policyv1alpha1.PropagationSpec{
		ResourceSelectors: []policyv1alpha1.ResourceSelector{
			{
				APIVersion: gvk.GroupVersion().String(),
				Kind:       gvk.Kind,
				Name:       info.Name,
				Namespace:  info.Namespace,
			},
		},
	}

	if o.AllClusters {
		spec.Placement.ClusterAffinity = &policyv1alpha1.ClusterAffinity{}
	}

	// for a namespaced-scope resource, we need to generate a PropagationPolicy object.
	// for a cluster-scope resource, we need to generate a ClusterPropagationPolicy object.
	var obj runtime.Object
	policyName := names.GeneratePolicyName(info.Namespace, info.Name, gvk.String())
	if info.Namespaced() {
		obj = &policyv1alpha1.PropagationPolicy{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "policy.karmada.io/v1alpha1",
				Kind:       "PropagationPolicy",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      policyName,
				Namespace: info.Namespace,
			},
			Spec: spec,
		}
	} else {
		obj = &policyv1alpha1.ClusterPropagationPolicy{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "policy.karmada.io/v1alpha1",
				Kind:       "ClusterPropagationPolicy",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: policyName,
			},
			Spec: spec,
		}
	}
	return obj
}
