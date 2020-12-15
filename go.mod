module github.com/llomgui/terraform-provider-openshift

require (
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.2.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/openshift/api v0.0.0-20200930075302-db52bc4ef99f
	github.com/openshift/client-go v0.0.0-20200929181438-91d71ef2122c
	golang.org/x/tools v0.0.0-20201202100533-7534955ac86b // indirect
	k8s.io/api v0.19.1
	k8s.io/apimachinery v0.19.1
	k8s.io/client-go v11.0.0+incompatible
)

replace k8s.io/client-go => k8s.io/client-go v0.19.1

go 1.15
