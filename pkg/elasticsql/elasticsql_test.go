package elasticsql

import (
	"encoding/json"
	"reflect"
	"testing"
)

var selectCaseMap = map[string]string{
	"process_id= 1":                 `{"query" : {"bool" : {"must" : [{"match" : {"process_id" : {"query" : "1"}}}]}}}`,
	"(process_id= 1)":               `{"query" : {"bool" : {"must" : [{"match" : {"process_id" : {"query" : "1"}}}]}}}`,
	"((process_id= 1))":             `{"query" : {"bool" : {"must" : [{"match" : {"process_id" : {"query" : "1"}}}]}}}`,
	"(process_id = 1 and status=1)": `{"query" : {"bool" : {"must" : [{"match" : {"process_id" : {"query" : "1"}}},{"match" : {"status" : {"query" : "1"}}}]}}}`,
	"process_id > 1":                `{"query" : {"bool" : {"must" : [{"range" : {"process_id" : {"gt" : "1"}}}]}}}`,
	"process_id < 1":                `{"query" : {"bool" : {"must" : [{"range" : {"process_id" : {"lt" : "1"}}}]}}}`,
	"process_id <= 1":               `{"query" : {"bool" : {"must" : [{"range" : {"process_id" : {"to" : "1"}}}]}}}`,
	"process_id >= '1'":             `{"query" : {"bool" : {"must" : [{"range" : {"process_id" : {"from" : "1"}}}]}}}`,
	"process_id != 1":               `{"query" : {"bool" : {"must" : [{"bool" : {"must_not" : [{"match" : {"process_id" : {"query" : "1"}}}]}}]}}}`,
	"process_id = 0 and status= 1 and channel = 4":                        `{"query" : {"bool" : {"must" : [{"match" : {"process_id" : {"query" : "0"}}},{"match" : {"status" : {"query" : "1"}}},{"match" : {"channel" : {"query" : "4"}}}]}}}`,
	"create_time between '2015-01-01 00:00:00' and '2015-01-01 00:00:00'": `{"query" : {"bool" : {"must" : [{"range" : {"create_time" : {"from" : "2015-01-01 00:00:00", "to" : "2015-01-01 00:00:00"}}}]}}}`,
	"process_id > 1 and status = 1":                                       `{"query" : {"bool" : {"must" : [{"range" : {"process_id" : {"gt" : "1"}}},{"match" : {"status" : {"query" : "1"}}}]}}}`,
	"create_time between '2015-01-01T00:00:00+0800' and '2017-01-01T00:00:00+0800' and process_id = 0 and status >= 1 and content = '三个男人' and phone = '15810324322'": `{"query" : {"bool" : {"must" : [{"range" : {"create_time" : {"from" : "2015-01-01T00:00:00+0800", "to" : "2017-01-01T00:00:00+0800"}}},{"match" : {"process_id" : {"query" : "0"}}},{"range" : {"status" : {"from" : "1"}}},{"match" : {"content" : {"query" : "三个男人"}}},{"match" : {"phone" : {"query" : "15810324322"}}}]}}}`,
	"id > 1 or process_id = 0":                                            `{"query" : {"bool" : {"should" : [{"range" : {"id" : {"gt" : "1"}}},{"match" : {"process_id" : {"query" : "0"}}}]}}}`,
	"id > 1 and d = 1 or process_id = 0 and x = 2":                        `{"query" : {"bool" : {"should" : [{"bool" : {"must" : [{"range" : {"id" : {"gt" : "1"}}},{"match" : {"d" : {"query" : "1"}}}]}},{"bool" : {"must" : [{"match" : {"process_id" : {"query" : "0"}}},{"match" : {"x" : {"query" : "2"}}}]}}]}}}`,
	"id > 1 order by id asc, order_id desc":                               `{"query" : {"bool" : {"must" : [{"range" : {"id" : {"gt" : "1"}}}]}},"sort" : [{"id": "asc"},{"order_id": "desc"}]}`,
	"(id > 1 and d = 1)":                                                  `{"query" : {"bool" : {"must" : [{"range" : {"id" : {"gt" : "1"}}},{"match" : {"d" : {"query" : "1"}}}]}}}`,
	"(id > 1 and d = 1) or (c=1)":                                         `{"query" : {"bool" : {"should" : [{"bool" : {"must" : [{"range" : {"id" : {"gt" : "1"}}},{"match" : {"d" : {"query" : "1"}}}]}},{"match" : {"c" : {"query" : "1"}}}]}}}`,
	"id > 1 or (process_id = 0)":                                          `{"query" : {"bool" : {"should" : [{"range" : {"id" : {"gt" : "1"}}},{"match" : {"process_id" : {"query" : "0"}}}]}}}`,
	"id in (1,2,3,4)":                                                     `{"query" : {"bool" : {"must" : [{"terms" : {"id" : [1, 2, 3, 4]}}]}}}`,
	"id in ('232', '323') and content = 'aaaa'":                           `{"query" : {"bool" : {"must" : [{"terms" : {"id" : ["232", "323"]}},{"match" : {"content" : {"query" : "aaaa"}}}]}}}`,
	"create_time between '2015-01-01 00:00:00' and '2014-02-02 00:00:00'": `{"query" : {"bool" : {"must" : [{"range" : {"create_time" : {"from" : "2015-01-01 00:00:00", "to" : "2014-02-02 00:00:00"}}}]}}}`,
	"a like '%a%'":                                                        `{"query" : {"bool" : {"must" : [{"match" : {"a" : {"query" : "a"}}}]}}}`,
	"`by` = 1":                                                            `{"query" : {"bool" : {"must" : [{"match" : {"by" : {"query" : "1"}}}]}}}`,
	"id not like '%aaa%'":                                                 `{"query" : {"bool" : {"must" : [{"bool" : {"must_not" : {"match" : {"id" : {"query" : "aaa"}}}}}]}}}`,
	"id not in (1,2,3)":                                                   `{"query" : {"bool" : {"must" : [{"bool" : {"must_not" : {"terms" : {"id" : [1, 2, 3]}}}}]}}}`,
	"limit 10,10":                                                         `{"query" : {"bool" : {"must": [{"match_all" : {}}]}},"from" : 10,"size" : 10}`,
	"limit 10":                                                            `{"query" : {"bool" : {"must": [{"match_all" : {}}]}},"size" : 10}`,
	"id != missing":                                                       `{"query" : {"bool" : {"must" : [{"bool" : {"must_not" : [{"missing":{"field":"id"}}]}}]}}}`,
	"id = missing":                                                        `{"query" : {"bool" : {"must" : [{"missing":{"field":"id"}}]}}}`,
	"order by `order`.abc":                                                `{"query" : {"bool" : {"must": [{"match_all" : {}}]}},"sort" : [{"order.abc": "asc"}]}`,
	"multi_match(query='this is a test', fields=(title,title.origin))":                       `{"query" : {"multi_match" : {"query" : "this is a test", "fields" : ["title","title.origin"]}}}`,
	"a= 1 and multi_match(query='this is a test', fields=(title,title.origin))":              `{"query" : {"bool" : {"must" : [{"match" : {"a" : {"query" : "1"}}},{"multi_match" : {"query" : "this is a test", "fields" : ["title","title.origin"]}}]}}}`,
	"a= 1 and multi_match(query='this is a test', fields=(title,title.origin), type=phrase)": `{"query" : {"bool" : {"must" : [{"match" : {"a" : {"query" : "1"}}},{"multi_match" : {"query" : "this is a test", "type" : "phrase", "fields" : ["title","title.origin"]}}]}}}`,
}

