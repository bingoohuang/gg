SQL Parse
-----------
**Elasticsql** uses a [sqlparser](https://github.com/xwb1989/sqlparser) to parse raw sql input by user. After the parse procedure, sql statement will be converted to an golang struct, and the where clauses will become AST tree.
```go
type Select struct {
	Comments    Comments
	Distinct    string
	SelectExprs SelectExprs
	From        TableExprs
	Where       *Where
	GroupBy     GroupBy
	Having      *Where
	OrderBy     OrderBy
	Limit       *Limit
	Lock        string
}
```
What we need to take into care are SelectExprs, From, Where, GroupBy, OrderBy, Limit. The **Select** clause will include the aggregation functions which will be used to build our aggregation DSL. The **From** will tell us which type user wants to query. The **Where** will help us build the query DSL. The **OrderBy** will be the sort DSL, and **Limit** will be the From and Size DSL. Finally, **GroupBy** clause will be used to form the terms when building aggregations.

Select clause
-------------
Select will not be handled unless group by expression appeared in your sql. Refer to group by clause for details.

From clause
--------------
We extract table name(elasticsearch type name) from the from clause.

Where clause
--------------
Where clause will be parsed to an AST tree. For example, where a = 1 and b = 2 and c = 3 will be parsed to the tree as below :
```
BoolExpr 
|    |=======|
|            |
AndExpr      ComparisonExpr
|     |
|     |==================================|
|                                        |  
|                                        |
Left(BoolExpr(ComparisonExpr))       Right(BoolExpr(ComparisonExpr))
```
What we need to do is just to recursively travese the AST tree, and generated the Elasticsearch bool query like this:
```json
{
    "query": {
        "bool": {
            "must": [
                {
                    "bool": {
                        "must": [
                            {
                                "match": {
                                    "a": {
                                        "query": "1",
                                        "type": "phrase"
                                    }
                                }
                            },
                            {
                                "match": {
                                    "b": {
                                        "query": "2",
                                        "type": "phrase"
                                    }
                                }
                            }
                        ]
                    }
                },
                {
                    "match": {
                        "c": {
                            "query": "3",
                            "type": "phrase"
                        }
                    }
                }
            ]
        }
    }
}
```
We treat the **ComparisonExpr** as the leaf node, the recursion end and return.

But the DSL is nested too deep. We could say that if the parent and the child are both **and expression**, we can just merge this two together to make the DSL more flat, like this:
```json
{
    "query": {
        "bool": {
            "must": [
                {
                    "match": {
                        "a": {
                            "query": "1",
                            "type": "phrase"
                        }
                    }
                },
                {
                    "match": {
                        "b": {
                            "query": "2",
                            "type": "phrase"
                        }
                    }
                },
                {
                    "match": {
                        "c": {
                            "query": "3",
                            "type": "phrase"
                        }
                    }
                }
            ]
        }
    }
}
```
Leaf Nodes translations
--------------
Comparison expressions and Range expressions(between/and in sql) are the most common leaf node. Here are the translation table:

|expression| translation                                           |
|----------|:------------------------------------------------------|
|a = 1     |{"match" : {"a" : {"query" : "1", "type" : "phrase"}}} |
|a >= 1    |{"range" : {"a" : {"from" : "1"}}}                     |
|a <= 1    |{"range" : {"a" : {"to" : "1"}}}                         |
|a > 1    |{"range" : {"a" : {"gt" : "1"}}}|
|a < 1     |{"range" : {"a" : {"lt" : "1"}}}|
|a != 1   |{"bool" : {"must_not" : [{"match" : {"a" : {"query" : "1", "type" : "phrase"}}}]}}|
|a in (1,2,3)|{"terms" : {"a" : [1,2,3]}} // strings will be quoted in quotes|
|a like '%a%'|like expression currently handled the same with equal, maybe will change in the future|
|a between 1 and 10|{"range" : {"a" : {"from" : "1", "to" : "10"}}}|
|not like| currently not handled|
|null check| currently not handled|

Carefully here, because of the limit of terms(the count of terms query cannot be bigger than 1024) query in Elasticsearch, you can only have less than 1024 in items here.

OrderBy clause
------------
Order By clause is simply translated to a sort array, such as:
```json
"sort": [
    {
        "id": "asc"
    },
    {
        "process_id": "desc"
    }
]
```
Limit clause
-----------
Limit is converted to the from and size.

When the query is an aggregation query, from and size will both be set to 0.

GroupBy clause
--------------
If the query contains Group By clause, then the query will be treated as an aggregation query.

First we get the aggregation fields from the fields after group by, and build the outer bucket terms. The aggregation queries will be nested in the same order as the appearance order of these fields.

Second we will extract the aggregation functions from the select clause to build the inner aggregation result set.

Please notice that, we will only extract the aggregations functions from your select statement. If \* or field name appear in the sql select statement, we will ignore that.

Here is an example : 
```sql
select count(*), sum(point), avg(height) from worksheet
group by channel, area
```
The generated dsl is:
```json
{
    "query": {
        "bool": {
            "must": [
                {
                    "match_all": {}
                }
            ]
        }
    },
    "from": 0,
    "size": 0,
    "aggregations": {
        "channel": {
            "aggregations": {
                "area": {
                    "aggregations": {
                        "AVG(height)": {
                            "avg": {
                                "field": "height"
                            }
                        },
                        "COUNT(*)": {
                            "value_count": {
                                "field": "_index"
                            }
                        },
                        "SUM(point)": {
                            "sum": {
                                "field": "point"
                            }
                        }
                    },
                    "terms": {
                        "field": "area",
                        "size": 0
                    }
                }
            },
            "terms": {
                "field": "channel",
                "size": 200
            }
        }
    }
}
```
The type is `worksheet`.