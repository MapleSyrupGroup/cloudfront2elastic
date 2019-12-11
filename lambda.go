package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// Struct for storing the log data in an object that can be serialized into JSON.
type LogLine struct {
	Date             string `json:"date"`
	ClientIP         string `json:"client_ip"`
	ClientMethod     string `json:"client_method"`
	RequestURIStem   string `json:"request_uri_stem"`
	HTTPResponseCode string `json:"http_response_code"`
	RefererPage      string `json:"referer_page"`
	ClientUserAgent  string `json:"client_user_agent"`
	BotDetected      string `json:"bot_detected"`
	QueryString      string `json:"query_string"`
	CFEdgeResult     string `json:"cf_edge_result"`
	HostHeader       string `json:"host_header"`
	HTTPProtocol     string `json:"http_protocol"`
	SSLProtocol      string `json:"ssl_protocol"`
	CFClientResult   string `json:"cf_client_result"`
	IndexName        string `json:"-"`
}

// Type and constant array for making the extraction more human readable.
type LogPart int

const (
	CFDate            LogPart = 0
	CFTime            LogPart = 1
	CIp               LogPart = 4
	CSMethod          LogPart = 5
	CSUriStem         LogPart = 7
	SCStatus          LogPart = 8
	CSReferer         LogPart = 9
	CSUserAgent       LogPart = 10
	CSQueryString     LogPart = 11
	XEdgeResultType   LogPart = 13
	XHostHeader       LogPart = 15
	CSProtocol        LogPart = 16
	SSLProtocol       LogPart = 20
	XEdgeResponseType LogPart = 22
)

func HandleRequest(ctx context.Context, event events.S3Event) {
	log.Print("New log file detected: " + event.Records[0].S3.Object.Key)
	bulkJson := ""
	var jsonBuffer bytes.Buffer
	zippedObject, err := getS3Object(event.Records[0].S3.Bucket.Name, event.Records[0].S3.Object.Key)
	if err != nil {
		log.Fatal("Unable to GET new file from S3.")
	}
	buf := bytes.NewBuffer(zippedObject)
	unzippedObject, err := gzip.NewReader(buf)
	if err != nil {
		log.Fatal("Unable to unzip new file.")
	}
	scanner := bufio.NewScanner(unzippedObject)
	for scanner.Scan() {
		lineText := scanner.Text()
		lineParts := strings.Split(lineText, "\t")
		if strings.HasPrefix(lineParts[0], "#") {
			continue
		} else {
			parsedLine, err := parseLogLine(lineParts)
			if err != nil {
				log.Fatal(err)
			} else {
				lineJsonBytes, err := json.Marshal(parsedLine)
				if err != nil {
					log.Fatal("Object could not be converted to JSON.")
				}
				lineJsonString := string(lineJsonBytes)
				jsonBuffer.WriteString("{ \"index\" : { \"_index\" : \"" + parsedLine.IndexName + "\"} }\n" + lineJsonString + "\n")
			}
		}

	}
	bulkJson = jsonBuffer.String()
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	elasticURI := os.Getenv("elastic_url")
	var payloadBytes = []byte(bulkJson)
	req, err := http.NewRequest("POST", elasticURI, bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(os.Getenv("elastic_user"), os.Getenv("elastic_password"))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Post failed")
	}
	defer resp.Body.Close()
	log.Printf("Elastic response code: %s", resp.Status)
}

func parseLogLine(logLineParts []string) (*LogLine, error) {
	if len(logLineParts) != 33 {
		return nil, fmt.Errorf("log line import failed: row length incorrect expected 26 got :%s", strconv.FormatInt(int64(len(logLineParts)), 10))
	}
	var parsedLine LogLine
	parsedLine.IndexName = "weblogs-" + logLineParts[CFDate]
	parsedLine.Date = logLineParts[CFDate] + " " + logLineParts[CFTime]
	parsedLine.ClientIP = logLineParts[CIp]
	parsedLine.ClientMethod = logLineParts[CSMethod]
	parsedLine.RequestURIStem = logLineParts[CSUriStem]
	parsedLine.HTTPResponseCode = logLineParts[SCStatus]
	parsedLine.RefererPage = logLineParts[CSReferer]
	parsedLine.ClientUserAgent = logLineParts[CSUserAgent]
	if strings.Contains(strings.ToLower(logLineParts[CSUserAgent]), "googlebot") {
		parsedLine.BotDetected = "GoogleBot"
	} else if strings.Contains(strings.ToLower(logLineParts[CSUserAgent]), "googleimageproxy") {
		parsedLine.BotDetected = "GoogleImageBot"
	} else if strings.Contains(strings.ToLower(logLineParts[CSUserAgent]), "bingbot") {
		parsedLine.BotDetected = "BingBot"
	} else if strings.Contains(strings.ToLower(logLineParts[CSUserAgent]), "baiduspider") {
		parsedLine.BotDetected = "BaiduSpider"
	} else if strings.Contains(strings.ToLower(logLineParts[CSUserAgent]), "yahoo!") {
		parsedLine.BotDetected = "Yahoo! Slurp"
	} else if strings.Contains(strings.ToLower(logLineParts[CSUserAgent]), "yandex") {
		parsedLine.BotDetected = "Yandex"
	} else {
		parsedLine.BotDetected = "Human"
	}
	parsedLine.QueryString = logLineParts[CSQueryString]
	parsedLine.CFEdgeResult = logLineParts[XEdgeResultType]
	parsedLine.HostHeader = logLineParts[XHostHeader]
	parsedLine.HTTPProtocol = logLineParts[CSProtocol]
	parsedLine.SSLProtocol = logLineParts[SSLProtocol]
	parsedLine.CFClientResult = logLineParts[XEdgeResponseType]

	return &parsedLine, nil
}

// Returns contents of S3 item as byte array
func getS3Object(bucket string, key string) ([]byte, error) {
	awsSess := session.New()
	s3req := s3.New(awsSess, aws.NewConfig().WithRegion(os.Getenv("aws_s3_region")))
	res, err := s3req.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, res.Body); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func main() {
	lambda.Start(HandleRequest)
}