func TestSupported(t *testing.T) {
	for k, v := range selectCaseMap {
		var dslMap map[string]interface{}
		err := json.Unmarshal([]byte(v), &dslMap)
		if err != nil {
			println(v)
			t.Error("test case json unmarshal err!")
		}

		// test convert
		dsl, err := Convert(k)
		var dslConvertedMap map[string]interface{}
		err = json.Unmarshal([]byte(dsl), &dslConvertedMap)
		if err != nil {
			t.Error("the generated dsl json unmarshal error!", k)
		}

		if !reflect.DeepEqual(dslMap, dslConvertedMap) {
			t.Error("the generated dsl is not equal to expected", k)
		}
	}
}

var unsupportedCaseList = []string{
	"insert into a values(1,2)",
	"update a set id = 1",
	"delete from a where id=1",
	"select * from ak where NOT(id=1)",
	"select * from ak where 1 = 1",
	"1=a",
	"id is null",
	" a= 1 and multi_match(zz=1, query='this is a test', fields=(title,title.origin), type=phrase)",
	"zz(k=2)",
}

func TestUnsupported(t *testing.T) {
	for _, v := range unsupportedCaseList {
		if _, err := Convert(v); err == nil {
			t.Error("can not be true, these cases are not supported!", v)
		}
	}
}

var badSQLList = []string{
	//"select aaac fr",
	"delete",
	"update x",
	"insert ",
}

func TestBadSQL(t *testing.T) {
	for _, v := range badSQLList {
		dsl, err := Convert(v)
		if err == nil {
			t.Error("can not be true, these cases are not supported!", v, dsl)
		}
	}
}
