package main

import (
	"encoding/xml"
	"fmt"
	"jzlservice/smssender"
	"strconv"
	"strings"
)

func MakeSMSStatus(task_id int64, spnumber string, body []byte) (r *smssender.SMSStatus, err error) {
	r = &smssender.SMSStatus{
		TaskId: task_id,
	}

	//返回值说明：
	//“0”代表提交成功，否则返回失败原因
	//注：返回值以ERROR开头表示有错误，如：有用户名密码错误、屏蔽字、余额不足等。
	result := string(body)
	if result == "0" {
		r.Status = "0"
		r.Message = spnumber
	} else if strings.HasPrefix(result, "ERROR") {
		r.Status = "-1"
		r.Message = strings.TrimPrefix(result, "ERROR")
	} else {
		r.Status = "-1"
		r.Message = result
	}

	return
}

func MakeSMSBalance(body []byte) (r *smssender.SMSBalance, err error) {
	r = &smssender.SMSBalance{}

	//返回值说明：
	//“0”或正数代表提交成功，否则返回失败错误码
	strBody := strings.TrimSpace(string(body))
	result, err := strconv.ParseInt(strBody, 0, 0)
	if result >= 0 {
		r.Status = "0"
	} else {
		r.Status = "-1"
	}
	r.Message = strBody

	return
}

/**
 * xmli结构数据，如下：
 * <?xml version="1.0" encoding="UTF-8"?>
 * <smsResult>
 *   <result>
 *        <spnumber>短信流水号</spnumber>
 *        <phone>手机号码</phone>
 *        <status>DELIVRD</status>
 *        <sendtime>2008-12-24 23:58:24</sendtime>
 *   </result>
 *   <result>
 *        <spnumber>0</spnumber>
 *        <phone>15313235171</phone>
 *        <status>UNDELIV</status>
 *        <sendtime>20121106112751</sendtime>
 *   </result>
 * </smsResult>
 * 说明: status 为 DELIVRD 成功，其他失败。
 * 注：返回值以ERROR开头表示有错误，如：有用户名密码错误等。如果状态还未返回则返回值为“0”
 * 注意大小写
 */
type SMSResult struct {
	SPNumber string `xml:"spnumber"`
	Phone    string `xml:"phone"`
	Status   string `xml:"status"`
	SendTime string `xml:"sendtime"`
}
type SMSResultData struct {
	XMLName xml.Name    `xml:"smsResult"`
	Results []SMSResult `xml:"result"`
}

func FormatDataTime(inStr string) string {
	if len(inStr) == 14 {
		return fmt.Sprintf("%v-%v-%v %v:%v:%v", inStr[0:4], inStr[4:6], inStr[6:8], inStr[8:10], inStr[10:12], inStr[12:14])
	}
	return inStr
}

func MakeSMSReport(body []byte) (r *smssender.SMSReport, err error) {
	r = &smssender.SMSReport{}

	result := string(body)
	if result == "0" {
		r.Status = "0"
		r.Message = "0"
	} else if strings.HasPrefix(result, "ERROR") {
		r.Status = "-1"
		r.Message = strings.TrimPrefix(result, "ERROR")
	} else {
		r.Status = "0"

		v := SMSResultData{}
		err = xml.Unmarshal(body, &v)
		if err != nil {
			LOG_ERROR("解析XML数据失败。失败原因：%v", err)
			return nil, err
		}

		for i, item := range v.Results {
			v.Results[i].SendTime = FormatDataTime(item.SendTime)
		}

		count := len(v.Results)
		r.Message = strconv.FormatInt(int64(count), 10)

		for _, item := range v.Results {
			p := &smssender.SMSReportItem{
				Spnumber: item.SPNumber,
				Mobile:   item.Phone,
				Status:   item.Status,
				Sendtime: item.SendTime,
			}
			r.Data = append(r.Data, p)
		}
	}

	return
}

/*
 * <?xml version="1.0" encoding="UTF-8" ?>
 * <moResult>
 *     <result>
 *         <phone>⼿手机号码</phone>
 *         <content>回复内容</content>
 *         <datetime>时间</datetime>
 *         <dest>⽬目的号码</dest>
 *     </result>
 *     <result>
 *         <phone>15811284187</phone>
 *         <content>您好，因我现在不在北京，请与王红娟联系收货，电话，13522738590</content>
 *         <datetime>20150130084041</datetime>
 *         <dest>10690266002</dest>
 *     </result>
 * </moResult>
 */
type MOResult struct {
	Phone    string `xml:"phone"`
	Content  string `xml:"content"`
	DateTime string `xml:"datetime"`
	Dest     string `xml:"dest"`
}
type MOResultData struct {
	XMLName xml.Name   `xml:"moResult"`
	Results []MOResult `xml:"result"`
}

func MakeSMSMOMessage(body []byte) (r *smssender.SMSMOMessage, err error) {
	r = &smssender.SMSMOMessage{}

	result := string(body)
	if result == "0" {
		r.Status = "0"
		r.Message = "0"
	} else if strings.HasPrefix(result, "ERROR") {
		r.Status = "-1"
		r.Message = strings.TrimPrefix(result, "ERROR")
	} else {
		r.Status = "0"

		v := MOResultData{}
		err = xml.Unmarshal(body, &v)
		if err != nil {
			LOG_ERROR("解析XML数据失败。失败原因：%v", err)
			return nil, err
		}

		count := len(v.Results)
		r.Message = strconv.FormatInt(int64(count), 10)

		for _, item := range v.Results {
			p := &smssender.SMSMOMessageItem{
				Mobile:      item.Phone,
				Content:     item.Content,
				Receivetime: item.DateTime,
				Serviceno:   item.Dest,
			}
			r.Data = append(r.Data, p)
		}
	}

	return
}
