package main

import (
	"fmt"
	iconv "github.com/djimenez/iconv-go"
	utils "github.com/ewangplay/go-utils"
	"io/ioutil"
	"jzlservice/smssender"
	"net/http"
)

// SMSSenderImpl implementaion
type SMSSenderImpl struct {
}

func (this *SMSSenderImpl) Ping() (r string, err error) {
	LOG_INFO("请求ping方法")
	return "pong", nil
}

func (this *SMSSenderImpl) SendSMS(sms_entries []*smssender.SMSEntry) (r []*smssender.SMSStatus, err error) {
	LOG_INFO("请求SendSMS方法")

	for _, entry := range sms_entries {
		LOG_DEBUG("短信[%v => %v]准备发送", entry.Content, entry.Receiver)

		result, err := this.SendMessage(entry)
		if err != nil {
			LOG_ERROR("短信[%v => %v]发送失败", entry.Content, entry.Receiver)

			continue
		}

		r = append(r, result)
	}

	return r, err
}

func (this *SMSSenderImpl) SendMessage(entry *smssender.SMSEntry) (r *smssender.SMSStatus, err error) {

	var outputStr string
	var sendSMSUrl string
	var sms_service_provider_addr string
	var sms_service_provider_port string
	var sms_service_provider_user string
	var sms_service_provider_password string
	var sms_service_provider_user_market string
	var sms_service_provider_password_market string
	var sms_service_provider_signature string
	var ok bool

	LOG_INFO("短信[%v => %v]开始发送", entry.Content, entry.Receiver)

	//retrive and check the url elems
	sms_service_provider_addr, ok = g_config.Get("sms_service_provider.addr")
	if !ok {
		outputStr = "短信服务的网络地址没有配置"
		LOG_ERROR(outputStr)
		return nil, fmt.Errorf(outputStr)
	}
	sms_service_provider_port, ok = g_config.Get("sms_service_provider.port")
	if !ok {
		outputStr = "短信服务的网络端口没有设置"
		LOG_ERROR(outputStr)
		return nil, fmt.Errorf(outputStr)
	}
	sms_service_provider_user, ok = g_config.Get("sms_service_provider.user")
	if !ok {
		outputStr = "短信服务的账号名称没有设置"
		LOG_ERROR(outputStr)
		return nil, fmt.Errorf(outputStr)
	}
	sms_service_provider_password, ok = g_config.Get("sms_service_provider.password")
	if !ok {
		outputStr = "短信服务的账号密码没有设置"
		LOG_ERROR(outputStr)
		return nil, fmt.Errorf(outputStr)
	}
	sms_service_provider_user_market, ok = g_config.Get("sms_service_provider.user.market")
	if !ok {
		outputStr = "短信服务的营销账号名称没有设置"
		LOG_ERROR(outputStr)
		return nil, fmt.Errorf(outputStr)
	}
	sms_service_provider_password_market, ok = g_config.Get("sms_service_provider.password.market")
	if !ok {
		outputStr = "短信服务的营销账号密码没有设置"
		LOG_ERROR(outputStr)
		return nil, fmt.Errorf(outputStr)
	}

	content := entry.Content

	//append enterprise signature to the content
	//优先使用通过参数传过来的签名，如果参数中的签名信息为空，那么再
	//使用配置文件中设置的签名信息，如果配置文件中没有配置签名信息，那么
	//消息不追加任何签名信息
	if entry.Signature != "" {
		content = content + "【" + entry.Signature + "】"
	} else {
		sms_service_provider_signature, ok = g_config.Get("sms_service_provider.signature")
		if ok && sms_service_provider_signature != "" {
			content = content + "【" + sms_service_provider_signature + "】"
		} else {
			content = content + "【】"
		}
	}

	//convert the content encode from utf-8 to gbk
	contentGBK, err := iconv.ConvertString(content, "utf-8", "gbk")
	if err != nil {
		LOG_ERROR("转换短信内容[%v]从utf-8编码到gbk失败. 失败原因：%v", content, err)
		return nil, err
	}

	//urlencode the content
	contentUrlEncode := utils.UrlEncode(contentGBK)

	//format the request url
	switch entry.Category {
	case 1: //normal
		sendSMSUrl = fmt.Sprintf("http://%v:%v/sms/push_mt.jsp?cpid=%v&cppwd=%v&phone=%v&spnumber=%v&msgcont=%v&extend=%v", sms_service_provider_addr, sms_service_provider_port, sms_service_provider_user, sms_service_provider_password, entry.Receiver, entry.SerialNumber, contentUrlEncode, entry.ServiceMinorNumber)
	case 2: //market
		sendSMSUrl = fmt.Sprintf("http://%v:%v/sms/push_mt.jsp?cpid=%v&cppwd=%v&phone=%v&spnumber=%v&msgcont=%v&extend=%v", sms_service_provider_addr, sms_service_provider_port, sms_service_provider_user_market, sms_service_provider_password_market, entry.Receiver, entry.SerialNumber, contentUrlEncode, entry.ServiceMinorNumber)
	default:
		sendSMSUrl = fmt.Sprintf("http://%v:%v/sms/push_mt.jsp?cpid=%v&cppwd=%v&phone=%v&spnumber=%v&msgcont=%v&extend=%v", sms_service_provider_addr, sms_service_provider_port, sms_service_provider_user, sms_service_provider_password, entry.Receiver, entry.SerialNumber, contentUrlEncode, entry.ServiceMinorNumber)
	}

	LOG_DEBUG("UrlEncode编码后的URL地址：%v", sendSMSUrl)

	resp, err := http.Get(sendSMSUrl)
	if err != nil {
		LOG_ERROR("请求短信发送URL[%v]失败，失败原因：%v", sendSMSUrl, err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		LOG_ERROR("读取HTTP响应信息失败，失败原因：%v", err)
		return nil, err
	}

	if resp.StatusCode != 200 {
		LOG_ERROR("短信[%v => %v]发送失败。HTTP状态码: %v", entry.Content, entry.Receiver, resp.StatusCode)
		return nil, fmt.Errorf("%v", resp.StatusCode)
	}

	r, err = MakeSMSStatus(entry.TaskId, entry.SerialNumber, body)
	if err != nil {
		LOG_ERROR("生成短信发送状态失败，失败原因：%v", err)
		return nil, err
	}

	LOG_INFO("短信[%v => %v]发送成功: %v", entry.Content, entry.Receiver, r)

	return r, nil
}

func (this *SMSSenderImpl) GetBalance(category int16) (r *smssender.SMSBalance, err error) {
	LOG_INFO("请求GetBalance方法")

	var outputStr string
	var getBalanceUrl string
	var sms_service_provider_addr string
	var sms_service_provider_port string
	var sms_service_provider_user string
	var sms_service_provider_password string
	var sms_service_provider_user_market string
	var sms_service_provider_password_market string
	var ok bool

	sms_service_provider_addr, ok = g_config.Get("sms_service_provider.addr")
	if !ok {
		outputStr = "短信服务的网络地址没有配置"
		LOG_ERROR(outputStr)
		return nil, fmt.Errorf(outputStr)
	}
	sms_service_provider_port, ok = g_config.Get("sms_service_provider.port")
	if !ok {
		outputStr = "短信服务的网络端口没有设置"
		LOG_ERROR(outputStr)
		return nil, fmt.Errorf(outputStr)
	}
	sms_service_provider_user, ok = g_config.Get("sms_service_provider.user")
	if !ok {
		outputStr = "短信服务的账号名称没有设置"
		LOG_ERROR(outputStr)
		return nil, fmt.Errorf(outputStr)
	}
	sms_service_provider_password, ok = g_config.Get("sms_service_provider.password")
	if !ok {
		outputStr = "短信服务的账号密码没有设置"
		LOG_ERROR(outputStr)
		return nil, fmt.Errorf(outputStr)
	}
	sms_service_provider_user_market, ok = g_config.Get("sms_service_provider.user.market")
	if !ok {
		outputStr = "短信服务的营销账号名称没有设置"
		LOG_ERROR(outputStr)
		return nil, fmt.Errorf(outputStr)
	}
	sms_service_provider_password_market, ok = g_config.Get("sms_service_provider.password.market")
	if !ok {
		outputStr = "短信服务的营销账号密码没有设置"
		LOG_ERROR(outputStr)
		return nil, fmt.Errorf(outputStr)
	}

	switch category {
	case 1: //normal
		getBalanceUrl = fmt.Sprintf("http://%v:%v/user/qamount.jsp?cpid=%v&pwd=%v", sms_service_provider_addr, sms_service_provider_port, sms_service_provider_user, sms_service_provider_password)
	case 2: //market
		getBalanceUrl = fmt.Sprintf("http://%v:%v/user/qamount.jsp?cpid=%v&pwd=%v", sms_service_provider_addr, sms_service_provider_port, sms_service_provider_user_market, sms_service_provider_password_market)
	default:
		getBalanceUrl = fmt.Sprintf("http://%v:%v/user/qamount.jsp?cpid=%v&pwd=%v", sms_service_provider_addr, sms_service_provider_port, sms_service_provider_user, sms_service_provider_password)
	}

	resp, err := http.Get(getBalanceUrl)
	if err != nil {
		LOG_ERROR("请求获取余额URL[%v]失败. 失败原因：%v", getBalanceUrl, err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		LOG_ERROR("读取HTTP响应信息失败，失败原因：%v", err)
		return nil, err
	}

	if resp.StatusCode != 200 {
		LOG_ERROR("获取短信余额信息失败。HTTP状态码: %v", resp.StatusCode)
		return nil, fmt.Errorf("%v", resp.StatusCode)
	}

	r, err = MakeSMSBalance(body)
	if err != nil {
		LOG_ERROR("生成短信余额数据失败，失败原因：%v", err)
		return nil, err
	}

	LOG_INFO("获取短信余额信息成功: %v", r)

	return r, nil
}

func (this *SMSSenderImpl) GetReport(category int16) (r *smssender.SMSReport, err error) {
	LOG_INFO("请求GetReport方法")

	var outputStr string
	var getReportUrl string
	var sms_service_provider_addr string
	var sms_service_provider_port string
	var sms_service_provider_user string
	var sms_service_provider_password string
	var sms_service_provider_user_market string
	var sms_service_provider_password_market string
	var ok bool

	sms_service_provider_addr, ok = g_config.Get("sms_service_provider.addr")
	if !ok {
		outputStr = "短信服务的网络地址没有配置"
		LOG_ERROR(outputStr)
		return nil, fmt.Errorf(outputStr)
	}
	sms_service_provider_port, ok = g_config.Get("sms_service_provider.port")
	if !ok {
		outputStr = "短信服务的网络端口没有设置"
		LOG_ERROR(outputStr)
		return nil, fmt.Errorf(outputStr)
	}
	sms_service_provider_user, ok = g_config.Get("sms_service_provider.user")
	if !ok {
		outputStr = "短信服务的账号名称没有设置"
		LOG_ERROR(outputStr)
		return nil, fmt.Errorf(outputStr)
	}
	sms_service_provider_password, ok = g_config.Get("sms_service_provider.password")
	if !ok {
		outputStr = "短信服务的账号密码没有设置"
		LOG_ERROR(outputStr)
		return nil, fmt.Errorf(outputStr)
	}
	sms_service_provider_user_market, ok = g_config.Get("sms_service_provider.user.market")
	if !ok {
		outputStr = "短信服务的营销账号名称没有设置"
		LOG_ERROR(outputStr)
		return nil, fmt.Errorf(outputStr)
	}
	sms_service_provider_password_market, ok = g_config.Get("sms_service_provider.password.market")
	if !ok {
		outputStr = "短信服务的营销账号密码没有设置"
		LOG_ERROR(outputStr)
		return nil, fmt.Errorf(outputStr)
	}

	switch category {
	case 1: //normal
		getReportUrl = fmt.Sprintf("http://%v:%v/sms/getreport.jsp?cpid=%v&cppwd=%v", sms_service_provider_addr, sms_service_provider_port, sms_service_provider_user, sms_service_provider_password)
	case 2: //market
		getReportUrl = fmt.Sprintf("http://%v:%v/sms/getreport.jsp?cpid=%v&cppwd=%v", sms_service_provider_addr, sms_service_provider_port, sms_service_provider_user_market, sms_service_provider_password_market)
	default:
		getReportUrl = fmt.Sprintf("http://%v:%v/sms/getreport.jsp?cpid=%v&cppwd=%v", sms_service_provider_addr, sms_service_provider_port, sms_service_provider_user, sms_service_provider_password)
	}

	resp, err := http.Get(getReportUrl)
	if err != nil {
		LOG_ERROR("请求获取发送短信状态的URL[%v]失败. 失败原因：%v", getReportUrl, err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		LOG_ERROR("读取HTTP响应信息失败，失败原因：%v", err)
		return nil, err
	}

	if resp.StatusCode != 200 {
		LOG_ERROR("获取短信发送状态失败。HTTP状态码: %v", resp.StatusCode)
		return nil, fmt.Errorf("%v", resp.StatusCode)
	}

	r, err = MakeSMSReport(body)
	if err != nil {
		LOG_ERROR("生成短信报告数据失败，失败原因：%v", err)
		return nil, err
	}

	LOG_INFO("获取短信发送报告成功: %v", r)

	return r, nil
}

func (this *SMSSenderImpl) GetMOMessage(category int16) (r *smssender.SMSMOMessage, err error) {
	LOG_INFO("请求GetMOMessage方法")

	var outputStr string
	var recvMessageUrl string
	var sms_service_provider_addr string
	var sms_service_provider_port string
	var sms_service_provider_user string
	var sms_service_provider_password string
	var sms_service_provider_user_market string
	var sms_service_provider_password_market string
	var ok bool

	sms_service_provider_addr, ok = g_config.Get("sms_service_provider.addr")
	if !ok {
		outputStr = "短信服务的网络地址没有配置"
		LOG_ERROR(outputStr)
		return nil, fmt.Errorf(outputStr)
	}
	sms_service_provider_port, ok = g_config.Get("sms_service_provider.port")
	if !ok {
		outputStr = "短信服务的网络端口没有设置"
		LOG_ERROR(outputStr)
		return nil, fmt.Errorf(outputStr)
	}
	sms_service_provider_user, ok = g_config.Get("sms_service_provider.user")
	if !ok {
		outputStr = "短信服务的账号名称没有设置"
		LOG_ERROR(outputStr)
		return nil, fmt.Errorf(outputStr)
	}
	sms_service_provider_password, ok = g_config.Get("sms_service_provider.password")
	if !ok {
		outputStr = "短信服务的账号密码没有设置"
		LOG_ERROR(outputStr)
		return nil, fmt.Errorf(outputStr)
	}
	sms_service_provider_user_market, ok = g_config.Get("sms_service_provider.user.market")
	if !ok {
		outputStr = "短信服务的营销账号名称没有设置"
		LOG_ERROR(outputStr)
		return nil, fmt.Errorf(outputStr)
	}
	sms_service_provider_password_market, ok = g_config.Get("sms_service_provider.password.market")
	if !ok {
		outputStr = "短信服务的营销账号密码没有设置"
		LOG_ERROR(outputStr)
		return nil, fmt.Errorf(outputStr)
	}

	switch category {
	case 1: //normal
		recvMessageUrl = fmt.Sprintf("http://%v:%v/sms/getmo.jsp?cpid=%v&cppwd=%v", sms_service_provider_addr, sms_service_provider_port, sms_service_provider_user, sms_service_provider_password)
	case 2: //market
		recvMessageUrl = fmt.Sprintf("http://%v:%v/sms/getmo.jsp?cpid=%v&cppwd=%v", sms_service_provider_addr, sms_service_provider_port, sms_service_provider_user_market, sms_service_provider_password_market)
	default:
		recvMessageUrl = fmt.Sprintf("http://%v:%v/sms/getmo.jsp?cpid=%v&cppwd=%v", sms_service_provider_addr, sms_service_provider_port, sms_service_provider_user, sms_service_provider_password)
	}

	resp, err := http.Get(recvMessageUrl)
	if err != nil {
		LOG_ERROR("请求获取上行短信的URL[%v]失败. 失败原因：%v", recvMessageUrl, err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		LOG_ERROR("读取HTTP响应信息失败，失败原因：%v", err)
		return nil, err
	}

	if resp.StatusCode != 200 {
		LOG_ERROR("获取上行短信失败。HTTP状态码: %v", resp.StatusCode)
		return nil, fmt.Errorf("%v", resp.StatusCode)
	}

	r, err = MakeSMSMOMessage(body)
	if err != nil {
		LOG_ERROR("生成短信上行数据失败，失败原因：%v", err)
		return nil, err
	}

	LOG_INFO("获取上行短信成功: %v", r)

	return r, nil
}

func (this *SMSSenderImpl) GetCategory() (r []int16, err error) {
	return []int16{1, 2}, nil
}
