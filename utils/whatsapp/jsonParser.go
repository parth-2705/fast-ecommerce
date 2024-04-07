package whatsapp

import (
	"encoding/json"
	"fmt"
	"hermes/utils"
	"os"
)

func ParseJSONForMPM(jsonFile string) (template []TemplateComponent, err error) {
	folder := os.Getenv("INTERACTIVE_MESSAGES_FOLDER_PATH") + "mpm/"
	json, err := utils.ReadJSON(folder + jsonFile + ".json")
	if err != nil {
		return
	}
	template, err = parseJSONHelper(json)
	return
}

func parseJSONHelper(jsonByte []byte) (template []TemplateComponent, err error) {
	var interfaceMap map[string]interface{}
	json.Unmarshal(jsonByte, &interfaceMap)
	header, body, footer, mpm := parseComponents(interfaceMap)
	if header != nil {
		template = append(template, TemplateComponent{
			Type:       "header",
			Parameters: header,
		})
	}
	if body != nil {
		template = append(template, TemplateComponent{
			Type:       "body",
			Parameters: body,
		})
	}
	if footer != nil {
		template = append(template, TemplateComponent{
			Type:       "footer",
			Parameters: footer,
		})
	}
	if mpm != nil {
		template = append(template, TemplateComponent{
			Type:       "button",
			SubType:    "mpm",
			Index:      0,
			Parameters: mpm,
		})
	}
	return
}

func parseComponents(interfaceMap map[string]interface{}) (header []TemplateParameter, body []TemplateParameter, footer []TemplateParameter, mpm []TemplateParameter) {
	if headerTemp, okHeader := interfaceMap["header"]; okHeader {
		tempArr := headerTemp.([]interface{})
		for _, item := range tempArr {
			header = append(header, mapStringInterfaceToTemplateParameter(item))
		}
	} else {
		header = nil
	}
	if bodyTemp, okBody := interfaceMap["body"]; okBody {
		tempArr := bodyTemp.([]interface{})
		for _, item := range tempArr {
			body = append(body, mapStringInterfaceToTemplateParameter(item))
		}
	} else {
		body = nil
	}
	if footerTemp, okFooter := interfaceMap["footer"]; okFooter {
		tempArr := footerTemp.([]interface{})
		for _, item := range tempArr {
			footer = append(footer, mapStringInterfaceToTemplateParameter(item))
		}
	} else {
		footer = nil
	}
	if mpmTemp, okMPM := interfaceMap["mpm"]; okMPM {
		tempArr := mpmTemp.([]interface{})
		for _, item := range tempArr {
			mpm = append(mpm, mapStringInterfaceToTemplateParameter(item))
		}
	} else {
		mpm = nil
	}
	return
}

func mapStringInterfaceToTemplateParameter(interfaceTemplateParamter interface{}) (templateParameter TemplateParameter) {
	byteParam, err := json.Marshal(interfaceTemplateParamter)
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}
	json.Unmarshal(byteParam, &templateParameter)
	return
}

