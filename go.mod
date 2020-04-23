module github.com/llomgui/terraform-provider-openshift

require (
	github.com/hashicorp/terraform-plugin-sdk v1.10.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/openshift/api v3.9.1-0.20190322043348-8741ff068a47+incompatible
	github.com/openshift/client-go v0.0.0-20180830153425-431ec9a26e50
	github.com/terraform-providers/terraform-provider-random v0.0.0-20190925200408-30dac3233094
	k8s.io/api v0.0.0-20200131232428-e3a917c59b04
	k8s.io/apimachinery v0.0.0-20200409202947-6e7c4b1e1854
	k8s.io/client-go v12.0.0+incompatible
)

replace k8s.io/client-go => k8s.io/client-go v0.0.0-20200410023015-75e09fce8f36

go 1.14
