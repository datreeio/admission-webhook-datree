package enums

const (
	ClusterName    = "CLUSTER_NAME"
	Token          = "DATREE_TOKEN"
	ClientId       = "DATREE_CLIENT_ID"
	Policy         = "DATREE_POLICY"
	Verbose        = "DATREE_VERBOSE"
	NoRecord       = "DATREE_NO_RECORD"
	Output         = "DATREE_OUTPUT"
	Enforce        = "DATREE_ENFORCE"
	ConfigFromHelm = "DATREE_CONFIG_FROM_HELM"
	Namespace      = "DATREE_NAMESPACE"
	PodName        = "POD_NAME"
)

type ActionOnFailure string

const (
	EnforceActionOnFailure ActionOnFailure = "enforce"
	MonitorActionOnFailure ActionOnFailure = "monitor"
)
