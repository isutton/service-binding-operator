module github.com/redhat-developer/service-binding-operator

go 1.13

require (
	contrib.go.opencensus.io/exporter/ocagent v0.4.11 // indirect
	contrib.go.opencensus.io/exporter/prometheus v0.1.0 // indirect
	contrib.go.opencensus.io/exporter/stackdriver v0.12.8 // indirect
	github.com/Azure/go-autorest v11.7.0+incompatible // indirect
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/coreos/etcd-operator v0.9.4
	github.com/emicklei/go-restful v2.9.6+incompatible // indirect
	github.com/go-logr/logr v0.1.0
	github.com/go-openapi/spec v0.19.0
	github.com/go-openapi/swag v0.19.0 // indirect
	github.com/google/go-containerregistry v0.0.0-20191218175032-34fb8ff33bed // indirect
	github.com/google/uuid v1.1.1 // indirect
	github.com/mailru/easyjson v0.0.0-20190403194419-1ea4449da983 // indirect
	github.com/openshift/api v3.9.1-0.20190424152011-77b8897ec79a+incompatible
	github.com/openshift/custom-resource-status v0.0.0-20190822192428-e62f2f3b79f3
	github.com/operator-backing-service-samples/postgresql-operator v0.0.0-20191023140509-5c3697ed3069
	github.com/operator-framework/operator-lifecycle-manager v0.0.0-20190605231540-b8a4faf68e36
	github.com/operator-framework/operator-sdk v0.12.0
	github.com/pmezard/go-difflib v1.0.0
	github.com/spf13/pflag v1.0.3
	github.com/stretchr/testify v1.3.0
	golang.org/x/build v0.0.0-20190314133821-5284462c4bec // indirect
	golang.org/x/crypto v0.0.0-20190701094942-4def268fd1a4 // indirect
	gopkg.in/yaml.v2 v2.2.2
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.0.0
	k8s.io/apiextensions-apiserver v0.0.0
	k8s.io/apimachinery v0.0.0
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/code-generator v0.0.0
	k8s.io/gengo v0.0.0-20191010091904-7fa3014cb28f
	k8s.io/kube-openapi v0.0.0-20190918143330-0270cf2f1c1d
	knative.dev/pkg v0.0.0-20191221032535-9fda5bd59a67 // indirect
	knative.dev/serving v0.9.0
	sigs.k8s.io/controller-runtime v0.3.0
	sigs.k8s.io/controller-tools v0.2.2
)

// Pinned to kubernetes-1.15.4
replace (
	github.com/openshift/api => github.com/openshift/api v0.0.0-20190424152011-77b8897ec79a
	k8s.io/api => k8s.io/api v0.0.0-20190918195907-bd6ac527cfd2
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190918201827-3de75813f604
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190817020851-f2f3a405f61d
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20190918200908-1e17798da8c1
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20190918202139-0b14c719ca62
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190918200256-06eb1244587a
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20190918203125-ae665f80358a
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.0.0-20190918202959-c340507a5d48
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20190612205613-18da4a14b22b
	k8s.io/component-base => k8s.io/component-base v0.0.0-20190918200425-ed2f0867c778
	k8s.io/cri-api => k8s.io/cri-api v0.0.0-20190817025403-3ae76f584e79
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.0.0-20190918203248-97c07dcbb623
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20190918201136-c3a845f1fbb2
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.0.0-20190918202837-c54ce30c680e
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.0.0-20190918202429-08c8357f8e2d
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.0.0-20190918202713-c34a54b3ec8e
	k8s.io/kubelet => k8s.io/kubelet v0.0.0-20190918202550-958285cf3eef
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.0.0-20190918203421-225f0541b3ea
	k8s.io/metrics => k8s.io/metrics v0.0.0-20190918202012-3c1ca76f5bda
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.0.0-20190918201353-5cc279503896
)
