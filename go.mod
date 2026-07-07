module github.com/rh-ecosystem-edge/eco-goinfra

go 1.26.0

toolchain go1.26.4

require (
	github.com/Masterminds/semver/v3 v3.5.0
	github.com/blang/semver/v4 v4.0.0
	github.com/containernetworking/cni v1.3.0
	github.com/go-openapi/errors v0.22.8
	github.com/go-openapi/strfmt v0.26.4
	github.com/go-openapi/swag v0.27.0
	github.com/go-openapi/validate v0.26.0
	github.com/google/go-cmp v0.7.0
	github.com/google/uuid v1.6.0
	github.com/hashicorp/vault/api v1.23.0
	github.com/hashicorp/vault/api/auth/approle v0.12.0
	github.com/hashicorp/vault/api/auth/kubernetes v0.12.0
	github.com/k8snetworkplumbingwg/multi-networkpolicy v1.0.1
	github.com/k8snetworkplumbingwg/network-attachment-definition-client v1.7.7
	github.com/k8snetworkplumbingwg/sriov-network-operator v1.6.0
	github.com/kedacore/keda-olm-operator v0.0.0-20260618141108-6814218d455e // aligned with k8s v0.35
	github.com/kedacore/keda/v2 v2.19.0 // aligned with k8s v0.35
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/kube-object-storage/lib-bucket-provisioner v0.0.0-20260420161730-5164e3746489
	github.com/lib/pq v1.12.3
	github.com/metal3-io/baremetal-operator/apis v0.13.1
	github.com/nmstate/kubernetes-nmstate/api v0.0.0-20260707144101-8853341855d6
	github.com/onsi/ginkgo/v2 v2.32.0
	github.com/openshift-kni/cluster-group-upgrades-operator v0.0.0-20260707161822-9b9043d4494b // release-4.22
	github.com/openshift-kni/lifecycle-agent v0.0.0-20260707161814-ec769a366476 // release-4.22
	github.com/openshift-kni/numaresources-operator v0.4.18-0.2024100201.0.20260707092512-254b162fcd0f // release-4.22
	github.com/openshift-kni/oran-o2ims v0.0.0-20260707122918-22ed0a55833b // release-4.22
	github.com/openshift/api v0.0.0-20260521125114-09730f85d883 // release-4.22
	github.com/openshift/client-go v0.0.0-20260330134249-7e1499aaacd7 // release-4.22
	github.com/openshift/cluster-logging-operator/api/observability v0.0.0-20260623121619-2db215f31af4
	github.com/openshift/cluster-nfd-operator v0.0.0-20260629131115-e53505ffcb61 // release-4.22
	github.com/openshift/cluster-node-tuning-operator v0.0.0-20260209053755-f5fe4460e852 // release-4.22, prior to controller-runtime v0.23
	github.com/openshift/custom-resource-status v1.1.3-0.20220503160415-f2fdb4999d87
	github.com/openshift/elasticsearch-operator v0.0.0-20250923121540-138a709613fd // release-5.8
	github.com/openshift/local-storage-operator v0.0.0-20260630133617-4d174e9c9eff // release-4.22
	github.com/ovn-kubernetes/ovn-kubernetes/go-controller v0.0.0-20260707145430-b93b6a72bc15
	github.com/pkg/errors v0.9.1
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.91.0 // aligned with k8s v0.35
	github.com/red-hat-storage/odf-operator v0.0.0-20260226164309-08c71191d483 // release-4.21, api/v1alpha1 was deprecated in 4.19 and removed in 4.22
	github.com/sirupsen/logrus v1.9.4
	github.com/stmcginnis/gofish v0.20.0 // v0.21.0 contains many breaking changes. Should be upgraded separately.
	github.com/stretchr/testify v1.11.1
	github.com/thoas/go-funk v0.9.3
	golang.org/x/crypto v0.53.0
	golang.org/x/exp v0.0.0-20260611194520-c48552f49976
	gopkg.in/k8snetworkplumbingwg/multus-cni.v4 v4.3.0
	gopkg.in/yaml.v2 v2.4.0
	gorm.io/gorm v1.31.2
	k8s.io/api v0.35.6
	k8s.io/apiextensions-apiserver v0.35.6
	k8s.io/apimachinery v0.35.6
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/klog/v2 v2.140.0
	k8s.io/kubectl v0.35.6
	k8s.io/kubelet v0.35.6
	k8s.io/utils v0.0.0-20260707023825-cf1189d6abe3
	maistra.io/api v0.0.0-20240319144440-ffa91c765143
	open-cluster-management.io/api v1.3.0
	open-cluster-management.io/governance-policy-propagator v0.18.1-0.20260302212915-815d063a291a // prior to controller-runtime v0.23
	open-cluster-management.io/multicloud-operators-subscription v0.16.0
	sigs.k8s.io/container-object-storage-interface-api v0.1.0
	sigs.k8s.io/controller-runtime v0.23.3
	sigs.k8s.io/yaml v1.6.0
)

