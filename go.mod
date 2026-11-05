module github.com/rh-ecosystem-edge/eco-goinfra

go 1.27.0

toolchain go1.27.3

require (
	github.com/Masterminds/semver/v3 v3.5.0
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2
	github.com/blang/semver/v4 v4.0.0
	github.com/containernetworking/cni v1.3.1
	github.com/go-logr/logr v1.4.4
	github.com/go-openapi/errors v0.22.9
	github.com/go-openapi/strfmt v0.27.4
	github.com/go-openapi/swag v0.29.2
	github.com/go-openapi/validate v1.0.1
	github.com/google/go-cmp v0.7.0
	github.com/google/uuid v1.6.0
	github.com/grafana/loki/operator/apis/loki v0.0.0-20241021105923-5e970e50b166
	github.com/hashicorp/vault/api v1.23.0
	github.com/hashicorp/vault/api/auth/approle v0.12.0
	github.com/hashicorp/vault/api/auth/kubernetes v0.12.0
	github.com/k8snetworkplumbingwg/multi-networkpolicy v1.0.1
	github.com/k8snetworkplumbingwg/network-attachment-definition-client v1.7.7
	github.com/k8snetworkplumbingwg/ptp-operator v0.0.0-20261105101233-5af7dd4d71de
	github.com/k8snetworkplumbingwg/sriov-network-operator v1.6.0
	github.com/kedacore/keda-olm-operator v0.0.0-20261103164440-8312a507b7f0
	github.com/kedacore/keda/v2 v2.21.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/kube-object-storage/lib-bucket-provisioner v0.0.0-20260917153727-be094b8c7ee9
	github.com/lib/pq v1.12.3
	github.com/metal3-io/baremetal-operator/apis v0.14.2
	github.com/nmstate/kubernetes-nmstate/api v0.0.0-20261103102427-4719031e11cd
	github.com/oapi-codegen/runtime v1.8.0
	github.com/onsi/ginkgo/v2 v2.34.0
	github.com/openshift-kni/cluster-group-upgrades-operator v0.0.0-20261105110240-c1eb317ad72e // release-5.0
	github.com/openshift-kni/lifecycle-agent v0.0.0-20261105164035-5d711fea4f14 // release-5.0
	github.com/openshift-kni/numaresources-operator v0.4.18-0.2024100201.0.20261105105914-b539a7601cd3 // release-5.0
	github.com/openshift-kni/oran-o2ims v0.0.0-20260830032435-b07f1e92f54d // release-5.0, aligned with k8s v0.36
	github.com/openshift/api v0.0.1
	github.com/openshift/client-go v0.0.1 // release-5.0
	github.com/openshift/cluster-logging-operator/api/observability v0.0.0-20261105045812-9b9d58ba71cd
	github.com/openshift/cluster-nfd-operator v0.0.0-20260901080303-34d2d3f4aefa // release-5.0
	github.com/openshift/cluster-node-tuning-operator v0.0.0-20261012224409-9e8095860b2d // release-5.0
	github.com/openshift/custom-resource-status v1.1.3-0.20220503160415-f2fdb4999d87
	github.com/openshift/elasticsearch-operator v0.0.0-20250923121540-138a709613fd // release-5.8
	github.com/openshift/local-storage-operator v0.0.0-20261002135716-3e6582a04c6b // release-5.0
	github.com/ovn-kubernetes/ovn-kubernetes/go-controller v0.0.0-20261105142609-aff0217cc4cc
	github.com/pkg/errors v0.9.1
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.93.1 // chosen to align with k8s v0.36
	github.com/red-hat-storage/odf-operator v0.0.0-20260930105922-effed015aa23 // release-4.21, api/v1alpha1 was deprecated in 4.19 and removed in 4.22
	github.com/sirupsen/logrus v1.10.2
	github.com/stmcginnis/gofish v0.27.0
	github.com/stretchr/testify v1.12.1
	github.com/thoas/go-funk v0.9.3
	github.com/vmware-tanzu/velero v1.18.4
	golang.org/x/crypto v0.58.0
	gopkg.in/k8snetworkplumbingwg/multus-cni.v4 v4.3.1 // aligned with k8s v0.36
	gopkg.in/yaml.v2 v2.4.0
	gorm.io/gorm v1.31.2
	k8s.io/api v0.36.5
	k8s.io/apiextensions-apiserver v0.36.5
	k8s.io/apimachinery v0.36.5
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/klog/v2 v2.140.0
	k8s.io/kubectl v0.36.5
	k8s.io/kubelet v0.36.5
	k8s.io/streaming v0.36.5
	k8s.io/utils v0.0.0-20261023123148-ca04ba58ecab
	maistra.io/api v0.0.0-20240319144440-ffa91c765143
	open-cluster-management.io/api v1.4.0
	open-cluster-management.io/governance-policy-propagator v0.20.1
	open-cluster-management.io/multicloud-operators-subscription v0.16.0
	sigs.k8s.io/container-object-storage-interface-api v0.1.0
	sigs.k8s.io/controller-runtime v0.24.1
	sigs.k8s.io/yaml v1.6.0
)

