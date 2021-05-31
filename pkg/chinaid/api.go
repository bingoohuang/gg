package chinaid

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"time"
)

// ProvinceAndCity 返回随机省/城市
func ProvinceAndCity() string {
	return ProvinceCity[RandInt(0, len(ProvinceCity))]
}

// Address 返回随机地址
func Address() string {
	return ProvinceAndCity() +
		RandChinese(2, 3) + "路" +
		strconv.Itoa(RandInt(1, 8000)) + "号" +
		RandChinese(2, 3) + "小区" +
		strconv.Itoa(RandInt(1, 20)) + "单元" +
		strconv.Itoa(RandInt(101, 2500)) + "室"
}

// BankNo 返回随机银行卡号，银行卡号符合LUHN 算法并且有正确的卡 bin 前缀
func BankNo() string {
	// 随机选中银行卡卡头
	bank := CardBins[RandInt(0, len(CardBins))]
	// 获取 卡前缀(cardBin)
	prefixes := bank.Prefixes
	// 获取当前银行卡正确长度
	cardNoLength := bank.Length
	// 生成 长度-1 位卡号
	preCardNo := strconv.Itoa(prefixes[RandInt(0, len(prefixes))]) +
		fmt.Sprintf("%0*d", cardNoLength-7, RandInt64(0, int64(math.Pow10(cardNoLength-7))))
	// LUHN 算法处理
	return LUHNProcess(preCardNo)
}

// LUHNProcess 通过 LUHN 合成卡号处理给定的银行卡号
func LUHNProcess(preCardNo string) string {
	checkSum := 0
	tmpCardNo := reverseString(preCardNo)
	for i, s := range tmpCardNo {
		// 数据层确保卡号正确
		tmp, _ := strconv.Atoi(string(s))
		// 由于卡号实际少了一位，所以反转后卡号第一位一定为偶数位
		// 同时 i 正好也是偶数，此时 i 将和卡号奇偶位同步
		if i%2 == 0 {
			// 偶数位 *2 是否为两位数(>9)
			if tmp*2 > 9 {
				// 如果为两位数则 -9
				checkSum += tmp*2 - 9
			} else {
				// 否则直接相加即可
				checkSum += tmp * 2
			}
		} else {
			// 奇数位直接相加
			checkSum += tmp
		}
	}
	if checkSum%10 != 0 {
		return preCardNo + strconv.Itoa(10-checkSum%10)
	} else {
		// 如果不巧生成的前 卡长度-1 位正好符合 LUHN 算法
		// 那么需要递归重新生成(需要符合 cardBind 中卡号长度)
		return BankNo()
	}
}

// Email 返回随机邮箱，邮箱目前只支持常见的域名后缀
func Email() string {
	return RandSmallLetters(8) + "@" + RandSmallLetters(5) + DomainSuffix[RandInt(0, len(DomainSuffix))]
}

// IssueOrg 返回身份证签发机关(eg: XXX公安局/XX区分局)
func IssueOrg() string {
	return CityName[RandInt(0, len(CityName))] + "公安局某某分局"
}

// ValidPeriod 返回身份证有效期限(eg: 20150906-20350906)，有效期限固定为 20 年
func ValidPeriod() string {
	begin := RandDate()
	end := begin.AddDate(20, 0, 0)
	return begin.Format("20060102") + "-" + end.Format("20060102")
}

// ChinaID 返回中国大陆地区身份证号.
func ChinaID() string {
	// AreaCode 随机一个+4位随机数字(不够左填充0)
	areaCode := AreaCode[RandInt(0, len(AreaCode))] +
		fmt.Sprintf("%0*d", 4, RandInt(1, 9999))
	birthday := RandDate().Format("20060102")
	randomCode := fmt.Sprintf("%0*d", 3, RandInt(0, 999))
	prefix := areaCode + birthday + randomCode
	return prefix + verifyCode(prefix)
}

// verifyCode 通过给定的身份证号生成最后一位的 verifyCode
func verifyCode(cardId string) string {
	tmp := 0
	for i, v := range Wi {
		t, _ := strconv.Atoi(string(cardId[i]))
		tmp += t * v
	}
	return ValCodeArr[tmp%11]
}

// RandDate 返回随机时间，时间区间从 1970 年 ~ 2020 年
func RandDate() time.Time {
	begin, _ := time.Parse("2006-01-02 15:04:05", "1970-01-01 00:00:00")
	end, _ := time.Parse("2006-01-02 15:04:05", "2020-01-01 00:00:00")
	return RandDateRange(begin, end)
}

// RandDateRange 返回随机时间，时间区间从 1970 年 ~ 2020 年
func RandDateRange(from, to time.Time) time.Time {
	return time.Unix(RandInt64(from.Unix(), to.Unix()), 0)
}

// Mobile 返回中国大陆地区手机号
func Mobile() string {
	return MobilePrefix[RandInt(0, len(MobilePrefix))] + fmt.Sprintf("%0*d", 8, RandInt(0, 100000000))
}

// Sex 返回性别
func Sex() string {
	if RandInt(0, 2) == 0 {
		return "男"
	}
	return "女"
}

// Name 返回中国姓名，姓名已经尽量返回常用姓氏和名字
func Name() string {
	return Surnames[RandInt(0, len(Surnames))] + RandChineseN(2)
}

// RandChineseN 指定长度随机中文字符(包含复杂字符)。
func RandChineseN(n int) string {
	var buf bytes.Buffer
	for i := 0; i < n; i++ {
		buf.WriteRune(rune(RandInt(19968, 40869)))
	}
	return buf.String()
}

// RandChinese 指定范围随机中文字符.
func RandChinese(minLen, maxLen int) string {
	return RandChineseN(RandInt(minLen, maxLen))
}

// RandSmallLetters 随机英文小写字母.
func RandSmallLetters(len int) string {
	data := make([]byte, len)
	for i := 0; i < len; i++ {
		data[i] = byte(rand.Intn(26) + 97)
	}
	return string(data)
}

// RandInt 指定范围随机 int
func RandInt(min, max int) int {
	return min + rand.Intn(max-min)
}

// RandInt64 指定范围随机 int64
func RandInt64(min, max int64) int64 {
	return min + rand.Int63n(max-min)
}

// 反转字符串
func reverseString(s string) string {
	runes := []rune(s)
	for from, to := 0, len(runes)-1; from < to; from, to = from+1, to-1 {
		runes[from], runes[to] = runes[to], runes[from]
	}
	return string(runes)
}
