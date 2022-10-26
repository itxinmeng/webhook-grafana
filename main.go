package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/wonderivan/logger"
)

type GrafanaAlert struct {
	Title       string `json:"title"`
	RuleID      int    `json:"ruleId"`
	RuleName    string `json:"ruleName"`
	RuleURL     string `json:"ruleUrl"`
	State       string `json:"state"`
	ImageURL    string `json:"imageUrl"`
	Message     string `json:"message"`
	EvalMatches []struct {
		Metric string  `json:"metric"`
		Value  float64 `json:"value"`
	} `json:"evalMatches"`
}

func grafanaServer(writer http.ResponseWriter, request *http.Request) {

	var alert GrafanaAlert

	if request.Method == http.MethodPut || request.Method == http.MethodPost {

		//读取告警内容
		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			return
		} else {
			//json字符串转换成结构体
			json.Unmarshal(body, &alert)
		}

		//获取告警项目id
		id := request.URL.Query().Get("id")

		state := alert.State
		ruleName := alert.RuleName

		var message string   //告警内容
		var title string     //告警标题
		var hostValue string //拼接结果字符串   host:value   主机: 告警值

		//根据状态生成告警标题
		if state == "ok" {
			title = "恢复告警"
			hostValue = "告警已恢复"
			message = alert.Message

		} else if strings.Contains(ruleName, "notification") {
			title = "测试告警"
			hostValue = "Test"

		} else {
			title = "故障告警"
			message = alert.Message

			//遍历获取当前告警值 获取告警主机
			var values []float64
			var hostList []string

			for _, v := range alert.EvalMatches {
				values = append(values, v.Value)
				hostList = append(hostList, v.Metric)

			}

			//将告警主机和当前值进行拼接     host: value
			for i := 0; i < len(hostList); i++ {
				hostValue += hostList[i] + ": " + strconv.FormatFloat(values[i], 'f', 2, 64) + "  "
			}

		}

		logger.SetLogger(`{"File": {"filename": "warning.log","level": "TRAC","append": true,"permit": "0660"}}`)
		//告警字段
		//1.title  2.ruleName  3.message  4.state  5.hostValue
		logger.Info(title, ruleName, message, state, hostValue, id)

		//根据传去不同的id 来做区分 发送短信 ....

	} else {
		fmt.Fprintf(writer, "ok")

	}

}

func main() {
	http.HandleFunc("/webhook", grafanaServer)
	//logger.SetLogger(`{"Console": {"level": "INFO"}}`)

	logger.SetLogger(`{"File": {"filename": "warning.log","level": "TRAC","append": true,"permit": "0660"}}`)
	logger.Info("Running at port 9000 ...")

	err := http.ListenAndServe(":9000", nil)
	if err != nil {
		logger.Error("ListenAndServe: ", err.Error())
	}
}