require (
	dario.cat/mergo v1.0.2 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20250102033503-faa5f7b0171c // indirect
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/sprig/v3 v3.3.0 // indirect
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/chai2010/gettext-go v1.0.2 // indirect
	github.com/coreos/fcct v0.5.0 // indirect
	github.com/coreos/go-json v0.0.0-20230131223807-18775e0fb4fb // indirect
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/coreos/go-systemd/v22 v22.7.0 // indirect
	github.com/coreos/ignition/v2 v2.26.0 // indirect
	github.com/coreos/vcontext v0.0.0-20231102161604-685dc7299dc5 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dprotaso/go-yit v0.0.0-20240618133044-5a0af90af097 // indirect
	github.com/emicklei/go-restful/v3 v3.13.0 // indirect
	github.com/evanphx/json-patch/v5 v5.9.11 // indirect
	github.com/exponent-io/jsonpath v0.0.0-20210407135951-1de76d718b3f // indirect
	github.com/expr-lang/expr v1.17.7 // indirect
	github.com/fatih/color v1.19.0 // indirect
	github.com/fsnotify/fsnotify v1.10.1 // indirect
	github.com/fxamacker/cbor/v2 v2.9.2 // indirect
	github.com/getkin/kin-openapi v0.140.0 // indirect
	github.com/ghodss/yaml v1.0.1-0.20220118164431-d8423dcdf344 // indirect
	github.com/go-errors/errors v1.5.1 // indirect
	github.com/go-jose/go-jose/v4 v4.1.4 // indirect
	github.com/go-logr/logr v1.4.3
	github.com/go-openapi/analysis v0.25.3 // indirect
	github.com/go-openapi/jsonpointer v0.24.0 // indirect
	github.com/go-openapi/jsonreference v0.21.6 // indirect
	github.com/go-openapi/loads v0.24.0 // indirect
	github.com/go-openapi/spec v0.22.6 // indirect
	github.com/go-openapi/swag/cmdutils v0.27.0 // indirect
	github.com/go-openapi/swag/conv v0.27.0 // indirect
	github.com/go-openapi/swag/fileutils v0.27.0 // indirect
	github.com/go-openapi/swag/jsonname v0.27.0 // indirect
	github.com/go-openapi/swag/jsonutils v0.27.0 // indirect
	github.com/go-openapi/swag/loading v0.27.0 // indirect
	github.com/go-openapi/swag/mangling v0.27.0 // indirect
	github.com/go-openapi/swag/netutils v0.27.0 // indirect
	github.com/go-openapi/swag/stringutils v0.27.0 // indirect
	github.com/go-openapi/swag/typeutils v0.27.0 // indirect
	github.com/go-openapi/swag/yamlutils v0.27.0 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/google/gnostic-models v0.7.1 // indirect
	github.com/google/pprof v0.0.0-20260604005048-7023385849c0 // indirect
	github.com/gorilla/websocket v1.5.4-0.20250319132907-e064f32e3674 // indirect
	github.com/grafana/loki/operator/apis/loki v0.0.0-20241021105923-5e970e50b166
	github.com/grafana/regexp v0.0.0-20250905093917-f7b3be9d1853 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.8 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/go-secure-stdlib/parseutil v0.2.0 // indirect
	github.com/hashicorp/go-secure-stdlib/strutil v0.1.2 // indirect
	github.com/hashicorp/go-sockaddr v1.0.7 // indirect
	github.com/hashicorp/hcl v1.0.1-vault-7 // indirect
	github.com/huandu/xstrings v1.5.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
	github.com/mattn/go-colorable v0.1.15 // indirect
	github.com/mattn/go-isatty v0.0.22 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/spdystream v0.5.1 // indirect
	github.com/moby/term v0.5.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/mxk/go-flowrate v0.0.0-20140419014527-cca7078d478f // indirect
	github.com/oapi-codegen/oapi-codegen/v2 v2.7.1 // indirect
	github.com/oapi-codegen/runtime v1.4.2
	github.com/oasdiff/yaml v0.1.0 // indirect
	github.com/oasdiff/yaml3 v0.0.13 // indirect
	github.com/oklog/ulid/v2 v2.1.1 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_golang v1.23.2 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v1.20.99 // indirect
	github.com/prometheus/procfs v0.20.1 // indirect
	github.com/r3labs/diff/v3 v3.0.2 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.2 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	github.com/speakeasy-api/jsonpath v0.6.3 // indirect
	github.com/speakeasy-api/openapi v1.19.2 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/cobra v1.10.2 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/vincent-petithory/dataurl v1.0.0 // indirect
	github.com/vishvananda/netns v0.0.5 // indirect
	github.com/vmihailenco/msgpack/v5 v5.4.1 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/vmware-labs/yaml-jsonpath v0.3.2 // indirect
	github.com/vmware-tanzu/velero v1.18.0
	github.com/x448/float16 v0.8.4 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	github.com/xlab/treeprint v1.2.0 // indirect
	go.opentelemetry.io/otel v1.44.0 // indirect
	go.opentelemetry.io/otel/trace v1.44.0 // indirect
	go.yaml.in/yaml/v2 v2.4.4 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/mod v0.37.0 // indirect
	golang.org/x/net v0.56.0 // indirect
	golang.org/x/oauth2 v0.36.0 // indirect
	golang.org/x/sync v0.21.0 // indirect
	golang.org/x/sys v0.46.0 // indirect
	golang.org/x/term v0.44.0 // indirect
	golang.org/x/text v0.39.0 // indirect
	golang.org/x/time v0.15.0 // indirect
	golang.org/x/tools v0.47.0 // indirect
	gomodules.xyz/jsonpatch/v2 v2.5.0 // indirect
	google.golang.org/protobuf v1.36.12-0.20260120151049-f2248ac996af // indirect
	gopkg.in/evanphx/json-patch.v4 v4.13.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/apiserver v0.35.6 // indirect
	k8s.io/cli-runtime v0.35.6 // indirect
	k8s.io/component-base v0.35.6 // indirect
	k8s.io/klog v1.0.0 // indirect
	k8s.io/kube-openapi v0.35.1 // indirect
	knative.dev/pkg v0.0.0-20260120122510-4a022ed9999a // indirect
	sigs.k8s.io/json v0.0.0-20250730193827-2d320260d730 // indirect
	sigs.k8s.io/kustomize/api v0.21.1 // indirect
	sigs.k8s.io/kustomize/kyaml v0.21.1 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v6 v6.4.1 // indirect
)

replace (
	github.com/imdario/mergo => github.com/imdario/mergo v0.3.16
	github.com/k8snetworkplumbingwg/sriov-network-operator => github.com/openshift/sriov-network-operator v0.0.0-20260526181104-0626dd1a7086 // release-4.22
	k8s.io/client-go => k8s.io/client-go v0.35.6
	// The cluster-node-tuning-operator release-4.22 uses version k8s.io/kube-openapi v0.35.1, which does not exist.
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20260127142750-a19766b6e2d4
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.22.5
)

tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen
