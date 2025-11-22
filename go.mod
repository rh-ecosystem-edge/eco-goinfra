module github.com/rh-ecosystem-edge/eco-goinfra

go 1.25

toolchain go1.25.4

require (
	github.com/Masterminds/semver/v3 v3.4.0
	github.com/blang/semver/v4 v4.0.0
	github.com/containernetworking/cni v1.3.0
	github.com/go-openapi/errors v0.22.4
	github.com/go-openapi/strfmt v0.25.0
	github.com/go-openapi/swag v0.25.3
	github.com/go-openapi/validate v0.25.1
	github.com/google/go-cmp v0.7.0
	github.com/google/uuid v1.6.0
	github.com/hashicorp/vault/api v1.22.0
	github.com/hashicorp/vault/api/auth/approle v0.11.0
	github.com/hashicorp/vault/api/auth/kubernetes v0.10.0
	github.com/k8snetworkplumbingwg/multi-networkpolicy v1.0.1
	github.com/k8snetworkplumbingwg/network-attachment-definition-client v1.7.7
	github.com/k8snetworkplumbingwg/sriov-network-operator v1.6.0
	github.com/kedacore/keda-olm-operator v0.0.0-20251121013310-215d3d336056
	github.com/kedacore/keda/v2 v2.18.1
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/kube-object-storage/lib-bucket-provisioner v0.0.0-20221122204822-d1a8c34382f1
	github.com/lib/pq v1.10.9
	github.com/metal3-io/baremetal-operator/apis v0.11.2
	github.com/nmstate/kubernetes-nmstate/api v0.0.0-20251121083040-ba2d4c63597f
	github.com/onsi/ginkgo/v2 v2.27.2
	github.com/openshift-kni/cluster-group-upgrades-operator v0.0.0-20250914044906-f2a5c9f05a5b // f2a5c9f05a5b0ecb3a58c1ae2168e55dbe98f81b
	github.com/openshift-kni/lifecycle-agent v0.0.0-20251121192321-6a1e18fedd81 // release-4.20
	github.com/openshift-kni/numaresources-operator v0.4.18-0.2024100201.0.20251122094151-eb6415170827 // release-4.20
	github.com/openshift-kni/oran-o2ims/api/hardwaremanagement v0.0.0-20250826145425-f200f7e70e0a // f200f7e70e0acecea284d5b5a507393fdf3c5f70
	github.com/openshift-kni/oran-o2ims/api/provisioning v0.0.0-20250826145425-f200f7e70e0a // f200f7e70e0acecea284d5b5a507393fdf3c5f70
	github.com/openshift/api v0.0.0-20251114171455-1886180ef430 // release-4.20
	github.com/openshift/client-go v0.0.0-20250811163556-6193816ae379 // release-4.20
	github.com/openshift/cluster-nfd-operator v0.0.0-20250929121503-98a074e63cd0 // release-4.20
	github.com/openshift/cluster-node-tuning-operator v0.0.0-20251108153041-2ed182ba5710 // release-4.20
	github.com/openshift/custom-resource-status v1.1.3-0.20220503160415-f2fdb4999d87
	github.com/openshift/elasticsearch-operator v0.0.0-20241202223819-cc1a232913d6 // release-5.8
	github.com/openshift/local-storage-operator v0.0.0-20251006201529-b394d7760c51 // release-4.20
	github.com/ovn-org/ovn-kubernetes/go-controller v0.0.0-20250901081027-bc4f9b80d20d // bc4f9b80d20d2d584b03c6a0f24835c73833c65d
	github.com/pkg/errors v0.9.1
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.85.0 // aligned with k8s v0.33
	github.com/red-hat-storage/odf-operator v0.0.0-20251121081437-2093de9a06d1 // release-4.20
	github.com/sirupsen/logrus v1.9.3
	github.com/stmcginnis/gofish v0.20.0
	github.com/stretchr/testify v1.11.1
	github.com/thoas/go-funk v0.9.3
	golang.org/x/crypto v0.45.0
	golang.org/x/exp v0.0.0-20251113190631-e25ba8c21ef6
	gopkg.in/k8snetworkplumbingwg/multus-cni.v4 v4.2.3
	gopkg.in/yaml.v2 v2.4.0
	gorm.io/gorm v1.31.1
	k8s.io/api v0.33.6
	k8s.io/apiextensions-apiserver v0.33.6
	k8s.io/apimachinery v0.33.6
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/klog/v2 v2.130.1
	k8s.io/kubectl v0.33.6
	k8s.io/kubelet v0.33.6
	k8s.io/utils v0.0.0-20251002143259-bc988d571ff4
	maistra.io/api v0.0.0-20240319144440-ffa91c765143
	open-cluster-management.io/api v1.1.0
	open-cluster-management.io/governance-policy-propagator v0.16.0
	open-cluster-management.io/multicloud-operators-subscription v0.16.0
	sigs.k8s.io/container-object-storage-interface-api v0.1.0
	sigs.k8s.io/controller-runtime v0.22.3
	sigs.k8s.io/yaml v1.6.0
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20230124172434-306776ec8161 // indirect
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/sprig/v3 v3.2.3 // indirect
	github.com/ajeddeloh/go-json v0.0.0-20200220154158-5ae607161559 // indirect
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2
	github.com/aws/aws-sdk-go-v2 v1.39.2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/chai2010/gettext-go v1.0.2 // indirect
	github.com/clarketm/json v1.17.1 // indirect
	github.com/coreos/fcct v0.5.0 // indirect
	github.com/coreos/go-json v0.0.0-20230131223807-18775e0fb4fb // indirect
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf // indirect
	github.com/coreos/go-systemd/v22 v22.5.0 // indirect
	github.com/coreos/ign-converter v0.0.0-20230417193809-cee89ea7d8ff // indirect
	github.com/coreos/ignition v0.35.0 // indirect
	github.com/coreos/ignition/v2 v2.23.0 // indirect
	github.com/coreos/vcontext v0.0.0-20231102161604-685dc7299dc5 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dprotaso/go-yit v0.0.0-20220510233725-9ba8df137936 // indirect
	github.com/emicklei/go-restful/v3 v3.13.0 // indirect
	github.com/evanphx/json-patch/v5 v5.9.11 // indirect
	github.com/exponent-io/jsonpath v0.0.0-20210407135951-1de76d718b3f // indirect
	github.com/expr-lang/expr v1.17.6 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/fxamacker/cbor/v2 v2.9.0 // indirect
	github.com/getkin/kin-openapi v0.127.0 // indirect
	github.com/ghodss/yaml v1.0.1-0.20220118164431-d8423dcdf344 // indirect
	github.com/go-errors/errors v1.5.1 // indirect
	github.com/go-jose/go-jose/v4 v4.1.3 // indirect
	github.com/go-logr/logr v1.4.3
	github.com/go-openapi/analysis v0.24.1 // indirect
	github.com/go-openapi/jsonpointer v0.22.3 // indirect
	github.com/go-openapi/jsonreference v0.21.3 // indirect
	github.com/go-openapi/loads v0.23.2 // indirect
	github.com/go-openapi/spec v0.22.1 // indirect
	github.com/go-openapi/swag/cmdutils v0.25.3 // indirect
	github.com/go-openapi/swag/conv v0.25.3 // indirect
	github.com/go-openapi/swag/fileutils v0.25.3 // indirect
	github.com/go-openapi/swag/jsonname v0.25.3 // indirect
	github.com/go-openapi/swag/jsonutils v0.25.3 // indirect
	github.com/go-openapi/swag/loading v0.25.3 // indirect
	github.com/go-openapi/swag/mangling v0.25.3 // indirect
	github.com/go-openapi/swag/netutils v0.25.3 // indirect
	github.com/go-openapi/swag/stringutils v0.25.3 // indirect
	github.com/go-openapi/swag/typeutils v0.25.3 // indirect
	github.com/go-openapi/swag/yamlutils v0.25.3 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/google/gnostic-models v0.7.1 // indirect
	github.com/google/pprof v0.0.0-20251114195745-4902fdda35c8 // indirect
	github.com/gorilla/websocket v1.5.4-0.20250319132907-e064f32e3674 // indirect
	github.com/grafana/loki/operator/apis/loki v0.0.0-20241021105923-5e970e50b166
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
	github.com/huandu/xstrings v1.4.0 // indirect
	github.com/imdario/mergo v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/invopop/yaml v0.3.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
	github.com/mailru/easyjson v0.9.0 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/spdystream v0.5.0 // indirect
	github.com/moby/term v0.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/mxk/go-flowrate v0.0.0-20140419014527-cca7078d478f // indirect
	github.com/oapi-codegen/oapi-codegen/v2 v2.4.1 // indirect
	github.com/oapi-codegen/runtime v1.1.1
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/openshift-kni/oran-o2ims/api/common v0.0.0-20250728092029-9ec1477b18f0 // indirect
	github.com/openshift/cluster-logging-operator/api/observability v0.0.0-20250422180113-5bae4ccfc5ef
	github.com/openshift/library-go v0.0.0-20251110200504-2685cf1242fc // indirect
	github.com/openshift/machine-config-operator v0.0.1-0.20250320230514-53e78f3692ee // indirect
	github.com/perimeterx/marshmallow v1.1.5 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_golang v1.23.2 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.67.3 // indirect
	github.com/prometheus/procfs v0.19.2 // indirect
	github.com/r3labs/diff/v3 v3.0.2 // indirect
	github.com/robfig/cron v1.2.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	github.com/speakeasy-api/openapi-overlay v0.9.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/cobra v1.10.1 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/vincent-petithory/dataurl v1.0.0 // indirect
	github.com/vishvananda/netns v0.0.5 // indirect
	github.com/vmihailenco/msgpack/v5 v5.4.1 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/vmware-labs/yaml-jsonpath v0.3.2 // indirect
	github.com/vmware-tanzu/velero v1.16.2
	github.com/x448/float16 v0.8.4 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	github.com/xlab/treeprint v1.2.0 // indirect
	go.mongodb.org/mongo-driver v1.17.6 // indirect
	go.opentelemetry.io/otel v1.37.0 // indirect
	go.opentelemetry.io/otel/trace v1.37.0 // indirect
	go.yaml.in/yaml/v2 v2.4.3 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	go4.org v0.0.0-20230225012048-214862532bf5 // indirect
	golang.org/x/mod v0.30.0 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/oauth2 v0.33.0 // indirect
	golang.org/x/sync v0.18.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/term v0.37.0 // indirect
	golang.org/x/text v0.31.0 // indirect
	golang.org/x/time v0.14.0 // indirect
	golang.org/x/tools v0.39.0 // indirect
	gomodules.xyz/jsonpatch/v2 v2.5.0 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.13.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/apiserver v0.33.6 // indirect
	k8s.io/cli-runtime v0.33.6 // indirect
	k8s.io/component-base v0.33.6 // indirect
	k8s.io/klog v1.0.0 // indirect
	k8s.io/kube-aggregator v0.33.5 // indirect
	k8s.io/kube-openapi v0.0.0-20250701173324-9bd5c66d9911 // indirect
	knative.dev/pkg v0.0.0-20250326102644-9f3e60a9244c // indirect
	sigs.k8s.io/json v0.0.0-20250730193827-2d320260d730 // indirect
	sigs.k8s.io/kube-storage-version-migrator v0.0.6-0.20230721195810-5c8923c5ff96 // indirect
	sigs.k8s.io/kustomize/api v0.21.0 // indirect
	sigs.k8s.io/kustomize/kyaml v0.21.0 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.7.0 // indirect
)

replace (
	github.com/imdario/mergo => github.com/imdario/mergo v0.3.16
	github.com/k8snetworkplumbingwg/sriov-network-operator => github.com/openshift/sriov-network-operator v0.0.0-20251006174000-8767df23a420 // release-4.20
	k8s.io/client-go => k8s.io/client-go v0.33.6
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.19.7
)

tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen
