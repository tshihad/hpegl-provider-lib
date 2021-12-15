package atf

const providerStanza = `
	provider hpegl {
		vmaas {}
		caas{}
		bmaas{}
	}
`

var accTestPath = "../../acc-testcases"

const (
	accKey  = "acc"
	jsonKey = "json"
	tfKey   = "tf"

	randMaxLimit   = 9999999
	testFuncPrefix = "Test"
	AccTestPathKey = "TF_ACC_TEST_PATH"
)