require (
	dario.cat/mergo v1.0.2 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20250102033503-faa5f7b0171c // indirect
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/sprig/v3 v3.3.0 // indirect
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
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
	github.com/evanphx/json-patch v5.9.11+incompatible // indirect
	github.com/evanphx/json-patch/v5 v5.9.11 // indirect
	github.com/exponent-io/jsonpath v0.0.0-20210407135951-1de76d718b3f // indirect
	github.com/expr-lang/expr v1.17.8 // indirect
	github.com/fsnotify/fsnotify v1.10.1 // indirect
	github.com/fxamacker/cbor/v2 v2.9.4 // indirect
	github.com/getkin/kin-openapi v0.147.0 // indirect
	github.com/ghodss/yaml v1.0.1-0.20220118164431-d8423dcdf344 // indirect
	github.com/go-errors/errors v1.5.1 // indirect
	github.com/go-jose/go-jose/v4 v4.1.5 // indirect
	github.com/go-openapi/analysis v1.0.1 // indirect
	github.com/go-openapi/jsonpointer v1.0.2 // indirect
	github.com/go-openapi/jsonreference v1.0.3 // indirect
	github.com/go-openapi/loads v0.25.3 // indirect
	github.com/go-openapi/spec v1.0.1 // indirect
	github.com/go-openapi/swag/cmdutils v0.29.2 // indirect
	github.com/go-openapi/swag/conv v0.29.2 // indirect
	github.com/go-openapi/swag/fileutils v0.29.2 // indirect
	github.com/go-openapi/swag/jsonutils v0.29.2 // indirect
	github.com/go-openapi/swag/loading v0.29.2 // indirect
	github.com/go-openapi/swag/mangling v0.29.2 // indirect
	github.com/go-openapi/swag/netutils v0.29.2 // indirect
	github.com/go-openapi/swag/pools v0.29.2 // indirect
	github.com/go-openapi/swag/stringutils v0.29.2 // indirect
	github.com/go-openapi/swag/typeutils v0.29.2 // indirect
	github.com/go-openapi/swag/yamlutils v0.29.2 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/google/gnostic-models v0.7.1 // indirect
	github.com/google/pprof v0.0.0-20260903180319-d6c3cb2f37ec // indirect
	github.com/gorilla/websocket v1.5.4-0.20250319132907-e064f32e3674 // indirect
	github.com/grafana/regexp v0.0.0-20250905093917-f7b3be9d1853 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.8 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/go-secure-stdlib/parseutil v0.2.0 // indirect
	github.com/hashicorp/go-secure-stdlib/strutil v0.1.2 // indirect
	github.com/hashicorp/go-sockaddr v1.0.7 // indirect
	github.com/hashicorp/hcl v1.0.1-vault-7 // indirect
	github.com/huandu/xstrings v1.6.2 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
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
	github.com/oapi-codegen/oapi-codegen/v2 v2.8.0 // indirect
	github.com/oasdiff/yaml v0.1.1 // indirect
	github.com/oasdiff/yaml3 v0.0.14 // indirect
	github.com/oklog/ulid/v2 v2.1.2 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_golang v1.24.1 // indirect
	github.com/prometheus/client_model v0.6.3 // indirect
	github.com/prometheus/common v1.20.99 // indirect
	github.com/prometheus/procfs v0.22.0 // indirect
	github.com/r3labs/diff/v3 v3.0.2 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.2 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	github.com/speakeasy-api/jsonpath v0.6.3 // indirect
	github.com/speakeasy-api/openapi v1.24.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/cobra v1.10.2 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/vincent-petithory/dataurl v1.0.0 // indirect
	github.com/vishvananda/netns v0.0.5 // indirect
	github.com/vmihailenco/msgpack/v5 v5.4.1 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/vmware-labs/yaml-jsonpath v0.3.2 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	github.com/xlab/treeprint v1.2.0 // indirect
	go.opentelemetry.io/otel v1.47.0 // indirect
	go.opentelemetry.io/otel/trace v1.47.0 // indirect
	go.yaml.in/yaml/v2 v2.4.4 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/mod v0.41.0 // indirect
	golang.org/x/net v0.62.0 // indirect
	golang.org/x/oauth2 v0.37.0 // indirect
	golang.org/x/sync v0.24.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/term v0.47.0 // indirect
	golang.org/x/text v0.43.0 // indirect
	golang.org/x/time v0.16.0 // indirect
	golang.org/x/tools v0.51.0 // indirect
	gomodules.xyz/jsonpatch/v2 v2.5.0 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.13.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/apiserver v0.36.5 // indirect
	k8s.io/cli-runtime v0.36.5 // indirect
	k8s.io/component-base v0.36.5 // indirect
	k8s.io/klog v1.0.0 // indirect
	k8s.io/kube-openapi v0.35.1 // indirect
	sigs.k8s.io/json v0.0.0-20250730193827-2d320260d730 // indirect
	sigs.k8s.io/kustomize/api v0.21.1 // indirect
	sigs.k8s.io/kustomize/kyaml v0.21.1 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v6 v6.4.2 // indirect
)

replace (
	github.com/imdario/mergo => github.com/imdario/mergo v0.3.16
	github.com/k8snetworkplumbingwg/ptp-operator => github.com/openshift/ptp-operator v0.0.0-20261023023103-cbe73c6129a9 // release-5.0
	github.com/k8snetworkplumbingwg/sriov-network-operator => github.com/openshift/sriov-network-operator v0.0.0-20260923181747-4d6a7b26cbe2 // release-5.0
	// PTP declares legacy v0.0.1 versions that outrank its compatible timestamped releases.
	github.com/openshift/api => github.com/openshift/api v0.0.0-20261015124049-70586205d5fa // release-5.0
	github.com/openshift/client-go => github.com/openshift/client-go v0.0.0-20261015075541-6446fecb0c88 // release-5.0
	k8s.io/client-go => k8s.io/client-go v0.36.5
	// The cluster-node-tuning-operator release-5.0 uses version k8s.io/kube-openapi v0.35.1, which does not exist.
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20260821135717-be32def86098 // master
)

tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen
