module github.com/llomgui/terraform-provider-openshift

require (
	github.com/client9/misspell v0.3.4
	github.com/golangci/golangci-lint v1.23.1
	github.com/hashicorp/terraform-plugin-sdk v1.5.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/openshift/api v3.9.1-0.20190322043348-8741ff068a47+incompatible
	github.com/openshift/client-go v0.0.0-20180830153425-431ec9a26e50
	github.com/terraform-providers/terraform-provider-random v0.0.0-20190925200408-30dac3233094
	k8s.io/api v0.0.0-20190620084959-7cf5895f2711
	k8s.io/apimachinery v0.0.0-20190612205821-1799e75a0719
	k8s.io/client-go v12.0.0+incompatible
)

go 1.13
