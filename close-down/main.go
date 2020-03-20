package main

import (
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/imroc/req"
	"github.com/theorders/aefire"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var (
	ErrBizNumNotProvided       = errors.New("사업자등록번호가 지정되지 않았습니다")
	ErrBizNum                  = errors.New("사업자등록번호 형식이 맞지 않습니다")
	ErrHometaxConnectionFailed = errors.New("홈택스 연결에 실패했습니다")
)

var homeTaxConfig HomeTaxConfig

func init() {
	homeTaxConfig = HomeTaxConfig{}

	homeTaxConfig.State.Close = "폐업"
	homeTaxConfig.State.Date = "([0-9]{4}-[0-9]{2}-[0-9]{2})"
	homeTaxConfig.State.Down = "휴업"
	homeTaxConfig.State.Unregistered = "사업을 하지 않고 있습니다"

	homeTaxConfig.TaxType.Date = "과세유형 전환된 날짜는 ([0-9]{4}년.[0-9]{2}월.[0-9]{2}일)"
	homeTaxConfig.TaxType.Free = "면세"
	homeTaxConfig.TaxType.NonProfit = "단체"
	homeTaxConfig.TaxType.Normal = "일반"
	homeTaxConfig.TaxType.Simple = "간이"
}

func ValidateBizNum(n string) bool {
	n = strings.Replace(n, "-", "", -1)
	n = strings.Replace(n, " ", "", -1)

	if len(n) != 10 {
		return false
	}

	multipliers := []int{1, 3, 7, 1, 3, 7, 1, 3, 5}
	checksum := 0

	for i, v := range multipliers {
		digit, _ := strconv.Atoi(string(n[i]))
		if i < 8 {
			checksum += v * digit % 10
		} else {
			checksum += (v * digit % 10) + (v * digit / 10)
		}
	}

	lastDigit, _ := strconv.Atoi(string(n[9]))
	checksum = checksum + lastDigit

	return checksum%10 == 0
}

// Handler is your Lambda function handler
// It uses Amazon API Gateway request/responses provided by the aws-lambda-go/events package,
// However you could use other event sources (S3, Kinesis etc), or JSON-decoded primitive types such as 'string'.
func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	// stdout and stderr are sent to AWS CloudWatch Logs
	log.Printf("Processing Lambda request %s\n", request.RequestContext.RequestID)

	bizNum := request.QueryStringParameters["bizNum"]

	// If no name is provided in the HTTP request body, throw an error
	if bizNum == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       aefire.ToJson(aefire.MapOf("message", "사업자등록번호가 지정되지 않았습니다")),
		}, nil
	}

	if !ValidateBizNum(bizNum) {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       aefire.ToJson(aefire.MapOf("message", "사업자등록번호 형식이 맞지 않습니다")),
		}, nil

	}

	q := req.New()

	res, err := q.Post(
		"https://teht.hometax.go.kr/wqAction.do?actionId=ATTABZAA001R08&screenId=UTEABAAA13&popupYn=false&realScreenId=",
		req.BodyXML(fmt.Sprintf(`<map id='ATTABZAA001R08'><pubcUserNo/><mobYn>N</mobYn><inqrTrgtClCd>1</inqrTrgtClCd><txprDscmNo>%s</txprDscmNo><dongCode>__MIDDLE__</dongCode><psbSearch>Y</psbSearch><map id='userReqInfoVO'/></map>`, bizNum)))

	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       aefire.ToJson(aefire.MapOf("message", "홈택스 연결이 실패했습니다")),
		}, nil
	}

	closeDown := ParseHomeTaxCloseDown(res.String(), &homeTaxConfig)
	closeDown.BizNum = bizNum

	return events.APIGatewayProxyResponse{
		Body:       aefire.ToJson(closeDown),
		StatusCode: 200,
	}, nil

}

func main() {
	lambda.Start(Handler)
}
