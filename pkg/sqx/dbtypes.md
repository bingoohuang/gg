# 多数据库类型支持

1. SQL 改写
   1. 标识符引用：字段名、表名等标识符引用符号，自动适配不同数据库（比如MySQL是反引号，其它是双引号)
   2. 分页查询：不同的数据库分页查询语法不一样，统一以 MySQL 语法 `limit offset,count` 写SQL，sqx 会根据实际数据库类型进行转换
   3. 变量形式：不同的数据库变量占位符不一样，统一以 MySQL 语法的 `?` 写SQL, sqx 会根据实际数据库类型进行转换
   4. 变量补齐：插入语句变量支持自动补齐，例如 `insert into t_x(name, age, addr) values(?)`，后面的变量占位符的数量，会自动匹配插入字段的数量（多退少补）
   5. 顺序变量：支持以变量位置索引，来表示变量占位符所在位置，需要使用第n个实际变量来绑定。例如 `update t set name = :2 where id = :1`，后面就可以按照 id, name的实际顺序来传递需要绑定的变量值
   6. 命名变量：支持以变量名称，来表示变量占位符，然后传递map或者struct对象来完成实际变量的绑定。例如 `update t set name = :name where id = :id`，后面可以传递 map[string]string {"name":"悟空","id":"1000"}，或者Person{Name: "悟空", ID: "1000"}来传递需要绑定的变量值
   7. 名称推断：命名变量在 update 和 insert 语句中，可以自动进行推断
      1. `update t set name = :? where id = :?`，会被自动推断成 `update t set name = :name where id = :id`
      2. `insert into t_x(name, age, addr) values(:?)`，会被自动推断成 `insert into t_x(name, age, addr) values(:name, :age, :addr)`
   8. 总数 SQL 生成，在需要查询总数（分页总数）时，可以根据给定的SQL，自动生成查询总数的SQL，比如 `select a,b,c from t where a > 1 order by b`，生成的查询总数 SQL 就是 `select count(*) from t where a > 1`
   9. 当 SQL 中存在 `in (?)`，将会根据传输变量个数，自动扩展为 `in(?,?)` 等
2. Upsert 支持
   1. MySQL 支持 insert ... on duplicate key update ...的语法，在其他数据库中，不适用，直接需要改写成 insert 和 update 两个语句，当 sqx 执行 insert 失败时，会尝试 update 操作
3. 查询结果 ORM 映射，支持
   1. 映射到 单个结构体 （只取查询结构第1行）
   2. 映射到 结构体切片
   3. 映射到 单个 string/int 等 （只取查询结构第1行第1列）
   4. 映射到 string/int 的切片 （只取查询结构第1列）
