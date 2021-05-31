package chinaid

import (
	"fmt"
	"testing"
)

func TestRandom(t *testing.T) {
	fmt.Println("姓名:", Name())
	fmt.Println("性别:", Sex())
	fmt.Println("地址:", Address())
	fmt.Println("手机:", Mobile())
	fmt.Println("身份证:", ChinaID())
	fmt.Println("有效期:", ValidPeriod())
	fmt.Println("发证机关:", IssueOrg())
	fmt.Println("邮箱:", Email())
	fmt.Println("银行卡:", BankNo())
	fmt.Println("日期:", RandDate())
}
