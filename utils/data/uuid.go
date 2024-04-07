package data

import "github.com/google/uuid"

func GetUUIDString(typePrefix string) string {
	return typePrefix + "-" + uuid.New().String()
}

func GetUUIDStringWithoutPrefix() string {
	return uuid.New().String()
}

func GetNDRWorkflowID() string {
	return "NDRWorkflow_1"
}
