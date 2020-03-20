package main

import (
	"regexp"
	"strings"
)

type HomeTaxConfig struct {
	State struct {
		Date         string `json:"date" firestore:"date"`
		Close        string `json:"close" firestore:"close"`
		Down         string `json:"down" firestore:"down"`
		Unregistered string `json:"unregistered" firestore:"unregistered"`
	} `json:"state" firestore:"state"`

	TaxType struct {
		Date      string `json:"date" firestore:"date"`
		Free      string `json:"free" firestore:"free"`
		NonProfit string `json:"nonProfit" firestore:"nonProfit"`
		Normal    string `json:"normal" firestore:"normal"`
		Simple    string `json:"simple" firestore:"simple"`
	} `json:"taxType" firestore:"taxType"`
}

type TaxType string
type State string

const (
	TaxTypeNormal    TaxType = "normal"    //부가가치세 일반과세자
	TaxTypeFree      TaxType = "free"      //부가가치세 면세과세자
	TaxTypeSimple    TaxType = "simple"    //부가가치세 간이과세자
	TaxTypeNonProfit TaxType = "nonProfit" //비영리법인 또는 국가기관, 고유번호가 부여된 단체

	StateUnregistered State = "unregistered" //0 : 미등록 (등록되지 않은 사업자번호)
	StateNormal       State = "normal"       //1 : 사업중
	StateClose        State = "close"        //2 : 폐업
	StateDown         State = "down"         //3 : 휴업
)

type CloseDown struct {
	BizNum            string  `json:"bizNum" firestore:"bizNum"`
	StateChangeDate   string  `json:"stateChangeDate" firestore:"stateChangeDate"`
	TaxTypeChangeDate string  `json:"taxTypeChangeDate" firestore:"taxTypeChangeDate"`
	TaxType           TaxType `json:"taxType" firestore:"taxType"`
	State             State   `json:"state" firestore:"state"`
}

func ParseHomeTaxCloseDown(resBody string, conf *HomeTaxConfig) CloseDown {
	cd := CloseDown{}

	regex := regexp.MustCompile(conf.State.Date)
	cd.StateChangeDate = strings.Replace(regex.FindString(resBody), "-", "", -1)

	if strings.Contains(resBody, conf.State.Close) {
		cd.State = StateClose
	} else if strings.Contains(resBody, conf.State.Down) {
		cd.State = StateDown
	} else if strings.Contains(resBody, conf.State.Unregistered) {
		cd.State = StateUnregistered
	} else {
		cd.State = StateNormal
	}

	regex = regexp.MustCompile(conf.TaxType.Date)
	m := regex.FindStringSubmatch(resBody)
	if len(m) > 1 {
		cd.TaxTypeChangeDate = m[1]
		cd.TaxTypeChangeDate = strings.Replace(cd.TaxTypeChangeDate, "년", "", -1)
		cd.TaxTypeChangeDate = strings.Replace(cd.TaxTypeChangeDate, "월", "", -1)
		cd.TaxTypeChangeDate = strings.Replace(cd.TaxTypeChangeDate, "일", "", -1)
		cd.TaxTypeChangeDate = strings.Replace(cd.TaxTypeChangeDate, " ", "", -1)
	}

	if strings.Contains(resBody, conf.TaxType.Simple) {
		cd.TaxType = TaxTypeSimple
	} else if strings.Contains(resBody, conf.TaxType.NonProfit) {
		cd.TaxType = TaxTypeNonProfit
	} else if strings.Contains(resBody, conf.TaxType.Free) {
		cd.TaxType = TaxTypeFree
	} else if strings.Contains(resBody, conf.TaxType.Normal) {
		cd.TaxType = TaxTypeNormal
	}

	return cd
}
