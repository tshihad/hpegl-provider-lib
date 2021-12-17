// (C) Copyright 2021 Hewlett Packard Enterprise Development LP

package atf

const providerStanza = `
	provider hpegl {
		vmaas {}
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
