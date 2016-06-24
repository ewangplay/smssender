/* 
 * thrift interface for smssender
 */

namespace cpp jzlservice.smssender
namespace go jzlservice.smssender
namespace py jzlservice.smssender
namespace php jzlservice.smssender
namespace perl jzlservice.smssender
namespace java jzlservice.smssender

/**
* struct SMSEntry
* sms entry structure description.
*/
struct SMSEntry {
    1: i64 task_id,                         //任务ID

    2: string serial_number = "",           //发送批次

    3: string content = "",                 //短信的内容

    4: string receiver = "",                //短信的接收者；如果时多个，之间用“，”间隔

    5: string signature = "",               //消息内容附加的签名信息，如：【枝兰】。注意这里提供的签名信息不包含中文的方括号边界，只是签名信息本身即可，比如：枝兰

    6: string service_minor_number = "",    //分配给服务调用者的服务小号，用来标示不同的调用者身份

    7: byte category = 0,                   //短信消息的种类，1：普通类短信（比如验证短信、通知短信等）；2：营销类短信
}

struct SMSStatus {
    1: i64 task_id,                         //任务ID

    2: string status = "",                  //发送状态。为0表示发送成功；非0表示发送失败

    3: string message = "",                 //详细的状态描述。当status为0，存放发送的批次号；当status为非0时，存放具体的错误描述
}

struct SMSBalance {
    1: string status = "",                  //查询状态。为0表示查询成功；非0表示查询失败

    2: string message = "",                 //详细的状态描述。当status为0，存放余额大小；当status为非0时，存放具体的错误描述
}

struct SMSReportItem {
    1: string spnumber = "",                //对应发送的批次号

    2: string mobile = "",                  //手机号码

    3: string status = "",                  //短信的送达状态。DELIVRD表示送达成功，其他表示送达失败同时表示具体的错误原因

    4: string sendtime = "",                //发送时间
}

struct SMSReport {
    1: string status = "",                  //查询状态。为0表示查询成功；非0表示查询失败

    2: string message = "",                 //详细的状态描述。当status为0，存放报告的条目数；当status为非0时，存放具体的错误描述

    3: list<SMSReportItem> data,            //当status为0时，存放具体的报告数据；当status为非0时，为空
}

struct SMSMOMessageItem {
    1: string mobile = "",                  //回复手机号码

    2: string content = "",                 //回复内容

    3: string serviceno = "",               //回复的服务号

    4: string receivetime = "",             //上行短信接收时间
}

struct SMSMOMessage {
    1: string status = "",                  //查询状态。为0表示查询成功；非0表示查询失败

    2: string message = "",                 //详细的状态描述。当status为0，存放上行短信的条目数；当status为非0时，存放具体的错误描述

    3: list<SMSMOMessageItem> data,         //当status为0时，存放具体的上行短信数据；当status为非0时，为空
}

/**
* smssender service
*/
service SMSSender {
	string ping(),		                    //服务的连通性测试接口
                                            //返回: pong

    list<SMSStatus> sendSMS(1: list<SMSEntry> sms_entries),        //群发短信

    SMSStatus sendMessage(1: SMSEntry sms_entry),                  //单一发送

    SMSBalance getBalance(1: i16 category),                //查询余额

    SMSReport getReport(1: i16 category),                  //查询发送报告

    SMSMOMessage getMOMessage(1: i16 category),            //获取上行短信
    
    list<i16> getCategory(),                                   //获取短信通道的类别列表
}

