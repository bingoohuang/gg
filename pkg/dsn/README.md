# dsn parser

Parse string in the following format to a structure:

1. user:pass@host:port/dbname
1. user/pass@host:port/dbname
1. user:pass@host/dbname
1. user:pass@host
1. user@host

Details:

1. user is required
1. host is required
1. pass can be empty
1. port can be empty
1. dbname can be empty
1. user and pass is separated by : or /
1. host and port is separated by :
1. dbname is at the end after / if it is not empty
