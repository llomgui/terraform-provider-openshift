module github.com/llomgui/terraform-provider-openshift

require (
	github.com/gogo/protobuf v1.2.1 // indirect
	github.com/hashicorp/terraform-plugin-sdk v1.10.0
	github.com/mattn/go-colorable v0.1.4 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/openshift/api v3.9.1-0.20190322043348-8741ff068a47+incompatible
	github.com/openshift/client-go v0.0.0-20180830153425-431ec9a26e50
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.4.0 // indirect
	github.com/terraform-providers/terraform-provider-random v0.0.0-20190925200408-30dac3233094
	golang.org/x/crypto v0.0.0-20191011191535-87dc89f01550 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	k8s.io/api v0.0.0-20200131232428-e3a917c59b04
	k8s.io/apimachinery v0.0.0-20200409202947-6e7c4b1e1854
	k8s.io/client-go v12.0.0+incompatible
)

replace k8s.io/client-go => k8s.io/client-go v0.0.0-20200410023015-75e09fce8f36

go 1.14
