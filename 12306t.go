package main

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"encoding/json"
	"math/rand"
	"time"
	"strconv"
	"regexp"
	"bufio"
)

//全局变量
var CurCookies []*http.Cookie
var CurCookieJar *cookiejar.Jar //管理cookie

var RailwayDate string
var PersonType string
var seatTypeStr string

//初始化
func init() {
	CurCookies = nil
	//var err error;
	CurCookieJar,_ = cookiejar.New(nil)
}

func printCookies(){
	var cookieNum int = len(CurCookies)
	fmt.Printf("cookieNum=%d\r\n", cookieNum)
	for i:=0;i<cookieNum ;i++  {
		var CurCK = CurCookies[i]
		fmt.Printf("curCK Raw: %s = %s\n", CurCK.Name, CurCK.Value)
	}
}

//

//get url response html
func getUrlRespHtml(strUrl string, postDict map[string]string) string{
	fmt.Printf("in getUrlRespHtml, strUrl=%s\n", strUrl)
	var respHtml string = ""

	httpClient := &http.Client{
		Jar:CurCookieJar,
	}
	var httpReq *http.Request
	if nil == postDict {
		fmt.Printf("is GET\n")
		httpReq, _ = http.NewRequest("GET", strUrl, nil)

	} else {
		fmt.Printf("is POST\n")
		postValues := url.Values{}
		for postKey, PostValue := range postDict{
			postValues.Set(postKey, PostValue)
		}

		postDataStr := ""
		for k, v := range postValues{
			postDataStr += k + "=" + v[0] + "&"
		}
		postDataStr = postDataStr[0:len(postDataStr)-1]
		fmt.Printf("postDataStr=%s\n", postDataStr)
		postDataBytes := []byte(postDataStr)
		postBytesReader := bytes.NewReader(postDataBytes)
		httpReq, _ = http.NewRequest("POST", strUrl, postBytesReader)
		httpReq.Header.Add("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
	}

	httpResp, err := httpClient.Do(httpReq)
	if err != nil {
		fmt.Printf("http get strUrl=%s response error=%s\n", strUrl, err.Error())
	}
	fmt.Printf("httpResp.Header=%s\n", httpResp.Header)
	fmt.Printf("httpResp.Status=%s\n", httpResp.Status)
	defer httpResp.Body.Close()
	body, errReadAll := ioutil.ReadAll(httpResp.Body)
	if errReadAll != nil {
		fmt.Printf("get response for strUrl=%s got error=%s\n", strUrl, errReadAll.Error())
	}
	//CurCookies = CurCookieJar.Cookies(httpReq.URL)
	CurCookieJar.SetCookies(httpReq.URL, CurCookies)
	respHtml = string(body)
	return respHtml
}

func getFile(strUrl string, postDict map[string]string, fileAllName string){

	httpClient := &http.Client{
		Jar:CurCookieJar,
	}

	var httpReq *http.Request
	if nil == postDict {
		httpReq, _ = http.NewRequest("GET", strUrl, nil)

	} else {
		postValues := url.Values{}
		for postKey, PostValue := range postDict{
			postValues.Set(postKey, PostValue)
		}
		postDataStr := postValues.Encode()
		postDataBytes := []byte(postDataStr)
		postBytesReader := bytes.NewReader(postDataBytes)
		httpReq, _ = http.NewRequest("POST", strUrl, postBytesReader)
		httpReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	httpResp, err := httpClient.Do(httpReq)
	if err != nil {
		fmt.Printf("http get strUrl=%s response error=%s\n", strUrl, err.Error())
	}

	defer httpResp.Body.Close()

	body, errReadAll := ioutil.ReadAll(httpResp.Body)
	if errReadAll != nil {
		fmt.Printf("get response for strUrl=%s got error=%s\n", strUrl, errReadAll.Error())
	}
	//CurCookies = CurCookieJar.Cookies(httpReq.URL)
	CurCookieJar.SetCookies(httpReq.URL, CurCookies)
	code, error := os.Create(fileAllName)

	if error != nil {
		fmt.Println(error)
	}
	code.Write([]byte(body))
	code.Close()
}


//解析服务器返回的json字符串
func jsonstr2map(jsonstr string)(m map[string]interface{}, err error){

	m = make(map[string]interface{})
	if err := json.Unmarshal([]byte(jsonstr), &m); err != nil {
		return nil, err
	}
	return m, nil
}


//根据城市名称返回城市代码
func cityCode(cityname string)(citycode string){

	//查询城市代码
	urlAllCity :="https://kyfw.12306.cn/otn/userCommon/allCitys"
	var postDictAllCity = make(map[string]string)
	postDictAllCity["station_name"] = ""
	postDictAllCity["_json_att"] = ""
	bodyAllCity :=getUrlRespHtml(urlAllCity, postDictAllCity)

	m, err := jsonstr2map(bodyAllCity)
	if err!=nil{
		fmt.Println(err)
	}
	mCitys := m["data"]
	fmt.Println(mCitys)
	return citycode
}

//根据回答生成验证码信息
func getanswer()string{
	i:=0
	var useranser string
	userinput := [9]int{0}
	captchas := map[int]string{
		1:"34%2C34",
		2:"109%2C39",
		3:"184%2C32",
		4:"251%2C43",
		5:"49%2C105",
		6:"110%2C109",
		7:"190%2C116",
		8:"251%2C106",
	}

	for{
		fmt.Scanf("%d\r\n", &userinput[i])
		if userinput[i] == -1{
			userinput[i] = 0
			break
		}
		i++
	}
	fmt.Println(userinput)
	for j:=0;j<8 ;j++  {
		if userinput[j] !=0 {
			useranser += captchas[userinput[j]] + "%2C"
		}
	}
	useranser = strings.Trim(useranser,"%2C")
	fmt.Println(useranser)
	return useranser
}


func main(){
	for{
		//获取验证码：
		oneGet := "https://kyfw.12306.cn/passport/captcha/captcha-image?login_site=E&module=login&rand=sjrand&0.07259546803051764"
		getFile(oneGet, nil, `D:\captcha.png`)

		//设置cookie
		//_passport_ct=810e6e97c8bc43b6807f2e1f52891543t3009; Path=/passport

		//通过生成的验证码写入验证码相应序号
		fmt.Println("input the captcha num")
		useranswer := getanswer()
		fmt.Println(useranswer)

		//验证验证码的正确性
		var postDict =make(map[string]string)
		postDict["answer"] = useranswer
		postDict["login_site"] = "E"
		postDict["rand"] = "sjrand"
		captchacheckurl:="https://kyfw.12306.cn/passport/captcha/captcha-check"
		body2 :=getUrlRespHtml(captchacheckurl, postDict)
		fmt.Println(body2)
		printCookies()
		if strings.Contains(body2,"验证码校验成功"){
			break
		}
	}

	//POST账号密码
	logincheck:="https://kyfw.12306.cn/passport/web/login"
	var postDict2 = make(map[string]string)
	accountName := ""
	accountPassword := ""
	fmt.Println("input the 12306 accout")
	fmt.Scanln(&accountName)
	fmt.Println("input the 12306 password")
	fmt.Scanln(&accountPassword)
	postDict2["username"] = accountName
	postDict2["password"] = accountPassword
	postDict2["appid"] ="otn"
	body3 :=getUrlRespHtml(logincheck, postDict2)
	fmt.Println(body3)
	printCookies()

	//通过POST appid获取newapptk
	checkAuth:="https://kyfw.12306.cn/passport/web/auth/uamtk"
	var postDict3 = make(map[string]string)
	postDict3["appid"] ="otn"
	body4 :=getUrlRespHtml(checkAuth, postDict3)
	fmt.Println("body4 is ->>>", body4)

	//POST响应得到的tk信息，判断是否通过
	tkurl:="https://kyfw.12306.cn/otn/uamauthclient"
	var postDict4 = make(map[string]string)
	m,err :=jsonstr2map(body4)
	if err!= nil{
		fmt.Println(err)
	}
	fmt.Println("m is ", m)
	postDict4["tk"] = m["newapptk"].(string)
	body5 :=getUrlRespHtml(tkurl, postDict4)
	fmt.Println(body5)


	//GET登录成功的页面
    mainPage:="https://kyfw.12306.cn/otn/index/initMy12306"
	getUrlRespHtml(mainPage, nil)
	//fmt.Println(body6)

	//订单流程
	orderProcess()

}

//订单流程
func orderProcess(){

	//构造查询车次的URL
	queryURL:=getQueryURL()
	fmt.Println(queryURL)

	//查询车次结果
	railwayResult := railwayInfo(queryURL)

	//显示查询结果
	formatRailwayInfo(railwayResult)

	//选择车次序号
	railwayInfo := chooseRailway(railwayResult)
	fmt.Println(railwayInfo)

	//检测用户是否登录
	isLogin := isLoginUser()
	if !isLogin{
		fmt.Println("no login please login!")
		loginOut()
		os.Exit(-1)
	}

	//提交订单请求
	isSubmit := submitOrderRequest(railwayInfo)
	if !isSubmit{
		for i:=0; i < 5; i++{
			isSubmit := submitOrderRequest(railwayInfo)
			if isSubmit{
				fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
				break
			}
		}
		fmt.Println("submitOrderRequest error")
		loginOut()
		os.Exit(-1)
	}

	//初始化订单返回信息
	initDCInfo :=initDc()
	fmt.Println(initDCInfo)

	//获取乘客信息
	passengersInfo := getPassengerDTOs(initDCInfo["globalRepeatSubmitToken"])

	//选择乘客
	passenger := choosePassenger(passengersInfo)

	seatTypeStr = seatType()

	//检查订单信息
	passengerTicketInfo, checkOrderInfoResult := checkOrderInfo(passenger,initDCInfo["globalRepeatSubmitToken"])
	if !checkOrderInfoResult{
		fmt.Println("订单失败")
	}

	//获取队列数量
	checkGetQueueCount := getQueueCount(railwayInfo, initDCInfo)
	if !checkGetQueueCount{
		fmt.Println("getQueueCount失败")
	}

	//确认排队队列
	checkConfirmSingleForQueue := confirmSingleForQueue(passengerTicketInfo, initDCInfo, railwayInfo)
	if !checkConfirmSingleForQueue{
		fmt.Println("checkConfirmSingleForQueue 失败")
	}

	//等待排队结果
	orderId := ""
	for i:=0; i < 5; i++{
		orderId = queryOrderWaitTime(initDCInfo["globalRepeatSubmitToken"])
		if orderId != ""{
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if orderId == ""{
		loginOut()
		fmt.Println(">>>>>>>>>>>>>>>>>>>")
		os.Exit(-1)
	}
	fmt.Println( "订单号码：", orderId)


	for i:=0; i < 5; i++{
		r := resultOrderForDcQueue(orderId, initDCInfo["globalRepeatSubmitToken"])
		if !r{
			continue
		}else {
			break
		}
	}

	loginOut()
}

//确认订单
func resultOrderForDcQueue(orderNo ,token string)(r bool){
	resultOrderForDcQueueURL := "https://kyfw.12306.cn/otn/confirmPassenger/resultOrderForDcQueue"
	postDict := map[string]string{}
	/*
	postDict["orderSequence_no"] = orderNo
	postDict["_json_att"] = ""
	postDict["REPEAT_SUBMIT_TOKEN"] = token
    */

	postDict["orderSequence_no"] = orderNo+"&"+"_json_att=&"+"REPEAT_SUBMIT_TOKEN="+token
	resp:=getUrlRespHtml(resultOrderForDcQueueURL, postDict)
	m, _:=jsonstr2map(resp)
	fmt.Println("qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq", resp)
	if m["status"] == true{
		fmt.Println("是否提交成功: ", m["data"].(map[string]interface{})["submitStatus"])
		if m["data"].(map[string]interface{})["submitStatus"] == true{
			return true
		}else {
			return false
		}
	}
	return false
}

//查询排队结果
func queryOrderWaitTime(submitToken string)(orderId string){
	queryOrderWaitTimeURL := "https://kyfw.12306.cn/otn/confirmPassenger/queryOrderWaitTime"

	randStr := "random=" + GenerateRangeNum(1528701410000, 1528701419999)
	tourType := "tourFlag=dc"
	jsonatt := "_json_att="
	token := "REPEAT_SUBMIT_TOKEN=" + submitToken

	queryOrderWaitTimeURL += "?" + randStr + "&" + tourType + "&" + jsonatt + "&" + token

	resp:= getUrlRespHtml(queryOrderWaitTimeURL, nil)
	fmt.Println("queryOrderWaitTime =====================", resp)
	m, _:= jsonstr2map(resp)

	if m["status"] == true{
		if m["data"].(map[string]interface{})["orderId"] != nil{
			orderId = m["data"].(map[string]interface{})["orderId"].(string)
		}else {
			orderId = ""
		}
	}else{
		orderId = ""
	}
	return

}

//生成随机数字字符串
func GenerateRangeNum(min, max int) string {
	rand.Seed(rand.Int63())
	randNum := rand.Intn(max - min)
	randNum = randNum + min
	fmt.Println()
	return strconv.Itoa(randNum)
}

//确认排队队列
func confirmSingleForQueue(passengerTicketInfo, initDCInfo map[string]string, railwayInfo []string)(result bool){
	confirmSingleForQueueURL := "https://kyfw.12306.cn/otn/confirmPassenger/confirmSingleForQueue"
	postDict := map[string]string{}
	postDict["passengerTicketStr"] = passengerTicketInfo["passengerTicketStr"]
	postDict["oldPassengerStr"] = passengerTicketInfo["oldPassengerStr"]
	postDict["randCode"] = ""
	postDict["purpose_codes"] = "00"
	postDict["key_check_isChange"] = initDCInfo["key_check_isChange"]
	postDict["leftTicketStr"] = urlEncode(initDCInfo["leftTicket"]) //urlencode leftTicket
	postDict["train_location"] = railwayInfo[15]
	postDict["choose_seats"] = ""
	postDict["seatDetailType"] = "000"
	postDict["whatsSelect"] = "1"
	postDict["roomType"] = "00"
	postDict["dwAll"] = "N"
	postDict["_json_att"] = ""
	postDict["REPEAT_SUBMIT_TOKEN"] = initDCInfo["globalRepeatSubmitToken"]

	resp:=getUrlRespHtml(confirmSingleForQueueURL, postDict)
	m, _:= jsonstr2map(resp)
	if m["status"]== true{
		result = true
	}else{
		result = false
	}
	return result

}

//获取排队数量
func getQueueCount(railwayInfo []string, initDCInfo map[string]string)(result bool){

	getQueueCountURL := "https://kyfw.12306.cn/otn/confirmPassenger/getQueueCount"

	postDict := map[string]string{}
	postDict["train_date"] = timeFormat(railwayInfo[13])
	postDict["train_no"] = railwayInfo[2]
	postDict["stationTrainCode"] = railwayInfo[3]
	postDict["seatType"] = "O"  //座位类型O!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	postDict["fromStationTelecode"] = railwayInfo[6]
	postDict["toStationTelecode"] = railwayInfo[7]
	postDict["leftTicket"] = urlEncode(initDCInfo["leftTicket"]) //urlencode leftTicket
	postDict["purpose_codes"] = "ADULT"//"00"
	postDict["train_location"] = railwayInfo[15]
	postDict["_json_att"] = ""
	postDict["REPEAT_SUBMIT_TOKEN"] = initDCInfo["globalRepeatSubmitToken"]

	resp := getUrlRespHtml(getQueueCountURL, postDict)
	fmt.Println("getQueueCount===========================", resp)
	m, _:= jsonstr2map(resp)
	if m["status"] == true{
		result = true
	}else{
		result = false
	}
	return result

}

//设置时间格式
func timeFormat(date1 string)(formatTime string){
	time1, _:=time.Parse("20060102",date1)
	day := ""
	if time1.Day() < 10{
		day = "0" + strconv.Itoa(time1.Day())
	}else{
		day = strconv.Itoa(time1.Day())
	}
	t1:=time1.Weekday().String()[0:3]+ " " + time1.Month().String()[0:3] + " " + day + " " + strconv.Itoa(time1.Year())+ " 00:00:00 GMT+0800 "
	t2:= "中国标准时间"
	formatTime = urlEncode(t1) + "("+ urlEncode(t2) + ")"
	//formatTime = time1.Weekday().String()[0:3]+ " " + time1.Month().String()[0:3] + " " + day + " " + strconv.Itoa(time1.Year()) + " 00:00:00 GMT+0800 (中国标准时间)"
	return formatTime
}

//检查订单信息
func checkOrderInfo(passenger map[string]interface{}, toke string)(passengerTickInfo map[string]string , result bool){
	checkOrderInfoURL := "https://kyfw.12306.cn/otn/confirmPassenger/checkOrderInfo"
	postDict := map[string]string{}
	postDict["cancel_flag"] = "2"
	postDict["bed_level_order_num"] = "000000000000000000000000000000"

	//格式：O,0,乘客类型,乘客名,证件类型,证件号,手机号码,保存常用联系人(Y或N)
	passergerticketStr:= seatTypeStr + "," + passenger["passenger_flag"].(string) +","+passenger["passenger_type"].(string)+","+ passenger["passenger_name"].(string)+","+passenger["passenger_id_type_code"].(string)+","+passenger["passenger_id_no"].(string)+","+passenger["mobile_no"].(string)+",N"
	fmt.Println(passergerticketStr)

	postDict["passengerTicketStr"] = urlEncode(passergerticketStr)
	passengerTickInfo = make(map[string]string)
	passengerTickInfo["passengerTicketStr"] = postDict["passengerTicketStr"]

	oldPassengerStr:= passenger["passenger_name"].(string)+","+passenger["passenger_id_type_code"].(string)+","+passenger["passenger_id_no"].(string)+","+passenger["passenger_type"].(string)+"_"
	fmt.Println(oldPassengerStr)
	postDict["oldPassengerStr"] = urlEncode(oldPassengerStr)
	passengerTickInfo["oldPassengerStr"] = postDict["oldPassengerStr"]

	postDict["tour_flag"] = "dc"
	postDict["randCode"] = ""
	postDict["whatsSelect"] = "1"
	postDict["_json_att"] = ""
	postDict["REPEAT_SUBMIT_TOKEN"] = toke

	resp:=getUrlRespHtml(checkOrderInfoURL, postDict)
	fmt.Println("chekOrderInfo===========================", resp)
	m, _:= jsonstr2map(resp)
	if m["status"] == true{
		result = true
	}else{
		result = false
	}
	return
}

//选择乘客
func choosePassenger(passengersInfo string)(passenger map[string]interface{}){
	var passengerNO int
	passengers := printPassengersInfo(passengersInfo)

	for{
		fmt.Println("选择乘客")
		stdin := bufio.NewReader(os.Stdin)
		fmt.Fscan(stdin, &passengerNO)
		if passengerNO>len(passengers) || passengerNO < 0{
			fmt.Println("error no")
			continue
		}
		break
	}
	passenger = make(map[string] interface{})
	passenger = passengers[passengerNO -1].(map[string]interface{})
	return passenger
}
//打印联系人信息
func printPassengersInfo(passengersInfo string)(passengers []interface{}){
	m,_:= jsonstr2map(passengersInfo)
	mm:= m["data"].(map[string]interface{})["normal_passengers"]

	fmt.Println(len(mm.([]interface{})))
	passengers = mm.([]interface{})
	fmt.Println("序号","姓名", "类型", "性别", "身份证号码")
	for k,v:=range mm.([]interface{}){
		fmt.Println(k+1, v.(map[string]interface{})["passenger_name"], v.(map[string]interface{})["passenger_type_name"], v.(map[string]interface{})["sex_name"], v.(map[string]interface{})["passenger_id_no"])
	}
	return
}

//获取乘客信息
func getPassengerDTOs(token string)(passengersInfo string){
	getPassengerURL := "https://kyfw.12306.cn/otn/confirmPassenger/getPassengerDTOs"
	postDict := map[string]string{}
	postDict["_json_att"] = ""
	postDict["REPEAT_SUBMIT_TOKEN"] = token
	passengersInfo = getUrlRespHtml(getPassengerURL, postDict)
	return passengersInfo
}

//初始化订单返回信息
func initDc()(initDcStr map[string]string){
	initDcURL := "https://kyfw.12306.cn/otn/confirmPassenger/initDc"
	postDict := map[string]string{}
	postDict["_json_att"] = ""
	resp := getUrlRespHtml(initDcURL, postDict)
	initDcStr = make(map[string]string)
	initDcStr["globalRepeatSubmitToken"] = parseUseReg(resp,"globalRepeatSubmitToken = '(.*?)'")
	initDcStr["key_check_isChange"] = parseUseReg(resp,"'key_check_isChange':'(.*?)'")
	initDcStr["leftTicket"] = parseUseReg(resp,"'leftTicketStr':'(.*?)'")
	return initDcStr
}

//提交订单请求
func submitOrderRequest(railwayInfo []string)(result bool){
	orderRequestURL := "https://kyfw.12306.cn/otn/leftTicket/submitOrderRequest"
	postDict := map[string]string{}
	/*
	postDict["secretStr"] = railwayInfo[0]
	postDict["train_date"] = RailwayDate
	postDict["back_train_date"] = time.Now().Format("2006-01-02")
	postDict["tour_flag"] = "dc"
	postDict["purpose_codes"] = PersonType
	postDict["query_from_station_name"] = urlDecode(getStationName(railwayInfo[6]))
	postDict["query_to_station_name"] = getStationName(railwayInfo[7])
	//postDict["undefined"] = ""
	*/
	temp :=  railwayInfo[0]+"&train_date=" + RailwayDate + "&back_train_date="+time.Now().Format("2006-01-02")+"&tour_flag=dc&purpose_codes="+PersonType+"&query_from_station_name="+getStationName(railwayInfo[6])+"&query_to_station_name="+getStationName(railwayInfo[7])+"&undefined"
	fmt.Println(temp)
	postDict["secretStr"] = temp
	resp := getUrlRespHtml(orderRequestURL, postDict)
	m, _:= jsonstr2map(resp)
	if m["status"] == true{
		result = true
	}else{
		fmt.Println(m["messages"].([]interface{})[0].(string))
		result = false
	}
	return result
}

//判断用户是否登录
func isLoginUser()(result bool){

	isLoginUserURL := "https://kyfw.12306.cn/otn/login/checkUser"
	postDict := map[string]string{}
	postDict["_json_att"] = ""
	resp:= getUrlRespHtml(isLoginUserURL, postDict)
	m,_ := jsonstr2map(resp)
	r:= m["data"].(map[string]interface{})["flag"]

	if r == true{
		result = true
	}else {
		fmt.Println(m["messages"].([]interface{})[0].(string))
		result = false
	}
	return result
}

//选择车次
func chooseRailway(railwayResult [][]string)(railwayInfo []string){
	var railwayNo int
	for{
		fmt.Println("input the num of the rail if you want to take")
		stdin := bufio.NewReader(os.Stdin)
		fmt.Fscan(stdin, &railwayNo)
		if railwayNo>len(railwayResult) || railwayNo < 0{
			fmt.Println("error no")
			continue
		}
		if railwayResult[railwayNo - 1][11] == "N"{
			fmt.Println("this railway is not available")
			continue
		}
		railwayInfo = railwayResult[railwayNo -1]
		break
	}
	return railwayInfo
}


//urlencode
func urlEncode(urlstr string)(urlEncodeStr string){
	return url.QueryEscape(urlstr)
}

//urldecode
func urlDecode(urlstr string)(urlDecodeStr string){
	urlDecodeStr, err:= url.QueryUnescape(urlstr)
	if err!=nil{
		fmt.Println(err)
		return ""
	}
	return urlDecodeStr
}


//通过正则表达式返回字符串信息
func parseUseReg(text, reg string)(reString string){
	r, _ := regexp.Compile(reg)//"globalRepeatSubmitToken = '(.*?)'"
	return r.FindStringSubmatch(text)[1]
}

//根据查询链接查询列车信息
func railwayInfo(queryStationURL string)(railwayResult [][]string){
	resp := getUrlRespHtml(queryStationURL, nil)
	m ,err:= jsonstr2map(resp)
	if err!=nil{
		fmt.Println("railwayInfo:", err)
	}
	resultInfo :=m["data"].(map[string]interface{})["result"]
	for _, v:=range resultInfo.([]interface{}){
		temp:= strings.Split(v.(string),"|")
		railwayResult = append(railwayResult,temp)
	}
	return railwayResult
}

//格式化列车车次信息
func formatRailwayInfo(railwayInfo [][]string){
	fmt.Println("共", len(railwayInfo),"趟车次信息，车次信息如下")
	fmt.Println("序号", "车 次", " 出发站->到达站 ", "出发时间->到达时间", "  历时  ", "商务座", "一等座", "二等座", "软卧", "动卧", "硬卧", "软座", "硬座", "无座", "能否购买")
	for k, v:=range railwayInfo{
		fmt.Println(k+1, " |", v[3], "|", getStationName(v[6]),"->", getStationName(v[7]), "|", v[8],"->", v[9], "| ", v[10], " | ",  v[32], " | ", v[31], " | ", v[30], " | ", v[23], " | ",  v[33], " | ",  v[28], " | ", v[24], " | ",  v[29], " | ", v[26], " | ", v[11],  " | ")
	}
}


//根据城市名称获得城市代码
func getStationCode(name string)(stationCode string){
	resp := readJs()
	stations := strings.Split(resp[strings.Index(resp,"@") + 1:],"@")
	for _, v := range stations{
		info :=strings.Split(v,"|")
		if strings.Compare(info[1], name) == 0 {
			stationCode = info[2]
			break
		}else{
			stationCode = ""
		}
	}
	return stationCode
}

//根据城市代码获得城市名称
func getStationName(code string)(stationName string){
	resp := readJs()
	stations := strings.Split(resp[strings.Index(resp,"@") + 1:],"@")
	for _, v := range stations{
		info :=strings.Split(v,"|")
		if strings.Compare(info[2], code) == 0 {
			stationName = info[1]
			break
		}else{
			stationName = ""
		}
	}
	return stationName
}

//判断文件是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

//读文件内容
func readJs()(jsContent string){
	path:=`D:\station.js`
	if b, _:= PathExists(path);!b{
		//qingqiuwenjian
		urlStationName := "https://kyfw.12306.cn/otn/resources/js/framework/station_name.js?station_version=1.9055"
		getFile(urlStationName,nil,path)
	}
	if contents,err := ioutil.ReadFile(path);err == nil {
		//因为contents是[]byte类型，直接转换成string类型后会多一行空格,需要使用strings.Replace替换换行符
		jsContent = strings.Replace(string(contents),"\n","",1)
	}
	return jsContent
}

//构造查询URL
func getQueryURL()(queryURL string){
	var startstation string
	var endsation string
	var stationdate string
	var stationseat string
	urlBase := "https://kyfw.12306.cn/otn/leftTicket/query?"
	urldate := "leftTicketDTO.train_date="
	fmt.Println("input stationdate format:yyyy-mm-dd")
	fmt.Scanln(&stationdate)
	RailwayDate = stationdate
	urldate += stationdate
	fmt.Println("input startstation")
	fmt.Scanln(&startstation)
	urlstartstation := "leftTicketDTO.from_station="
	urlstartstation += getStationCode(startstation)
	fmt.Println("input endsation")
	fmt.Scanln(&endsation)
	urlendstation := "leftTicketDTO.to_station="
	urlendstation += getStationCode(endsation)
	fmt.Println("input seat 0:ADULT")
	fmt.Scanln( &stationseat)

	switch stationseat {
	case"0":
		stationseat = "ADULT"
		PersonType = "ADULT"
	case"1":
		stationseat = "0"
	}

	urlseat := "purpose_codes="
	urlseat += stationseat
	return urlBase+urldate+"&"+urlstartstation+"&"+urlendstation+"&"+urlseat
}

//选择坐席
func seatType()(result string){
	flag := true
	for flag{
		fmt.Println("选择席位")
		fmt.Println("=====================================================")
		fmt.Println("  O：二等座，M：一等座，9：商务座，3：硬卧，1：硬座  ")
		fmt.Println("=====================================================")
		seat := ""
		fmt.Scanln(&seat)
		if seat == "O" || seat == "M" || seat == "9" || seat == "3" || seat == "1"{
			flag = false
			result = seat
		}else {
			fmt.Println("输入错误，请输入：O或M或9或3或1")
		}
	}
	return
}

//logout注销用户
func loginOut(){
	loginOutURL := "https://kyfw.12306.cn/otn/login/loginOut"
	getUrlRespHtml(loginOutURL,nil)
}


