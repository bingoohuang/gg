fork from [China ID](https://github.com/mritd/chinaid)

> chinaid 是一个用于生成中国各种信息的测试库，比如姓名、身份证号、地址、邮箱、银行卡号等。

本项目生成的测试数据尽量付合真实数据以模拟用户真实行为:

- 姓名: 使用常用的姓氏外加常见的名字，尽量使数据 "正常"
- 身份证号: 采用标准身份证规则生成(校验码有效)
- 手机号: 常用的手机号头部外加随机数字
- 银行卡号: 银行卡号采用正确的卡 bin 生成(LUHN 算法有效)
- 邮箱: 随机的前缀外加常用的域名后缀
- 地址: 省/城市信息使用真实数据，具体地址随机生成

```go
fmt.Println("姓名:", chinaid.Name())
fmt.Println("性别:", chinaid.Sex())
fmt.Println("地址:", chinaid.Address())
fmt.Println("手机:", chinaid.Mobile())
fmt.Println("身份证:", chinaid.ChinaID())
fmt.Println("有效期:", chinaid.ValidPeriod())
fmt.Println("发证机关:", chinaid.IssueOrg())
fmt.Println("邮箱:", chinaid.Email())
fmt.Println("银行卡:", chinaid.BankNo())
fmt.Println("日期:", chinaid.RandDate())
```

```sh
姓名: 武锴脹
性别: 男
地址: 四川省攀枝花市嫯航路3755号婘螐小区3单元1216室
手机: 18507708621
身份证: 156315197605103397
有效期: 20020716-20220716
发证机关: 平凉市公安局某某分局
邮箱: wvcykkyh@kjsth.co
银行卡: 6230959897028597497
日期: 1977-06-16 23:41:28 +0800 CST
```
